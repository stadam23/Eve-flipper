package api

import (
	"encoding/json"
	"testing"

	"eve-flipper/internal/db"
	"eve-flipper/internal/engine"
	"eve-flipper/internal/sde"
)

func newRegionalHistoryBackfillServer(database *db.DB) *Server {
	sdeData := &sde.Data{
		Systems: map[int32]*sde.SolarSystem{
			30000142: {ID: 30000142, Name: "Jita", RegionID: 10000002, Security: 0.9},
			30002187: {ID: 30002187, Name: "Amarr", RegionID: 10000043, Security: 0.8},
		},
		SystemByName: map[string]int32{
			"jita":  30000142,
			"amarr": 30002187,
		},
		Regions: map[int32]*sde.Region{
			10000002: {ID: 10000002, Name: "The Forge"},
			10000043: {ID: 10000043, Name: "Domain"},
		},
		RegionByName: map[string]int32{
			"the forge": 10000002,
			"domain":    10000043,
		},
	}
	scanner := engine.NewScanner(sdeData, nil)
	scanner.History = database
	return &Server{
		db:      database,
		scanner: scanner,
		sdeData: sdeData,
		ready:   true,
	}
}

func TestRebuildRegionalHistoryRows_FromLegacyRegionDayRecord(t *testing.T) {
	database := openAPITestDB(t)
	srv := newRegionalHistoryBackfillServer(database)

	paramsJSON, err := json.Marshal(scanRequest{
		SystemName:         "Jita",
		TargetRegion:       "Domain",
		TargetMarketSystem: "Jita",
		AvgPricePeriod:     14,
		MinMargin:          0,
		SalesTaxPercent:    8,
		BrokerFeePercent:   0,
	})
	if err != nil {
		t.Fatalf("json.Marshal params: %v", err)
	}

	record := &db.ScanRecord{
		ID:     1,
		Tab:    "region",
		System: "Jita",
		Params: paramsJSON,
	}

	raw := []engine.FlipResult{
		{
			TypeID:           34,
			TypeName:         "Tritanium",
			Volume:           0.01,
			BuyPrice:         100,
			SellPrice:        120,
			BuyStation:       "Jita 4-4",
			SellStation:      "Amarr VIII",
			BuySystemName:    "Jita",
			BuySystemID:      30000142,
			BuyRegionID:      10000002,
			BuyRegionName:    "The Forge",
			SellSystemName:   "Amarr",
			SellSystemID:     30002187,
			SellRegionID:     10000043,
			SellRegionName:   "Domain",
			UnitsToBuy:       10,
			SellOrderRemain:  500,
			S2BPerDay:        5,
			TargetSellSupply: 25,
			SellJumps:        4,
		},
	}

	rows := srv.rebuildRegionalHistoryRows(record, raw)
	if len(rows) != 1 {
		t.Fatalf("rows len = %d, want 1", len(rows))
	}
	row := rows[0]
	if row.TypeID != 34 {
		t.Fatalf("type_id = %d, want 34", row.TypeID)
	}
	if row.DayPeriodProfit <= 0 {
		t.Fatalf("day period profit = %.2f, want > 0", row.DayPeriodProfit)
	}
	if row.DayTargetDOS <= 0 {
		t.Fatalf("day target DOS = %.2f, want > 0", row.DayTargetDOS)
	}
	if row.DaySourceAvgPrice != 100 {
		t.Fatalf("day source avg price = %.2f, want 100", row.DaySourceAvgPrice)
	}
}

func TestRebuildRegionalHistoryRows_SkipsClassicRegionRecord(t *testing.T) {
	database := openAPITestDB(t)
	srv := newRegionalHistoryBackfillServer(database)

	paramsJSON, err := json.Marshal(scanRequest{
		SystemName: "Jita",
		BuyRadius:  5,
		SellRadius: 10,
	})
	if err != nil {
		t.Fatalf("json.Marshal params: %v", err)
	}

	record := &db.ScanRecord{
		ID:     2,
		Tab:    "region",
		System: "Jita",
		Params: paramsJSON,
	}
	raw := []engine.FlipResult{
		{TypeID: 34, TypeName: "Tritanium", BuyPrice: 100, SellPrice: 120, UnitsToBuy: 5},
	}

	rows := srv.rebuildRegionalHistoryRows(record, raw)
	if len(rows) != 0 {
		t.Fatalf("rows len = %d, want 0 for classic region scan", len(rows))
	}
}
