package zkillboard

import (
	"fmt"
	"sync"
	"time"

	"eve-flipper/internal/esi"
	"eve-flipper/internal/logger"
	"eve-flipper/internal/sde"
)

// ESIKillmail represents a full killmail from the ESI API.
type ESIKillmail struct {
	KillmailID    int64     `json:"killmail_id"`
	KillmailTime  string    `json:"killmail_time"`
	SolarSystemID int32     `json:"solar_system_id"`
	Victim        ESIVictim `json:"victim"`
}

// ESIVictim contains victim info including fitted items.
type ESIVictim struct {
	ShipTypeID int32     `json:"ship_type_id"`
	Items      []ESIItem `json:"items"`
}

// ESIItem is a single item from a killmail (fitted module, ammo, drone, cargo).
type ESIItem struct {
	TypeID            int32 `json:"type_id"`
	Flag              int32 `json:"flag"`
	QuantityDestroyed int32 `json:"quantity_destroyed"`
	QuantityDropped   int32 `json:"quantity_dropped"`
	Singleton         int32 `json:"singleton"`
}

// ItemDemandProfile aggregates destruction data for a single item type.
type ItemDemandProfile struct {
	TypeID         int32   `json:"type_id"`
	TypeName       string  `json:"type_name"`
	Category       string  `json:"category"` // "ship", "module", "ammo", "drone"
	TotalDestroyed int64   `json:"total_destroyed"`
	KillmailCount  int     `json:"killmail_count"`
	AvgPerKillmail float64 `json:"avg_per_killmail"`
	EstDailyDemand float64 `json:"est_daily_demand"`
}

// RegionDemandProfile contains aggregated fitting demand data for a region.
type RegionDemandProfile struct {
	RegionID      int32
	SampledKills  int
	TotalKills24h int
	Items         map[int32]*ItemDemandProfile
	UpdatedAt     time.Time
}

// categorizeItem classifies an item by its EVE inventory flag.
// See: https://docs.esi.evetech.net/docs/asset_location_flag
func categorizeItem(flag int32, volume float64) string {
	switch {
	// High slots (11-18), Medium slots (19-26), Low slots (27-34), Rig slots (92-99), Subsystem (125-132)
	case (flag >= 11 && flag <= 34) || (flag >= 92 && flag <= 99) || (flag >= 125 && flag <= 132):
		return "module"
	// Drone bay (87), Fighter bay (158-162)
	case flag == 87 || (flag >= 158 && flag <= 162):
		return "drone"
	// Cargo (5) — classify by volume: ammo/charges are very small
	case flag == 5:
		if volume > 0 && volume <= 0.1 {
			return "ammo"
		}
		return "" // skip generic cargo
	// Charges loaded in high/med/low slots share the module flag range
	// but QuantityDestroyed > 1 typically indicates ammo/charges
	default:
		return ""
	}
}

// AnalyzeRegionFittings fetches recent killmails and aggregates destroyed items.
func (d *DemandAnalyzer) AnalyzeRegionFittings(regionID int32, esiClient *esi.Client, sdeData *sde.Data, maxKillmails int) (*RegionDemandProfile, error) {
	if maxKillmails <= 0 {
		maxKillmails = 100
	}

	// Get recent killmails from zkillboard (last 24 hours)
	recentKills, err := d.client.GetRecentKills(regionID, 86400)
	if err != nil {
		return nil, fmt.Errorf("get recent kills: %w", err)
	}

	totalKills24h := len(recentKills)
	if totalKills24h == 0 {
		return &RegionDemandProfile{
			RegionID:      regionID,
			SampledKills:  0,
			TotalKills24h: 0,
			Items:         make(map[int32]*ItemDemandProfile),
			UpdatedAt:     time.Now(),
		}, nil
	}

	// Sample up to maxKillmails (most recent first — zkillboard returns newest first)
	sample := recentKills
	if len(sample) > maxKillmails {
		sample = sample[:maxKillmails]
	}

	// Extract killmail IDs and hashes
	type killRef struct {
		id   int64
		hash string
	}
	var refs []killRef
	for _, km := range sample {
		// zkillboard returns: {"killmail_id": 12345, "zkb": {"hash": "abc..."}}
		idFloat, ok := km["killmail_id"].(float64)
		if !ok {
			continue
		}
		zkbRaw, ok := km["zkb"].(map[string]interface{})
		if !ok {
			continue
		}
		hash, ok := zkbRaw["hash"].(string)
		if !ok {
			continue
		}
		refs = append(refs, killRef{id: int64(idFloat), hash: hash})
	}

	if len(refs) == 0 {
		return &RegionDemandProfile{
			RegionID:      regionID,
			SampledKills:  0,
			TotalKills24h: totalKills24h,
			Items:         make(map[int32]*ItemDemandProfile),
			UpdatedAt:     time.Now(),
		}, nil
	}

	logger.Info("Demand", fmt.Sprintf("Analyzing %d killmails for region %d (total 24h: %d)...", len(refs), regionID, totalKills24h))

	// Fetch full killmail details from ESI concurrently
	type kmResult struct {
		km  *ESIKillmail
		err error
	}
	results := make(chan kmResult, len(refs))

	// Process in batches to respect ESI rate limits
	batchSize := 20
	for i := 0; i < len(refs); i += batchSize {
		end := i + batchSize
		if end > len(refs) {
			end = len(refs)
		}

		var wg sync.WaitGroup
		for _, ref := range refs[i:end] {
			wg.Add(1)
			go func(r killRef) {
				defer wg.Done()
				var km ESIKillmail
				url := fmt.Sprintf("https://esi.evetech.net/latest/killmails/%d/%s/?datasource=tranquility", r.id, r.hash)
				if err := esiClient.GetJSON(url, &km); err != nil {
					results <- kmResult{err: err}
					return
				}
				results <- kmResult{km: &km}
			}(ref)
		}
		wg.Wait()
	}
	close(results)

	// Aggregate destroyed items
	items := make(map[int32]*ItemDemandProfile)
	sampledCount := 0

	for r := range results {
		if r.err != nil || r.km == nil {
			continue
		}
		sampledCount++

		// Count the ship hull as destroyed
		if r.km.Victim.ShipTypeID > 0 {
			shipID := r.km.Victim.ShipTypeID
			if _, ok := items[shipID]; !ok {
				name := ""
				if sdeData != nil {
					if t, ok := sdeData.Types[shipID]; ok {
						name = t.Name
					}
				}
				items[shipID] = &ItemDemandProfile{
					TypeID:   shipID,
					TypeName: name,
					Category: "ship",
				}
			}
			items[shipID].TotalDestroyed++
			items[shipID].KillmailCount++
		}

		// Process fitted items
		for _, item := range r.km.Victim.Items {
			destroyed := int64(item.QuantityDestroyed)
			if destroyed <= 0 {
				continue
			}

			// Get item volume from SDE for categorization
			var itemVolume float64
			var itemName string
			if sdeData != nil {
				if t, ok := sdeData.Types[item.TypeID]; ok {
					itemVolume = t.Volume
					itemName = t.Name
				}
			}

			category := categorizeItem(item.Flag, itemVolume)
			if category == "" {
				// Try to infer from quantity: large quantities in module slots = ammo/charges
				if item.QuantityDestroyed > 10 && item.Flag >= 11 && item.Flag <= 34 {
					category = "ammo"
				} else {
					continue
				}
			}

			if _, ok := items[item.TypeID]; !ok {
				items[item.TypeID] = &ItemDemandProfile{
					TypeID:   item.TypeID,
					TypeName: itemName,
					Category: category,
				}
			}
			items[item.TypeID].TotalDestroyed += destroyed
			items[item.TypeID].KillmailCount++
		}
	}

	// Calculate per-killmail averages and extrapolate daily demand
	scaleFactor := 1.0
	if sampledCount > 0 && totalKills24h > sampledCount {
		scaleFactor = float64(totalKills24h) / float64(sampledCount)
	}

	for _, profile := range items {
		if profile.KillmailCount > 0 {
			profile.AvgPerKillmail = float64(profile.TotalDestroyed) / float64(profile.KillmailCount)
		}
		profile.EstDailyDemand = float64(profile.TotalDestroyed) * scaleFactor

		// Ammo/charges are consumed during combat — multiply by 3x
		// (killed ships had ~1/3 of what was actually used in the fight)
		if profile.Category == "ammo" {
			profile.EstDailyDemand *= 3.0
		}
	}

	logger.Success("Demand", fmt.Sprintf("Region %d: analyzed %d killmails, found %d unique items", regionID, sampledCount, len(items)))

	return &RegionDemandProfile{
		RegionID:      regionID,
		SampledKills:  sampledCount,
		TotalKills24h: totalKills24h,
		Items:         items,
		UpdatedAt:     time.Now(),
	}, nil
}
