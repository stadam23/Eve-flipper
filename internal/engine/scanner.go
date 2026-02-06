package engine

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"sync"

	"eve-flipper/internal/esi"
	"eve-flipper/internal/sde"
)

const (
	// DefaultMaxResults is the default number of results returned when not specified.
	DefaultMaxResults = 100
	// UnreachableJumps is the fallback jump count when no path exists.
	UnreachableJumps = 999
)

// EffectiveMaxResults returns the max results limit, using DefaultMaxResults if v <= 0.
func EffectiveMaxResults(v int, defaultVal int) int {
	if v <= 0 {
		return defaultVal
	}
	return v
}

// HistoryProvider is an interface for fetching and caching market history.
type HistoryProvider interface {
	GetMarketHistory(regionID int32, typeID int32) ([]esi.HistoryEntry, bool)
	SetMarketHistory(regionID int32, typeID int32, entries []esi.HistoryEntry)
}

// Scanner orchestrates market scans using SDE data and the ESI client.
type Scanner struct {
	SDE            *sde.Data
	ESI            *esi.Client
	History        HistoryProvider
	ContractsCache      *esi.ContractsCache      // Cache for contracts (5 min TTL)
	ContractItemsCache  *esi.ContractItemsCache  // Cache for contract items (immutable)
}

// NewScanner creates a Scanner with the given static data and ESI client.
func NewScanner(data *sde.Data, client *esi.Client) *Scanner {
	return &Scanner{
		SDE:            data,
		ESI:            client,
		ContractsCache:     esi.NewContractsCache(),
		ContractItemsCache: esi.NewContractItemsCache(),
	}
}

// Scan finds profitable flip opportunities based on the given parameters.
func (s *Scanner) Scan(params ScanParams, progress func(string)) ([]FlipResult, error) {
	progress("Finding systems within radius...")
	var buySystems, sellSystems map[int32]int
	var wg sync.WaitGroup
	wg.Add(2)
	minSec := params.MinRouteSecurity
	go func() {
		defer wg.Done()
		if minSec > 0 {
			buySystems = s.SDE.Universe.SystemsWithinRadiusMinSecurity(params.CurrentSystemID, params.BuyRadius, minSec)
		} else {
			buySystems = s.SDE.Universe.SystemsWithinRadius(params.CurrentSystemID, params.BuyRadius)
		}
	}()
	go func() {
		defer wg.Done()
		if minSec > 0 {
			sellSystems = s.SDE.Universe.SystemsWithinRadiusMinSecurity(params.CurrentSystemID, params.SellRadius, minSec)
		} else {
			sellSystems = s.SDE.Universe.SystemsWithinRadius(params.CurrentSystemID, params.SellRadius)
		}
	}()
	wg.Wait()

	buyRegions := s.SDE.Universe.RegionsInSet(buySystems)
	sellRegions := s.SDE.Universe.RegionsInSet(sellSystems)

	log.Printf("[DEBUG] Scan: buySystems=%d, sellSystems=%d, buyRegions=%d, sellRegions=%d",
		len(buySystems), len(sellSystems), len(buyRegions), len(sellRegions))

	progress(fmt.Sprintf("Fetching orders from %d+%d regions...", len(buyRegions), len(sellRegions)))
	idx := s.fetchAndIndex(buyRegions, buySystems, sellRegions, sellSystems)
	return s.calculateResults(params, idx, buySystems, progress)
}

// ScanMultiRegion finds profitable flip opportunities across whole regions.
func (s *Scanner) ScanMultiRegion(params ScanParams, progress func(string)) ([]FlipResult, error) {
	minSec := params.MinRouteSecurity

	// If target region is specified, use it directly instead of radius search
	if params.TargetRegionID > 0 {
		progress(fmt.Sprintf("Searching target region %d...", params.TargetRegionID))

		var buySystemsRadius map[int32]int
		if minSec > 0 {
			buySystemsRadius = s.SDE.Universe.SystemsWithinRadiusMinSecurity(params.CurrentSystemID, params.BuyRadius, minSec)
		} else {
			buySystemsRadius = s.SDE.Universe.SystemsWithinRadius(params.CurrentSystemID, params.BuyRadius)
		}
		buyRegions := s.SDE.Universe.RegionsInSet(buySystemsRadius)
		buySystems := s.SDE.Universe.SystemsInRegions(buyRegions)

		sellRegions := map[int32]bool{params.TargetRegionID: true}
		sellSystems := s.SDE.Universe.SystemsInRegions(sellRegions)

		progress(fmt.Sprintf("Fetching orders: buy from %d regions, sell to target region...", len(buyRegions)))
		idx := s.fetchAndIndex(buyRegions, buySystems, sellRegions, sellSystems)
		return s.calculateResults(params, idx, buySystemsRadius, progress)
	}

	// Default behavior: search by radius
	progress("Finding regions by radius...")
	var buySystemsRadius, sellSystemsRadius map[int32]int
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if minSec > 0 {
			buySystemsRadius = s.SDE.Universe.SystemsWithinRadiusMinSecurity(params.CurrentSystemID, params.BuyRadius, minSec)
		} else {
			buySystemsRadius = s.SDE.Universe.SystemsWithinRadius(params.CurrentSystemID, params.BuyRadius)
		}
	}()
	go func() {
		defer wg.Done()
		if minSec > 0 {
			sellSystemsRadius = s.SDE.Universe.SystemsWithinRadiusMinSecurity(params.CurrentSystemID, params.SellRadius, minSec)
		} else {
			sellSystemsRadius = s.SDE.Universe.SystemsWithinRadius(params.CurrentSystemID, params.SellRadius)
		}
	}()
	wg.Wait()

	buyRegions := s.SDE.Universe.RegionsInSet(buySystemsRadius)
	sellRegions := s.SDE.Universe.RegionsInSet(sellSystemsRadius)
	buySystems := s.SDE.Universe.SystemsInRegions(buyRegions)
	sellSystems := s.SDE.Universe.SystemsInRegions(sellRegions)

	progress(fmt.Sprintf("Fetching orders from %d+%d regions...", len(buyRegions), len(sellRegions)))
	idx := s.fetchAndIndex(buyRegions, buySystems, sellRegions, sellSystems)
	return s.calculateResults(params, idx, buySystemsRadius, progress)
}

// --- Streaming order index types ---

type sellInfo struct {
	Price        float64
	VolumeRemain int32
	LocationID   int64
	SystemID     int32
	OrderCount   int
}

type buyInfo struct {
	Price        float64
	VolumeRemain int32
	LocationID   int64
	SystemID     int32
	OrderCount   int
}

type locKey struct {
	typeID     int32
	locationID int64
}

// scanIndex holds pre-built maps from the streaming fetch phase.
// Built concurrently while orders are still arriving from ESI.
type scanIndex struct {
	cheapestSell map[int32]sellInfo
	sellCounts   map[locKey]int
	highestBuy   map[int32]buyInfo
	buyCounts    map[locKey]int
	// Raw orders kept for execution plan (indexed by location+type).
	sellOrders []esi.MarketOrder
	buyOrders  []esi.MarketOrder
}

// hubRegionPriority maps known high-traffic region IDs to priority (lower = first).
// These regions have the most orders and should be fetched earliest for pipeline benefit.
var hubRegionPriority = map[int32]int{
	10000002: 0, // The Forge (Jita)
	10000043: 1, // Domain (Amarr)
	10000032: 2, // Sinq Laison (Dodixie)
	10000042: 3, // Metropolis (Hek)
	10000030: 4, // Heimatar (Rens)
}

// fetchOrdersStream starts fetching orders for all regions concurrently and
// streams batches of filtered orders through the returned channel.
// Hub regions are launched first so the pipeline starts building maps from
// the largest data sets sooner.
func (s *Scanner) fetchOrdersStream(
	regions map[int32]bool,
	orderType string,
	validSystems map[int32]int,
) <-chan []esi.MarketOrder {
	ch := make(chan []esi.MarketOrder, len(regions))

	// Sort regions: hubs first, then the rest.
	sorted := make([]int32, 0, len(regions))
	for rid := range regions {
		sorted = append(sorted, rid)
	}
	sort.Slice(sorted, func(i, j int) bool {
		pi, oki := hubRegionPriority[sorted[i]]
		pj, okj := hubRegionPriority[sorted[j]]
		if oki && okj {
			return pi < pj
		}
		if oki {
			return true
		}
		if okj {
			return false
		}
		return sorted[i] < sorted[j]
	})

	var wg sync.WaitGroup
	for _, regionID := range sorted {
		wg.Add(1)
		go func(rid int32) {
			defer wg.Done()
			orders, err := s.ESI.FetchRegionOrders(rid, orderType)
			if err != nil {
				return
			}
			// Filter to valid systems
			filtered := make([]esi.MarketOrder, 0, len(orders)/2)
			for _, o := range orders {
				if _, ok := validSystems[o.SystemID]; ok {
					filtered = append(filtered, o)
				}
			}
			if len(filtered) > 0 {
				ch <- filtered
			}
		}(regionID)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}

// fetchAndIndex launches parallel streaming fetches for sell and buy orders,
// building the scanIndex incrementally as regions complete.
func (s *Scanner) fetchAndIndex(
	buyRegions map[int32]bool, buySystems map[int32]int,
	sellRegions map[int32]bool, sellSystems map[int32]int,
) *scanIndex {
	sellCh := s.fetchOrdersStream(buyRegions, "sell", buySystems)
	buyCh := s.fetchOrdersStream(sellRegions, "buy", sellSystems)

	idx := &scanIndex{
		cheapestSell: make(map[int32]sellInfo),
		sellCounts:   make(map[locKey]int),
		highestBuy:   make(map[int32]buyInfo),
		buyCounts:    make(map[locKey]int),
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Consumer 1: build sell-side maps while orders stream in
	go func() {
		defer wg.Done()
		for batch := range sellCh {
			idx.sellOrders = append(idx.sellOrders, batch...)
			for _, o := range batch {
				idx.sellCounts[locKey{o.TypeID, o.LocationID}]++
				if cur, ok := idx.cheapestSell[o.TypeID]; !ok || o.Price < cur.Price {
					idx.cheapestSell[o.TypeID] = sellInfo{o.Price, o.VolumeRemain, o.LocationID, o.SystemID, 0}
				}
			}
		}
		// Fill order counts for cheapest sell locations
		for tid, info := range idx.cheapestSell {
			info.OrderCount = idx.sellCounts[locKey{tid, info.LocationID}]
			idx.cheapestSell[tid] = info
		}
	}()

	// Consumer 2: build buy-side maps while orders stream in
	go func() {
		defer wg.Done()
		for batch := range buyCh {
			idx.buyOrders = append(idx.buyOrders, batch...)
			for _, o := range batch {
				idx.buyCounts[locKey{o.TypeID, o.LocationID}]++
				if cur, ok := idx.highestBuy[o.TypeID]; !ok || o.Price > cur.Price {
					idx.highestBuy[o.TypeID] = buyInfo{o.Price, o.VolumeRemain, o.LocationID, o.SystemID, 0}
				}
			}
		}
		for tid, info := range idx.highestBuy {
			info.OrderCount = idx.buyCounts[locKey{tid, info.LocationID}]
			idx.highestBuy[tid] = info
		}
	}()

	wg.Wait()

	log.Printf("[DEBUG] fetchAndIndex: %d sell orders, %d buy orders", len(idx.sellOrders), len(idx.buyOrders))
	log.Printf("[DEBUG] cheapestSell: %d types, highestBuy: %d types", len(idx.cheapestSell), len(idx.highestBuy))
	return idx
}

// calculateResults is the shared profit calculation logic.
// bfsDistances = pre-computed distances from origin (used for buyJumps lookup).
func (s *Scanner) calculateResults(
	params ScanParams,
	idx *scanIndex,
	bfsDistances map[int32]int,
	progress func(string),
) ([]FlipResult, error) {
	cheapestSell := idx.cheapestSell
	highestBuy := idx.highestBuy
	sellOrders := idx.sellOrders
	buyOrders := idx.buyOrders

	progress("Calculating profits...")
	taxMult := 1.0 - params.SalesTaxPercent/100
	if taxMult < 0 {
		taxMult = 0
	}
	brokerMult := params.BrokerFeePercent / 100 // 0 for instant trades

	var results []FlipResult

	for typeID, sell := range cheapestSell {
		buy, ok := highestBuy[typeID]
		if !ok || buy.Price <= sell.Price {
			continue
		}

		// OPT: early margin check before item lookup
		// Broker fee applies to both sides when placing limit orders (0 for instant trades)
		effectiveBuyPrice := sell.Price * (1 + brokerMult)
		effectiveSellPrice := buy.Price * taxMult * (1 - brokerMult)
		profitPerUnit := effectiveSellPrice - effectiveBuyPrice
		if profitPerUnit <= 0 {
			continue
		}
		margin := profitPerUnit / effectiveBuyPrice * 100
		if margin < params.MinMargin {
			continue
		}

		itemType, ok := s.SDE.Types[typeID]
		if !ok || itemType.Volume <= 0 {
			continue
		}

		unitsF := math.Floor(params.CargoCapacity / itemType.Volume)
		if unitsF > math.MaxInt32 {
			unitsF = math.MaxInt32
		}
		units := int32(unitsF)
		if units <= 0 {
			continue
		}
		if sell.VolumeRemain < units {
			units = sell.VolumeRemain
		}
		if buy.VolumeRemain < units {
			units = buy.VolumeRemain
		}

		// MaxInvestment filter
		investment := sell.Price * float64(units)
		if params.MaxInvestment > 0 && investment > params.MaxInvestment {
			units = int32(params.MaxInvestment / sell.Price)
			if units <= 0 {
				continue
			}
		}

		totalProfit := profitPerUnit * float64(units)

		// OPT: use BFS distances when available, fallback to Dijkstra (with optional route security filter)
		minSec := params.MinRouteSecurity
		buyJumps := s.jumpsBetweenWithBFS(params.CurrentSystemID, sell.SystemID, bfsDistances, minSec)
		sellJumps := s.jumpsBetweenWithSecurity(sell.SystemID, buy.SystemID, minSec)
		totalJumps := buyJumps + sellJumps

		var profitPerJump float64
		if totalJumps > 0 {
			profitPerJump = totalProfit / float64(totalJumps)
		}

		buyRegionID := int32(0)
		if sys, ok := s.SDE.Systems[sell.SystemID]; ok {
			buyRegionID = sys.RegionID
		}
		sellRegionID := int32(0)
		if sys, ok := s.SDE.Systems[buy.SystemID]; ok {
			sellRegionID = sys.RegionID
		}
		results = append(results, FlipResult{
			TypeID:          typeID,
			TypeName:        itemType.Name,
			Volume:          itemType.Volume,
			BuyPrice:        sell.Price,
			BuyStation:      "",
			BuySystemName:   s.systemName(sell.SystemID),
			BuySystemID:     sell.SystemID,
			BuyRegionID:     buyRegionID,
			BuyRegionName:   s.regionName(buyRegionID),
			BuyLocationID:   sell.LocationID,
			SellPrice:       buy.Price,
			SellStation:     "",
			SellSystemName:  s.systemName(buy.SystemID),
			SellSystemID:    buy.SystemID,
			SellRegionID:    sellRegionID,
			SellRegionName:  s.regionName(sellRegionID),
			SellLocationID:  buy.LocationID,
			ProfitPerUnit:   profitPerUnit,
			MarginPercent:   margin,
			UnitsToBuy:      units,
			BuyOrderRemain:  buy.VolumeRemain,
			SellOrderRemain: sell.VolumeRemain,
			TotalProfit:     totalProfit,
			ProfitPerJump:   sanitizeFloat(profitPerJump),
			BuyJumps:        buyJumps,
			SellJumps:       sellJumps,
			TotalJumps:      totalJumps,
			BuyCompetitors:  sell.OrderCount, // sell orders at buy station (we compete to buy)
			SellCompetitors: buy.OrderCount,  // buy orders at sell station (we compete to sell)
		})
	}

	log.Printf("[DEBUG] found %d results before sort/trim", len(results))

	// Sort by profit, keep top 100
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalProfit > results[j].TotalProfit
	})
	limit := EffectiveMaxResults(params.MaxResults, DefaultMaxResults)
	if len(results) > limit {
		results = results[:limit]
	}

	// Enrich with execution-plan expected prices (same order book, no extra ESI).
	// Filter orders by location_id for more accurate slippage estimates.
	if len(results) > 0 {
		progress("Expected fill prices...")
		type locTypeKey struct {
			locationID int64
			typeID     int32
		}
		// Index sell orders by location+type (for buy-side execution plan at specific station)
		sellByLT := make(map[locTypeKey][]esi.MarketOrder)
		for _, o := range sellOrders {
			k := locTypeKey{o.LocationID, o.TypeID}
			sellByLT[k] = append(sellByLT[k], o)
		}
		// Index buy orders by location+type (for sell-side execution plan at specific station)
		buyByLT := make(map[locTypeKey][]esi.MarketOrder)
		for _, o := range buyOrders {
			k := locTypeKey{o.LocationID, o.TypeID}
			buyByLT[k] = append(buyByLT[k], o)
		}
		for i := range results {
			r := &results[i]
			planBuy := ComputeExecutionPlan(sellByLT[locTypeKey{r.BuyLocationID, r.TypeID}], r.UnitsToBuy, true)
			planSell := ComputeExecutionPlan(buyByLT[locTypeKey{r.SellLocationID, r.TypeID}], r.UnitsToBuy, false)
			r.ExpectedBuyPrice = planBuy.ExpectedPrice
			r.ExpectedSellPrice = planSell.ExpectedPrice
			r.SlippageBuyPct = planBuy.SlippagePercent
			r.SlippageSellPct = planSell.SlippagePercent
			if r.ExpectedBuyPrice > 0 && r.ExpectedSellPrice > 0 {
				effBuy := r.ExpectedBuyPrice * (1 + brokerMult)
				effSell := r.ExpectedSellPrice * taxMult * (1 - brokerMult)
				r.ExpectedProfit = (effSell - effBuy) * float64(r.UnitsToBuy)
			}
		}
	}

	// OPT: prefetch station names in parallel (only for top N)
	if len(results) > 0 {
		progress("Fetching station names...")
		topStations := make(map[int64]bool)
		for i := range results {
			topStations[results[i].BuyLocationID] = true
			topStations[results[i].SellLocationID] = true
		}
		s.ESI.PrefetchStationNames(topStations)

		// Fill station names from cache (instant, all prefetched)
		// For citadels (player structures), fallback to system name
		for i := range results {
			results[i].BuyStation = s.ESI.StationName(results[i].BuyLocationID)
			results[i].SellStation = s.ESI.StationName(results[i].SellLocationID)

			// If sell station is unresolved citadel, show system name instead
			if strings.HasPrefix(results[i].SellStation, "Location ") {
				if sys, ok := s.SDE.Systems[results[i].SellSystemID]; ok {
					results[i].SellStation = fmt.Sprintf("Structure @ %s", sys.Name)
				}
			}
			// Same for buy station
			if strings.HasPrefix(results[i].BuyStation, "Location ") {
				if sys, ok := s.SDE.Systems[results[i].BuySystemID]; ok {
					results[i].BuyStation = fmt.Sprintf("Structure @ %s", sys.Name)
				}
			}
		}
	}

	// Enrich with market history (volume, velocity, trend)
	s.enrichWithHistory(results, progress)

	// Compute DailyProfit = ProfitPerUnit * min(UnitsToBuy, DailyVolume)
	for i := range results {
		sellablePerDay := int64(results[i].UnitsToBuy)
		if results[i].DailyVolume > 0 && results[i].DailyVolume < sellablePerDay {
			sellablePerDay = results[i].DailyVolume
		}
		results[i].DailyProfit = results[i].ProfitPerUnit * float64(sellablePerDay)
	}

	// Post-filter: min daily volume
	if params.MinDailyVolume > 0 {
		filtered := results[:0]
		for _, r := range results {
			if r.DailyVolume >= params.MinDailyVolume {
				filtered = append(filtered, r)
			}
		}
		results = filtered
	}

	progress(fmt.Sprintf("Found %d profitable trades", len(results)))
	return results, nil
}

// fetchOrders is the legacy blocking version, kept for non-scan callers.
func (s *Scanner) fetchOrders(regions map[int32]bool, orderType string, validSystems map[int32]int) []esi.MarketOrder {
	ch := s.fetchOrdersStream(regions, orderType, validSystems)
	var all []esi.MarketOrder
	for batch := range ch {
		all = append(all, batch...)
	}
	log.Printf("[DEBUG] fetchOrders(%s): %d orders after filtering", orderType, len(all))
	return all
}

func (s *Scanner) jumpsBetween(from, to int32) int {
	return s.jumpsBetweenWithSecurity(from, to, 0)
}

// jumpsBetweenWithSecurity returns jump count using only systems with security >= minSecurity (0 = no filter).
func (s *Scanner) jumpsBetweenWithSecurity(from, to int32, minSecurity float64) int {
	var d int
	if minSecurity > 0 {
		d = s.SDE.Universe.ShortestPathMinSecurity(from, to, minSecurity)
	} else {
		d = s.SDE.Universe.ShortestPath(from, to)
	}
	if d < 0 {
		return UnreachableJumps
	}
	return d
}

// jumpsBetweenWithBFS uses pre-computed BFS distances when 'from' is the origin.
func (s *Scanner) jumpsBetweenWithBFS(from, to int32, bfsDistances map[int32]int, minRouteSecurity float64) int {
	if d, ok := bfsDistances[to]; ok {
		return d
	}
	return s.jumpsBetweenWithSecurity(from, to, minRouteSecurity)
}

// sanitizeFloat replaces NaN/Inf with 0 to prevent JSON marshal errors.
func sanitizeFloat(f float64) float64 {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0
	}
	return f
}

func (s *Scanner) systemName(systemID int32) string {
	if sys, ok := s.SDE.Systems[systemID]; ok {
		return sys.Name
	}
	return fmt.Sprintf("System %d", systemID)
}

func (s *Scanner) regionName(regionID int32) string {
	if r, ok := s.SDE.Regions[regionID]; ok {
		return r.Name
	}
	return fmt.Sprintf("Region %d", regionID)
}

// enrichWithHistory fetches market history for top results and fills DailyVolume/Velocity/PriceTrend.
// regionID is the sell region (where we care about volume).
func (s *Scanner) enrichWithHistory(results []FlipResult, progress func(string)) {
	if s.History == nil || len(results) == 0 {
		return
	}

	progress("Fetching market history...")

	// Determine region for each result (sell side)
	type historyKey struct {
		regionID int32
		typeID   int32
	}
	needed := make(map[historyKey]int) // key -> index in results
	for i := range results {
		regionID := s.SDE.Universe.SystemRegion[results[i].SellSystemID]
		if regionID == 0 {
			continue
		}
		key := historyKey{regionID, results[i].TypeID}
		needed[key] = i
	}

	// Fetch history concurrently (limited)
	type histResult struct {
		idx   int
		stats esi.MarketStats
	}
	ch := make(chan histResult, len(needed))
	sem := make(chan struct{}, 10) // limit concurrent history requests

	for key, idx := range needed {
		sem <- struct{}{}
		go func(k historyKey, i int) {
			defer func() { <-sem }()

			// Try cache first
			entries, ok := s.History.GetMarketHistory(k.regionID, k.typeID)
			if !ok {
				var err error
				entries, err = s.ESI.FetchMarketHistory(k.regionID, k.typeID)
				if err != nil {
					ch <- histResult{i, esi.MarketStats{}}
					return
				}
				s.History.SetMarketHistory(k.regionID, k.typeID, entries)
			}

			totalListed := results[i].SellOrderRemain + results[i].BuyOrderRemain
			stats := esi.ComputeMarketStats(entries, totalListed)
			ch <- histResult{i, stats}
		}(key, idx)
	}

	for range needed {
		r := <-ch
		results[r.idx].DailyVolume = r.stats.DailyVolume
		results[r.idx].Velocity = sanitizeFloat(r.stats.Velocity)
		results[r.idx].PriceTrend = sanitizeFloat(r.stats.PriceTrend)
	}
}
