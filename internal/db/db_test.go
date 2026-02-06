package db

import (
	"database/sql"
	"testing"

	"eve-flipper/internal/config"
	"eve-flipper/internal/engine"

	_ "modernc.org/sqlite"
)

// openTestDB opens an in-memory SQLite DB and runs migrations (for testing only).
func openTestDB(t *testing.T) *DB {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", ":memory:?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	d := &DB{sql: sqlDB}
	if err := d.migrate(); err != nil {
		sqlDB.Close()
		t.Fatalf("migrate: %v", err)
	}
	return d
}

func TestDB_MigrateAndHistoryRoundTrip(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	id := d.InsertHistory("radius", "Jita", 10, 1_500_000.5)
	if id <= 0 {
		t.Fatal("InsertHistory returned 0")
	}

	records := d.GetHistory(5)
	if len(records) != 1 {
		t.Fatalf("GetHistory(5) len = %d, want 1", len(records))
	}
	if records[0].ID != id {
		t.Errorf("GetHistory ID = %d, want %d", records[0].ID, id)
	}
	if records[0].Tab != "radius" || records[0].System != "Jita" {
		t.Errorf("Tab/System = %q/%q, want radius/Jita", records[0].Tab, records[0].System)
	}
	if records[0].Count != 10 {
		t.Errorf("Count = %d, want 10", records[0].Count)
	}
	if records[0].TopProfit != 1_500_000.5 {
		t.Errorf("TopProfit = %v, want 1500000.5", records[0].TopProfit)
	}
}

func TestDB_FlipResultsRoundTrip(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	id := d.InsertHistory("radius", "Jita", 1, 100)
	if id <= 0 {
		t.Fatal("InsertHistory failed")
	}

	results := []engine.FlipResult{
		{
			TypeID: 100, TypeName: "Test Item",
			BuyPrice: 90, SellPrice: 100,
			ProfitPerUnit: 10, MarginPercent: 11.11,
			UnitsToBuy: 50, TotalProfit: 500,
			BuyStation: "A", SellStation: "B",
			BuySystemName: "Sys1", SellSystemName: "Sys2",
			BuySystemID: 1, SellSystemID: 2,
			BuyOrderRemain: 100, SellOrderRemain: 200,
			ProfitPerJump: 100, BuyJumps: 1, SellJumps: 2, TotalJumps: 3,
		},
	}
	d.InsertFlipResults(id, results)

	got := d.GetFlipResults(id)
	if len(got) != 1 {
		t.Fatalf("GetFlipResults len = %d, want 1", len(got))
	}
	r := got[0]
	if r.TypeID != 100 || r.TypeName != "Test Item" {
		t.Errorf("TypeID/TypeName = %d/%q", r.TypeID, r.TypeName)
	}
	if r.BuyPrice != 90 || r.SellPrice != 100 {
		t.Errorf("Buy/Sell = %v/%v", r.BuyPrice, r.SellPrice)
	}
	if r.ProfitPerUnit != 10 || r.TotalProfit != 500 {
		t.Errorf("ProfitPerUnit/TotalProfit = %v/%v", r.ProfitPerUnit, r.TotalProfit)
	}
	if r.UnitsToBuy != 50 {
		t.Errorf("UnitsToBuy = %d", r.UnitsToBuy)
	}
}

func TestDB_GetHistoryByID(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	id := d.InsertHistory("contracts", "Amarr", 5, 2_000_000)
	rec := d.GetHistoryByID(id)
	if rec == nil {
		t.Fatal("GetHistoryByID returned nil")
	}
	if rec.ID != id || rec.System != "Amarr" || rec.Count != 5 {
		t.Errorf("record = %+v", rec)
	}

	if d.GetHistoryByID(99999) != nil {
		t.Error("GetHistoryByID(99999) should return nil")
	}
}

func TestDB_InsertFlipResults_ZeroScanIDNoOp(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	d.InsertFlipResults(0, []engine.FlipResult{{TypeID: 1}})
	got := d.GetFlipResults(0)
	if len(got) != 0 {
		t.Errorf("InsertFlipResults(0, ...) should not insert; GetFlipResults(0) len = %d", len(got))
	}
}

func TestDB_ConfigRoundTrip(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	cfg := &config.Config{
		SystemName:      "Amarr",
		CargoCapacity:   8000,
		BuyRadius:       7,
		SellRadius:      12,
		MinMargin:       10,
		SalesTaxPercent: 6,
		Opacity:         200,
		WindowW:         1024,
		WindowH:         768,
	}
	if err := d.SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	got := d.LoadConfig()
	if got.SystemName != cfg.SystemName || got.CargoCapacity != cfg.CargoCapacity {
		t.Errorf("LoadConfig = SystemName %q CargoCapacity %v", got.SystemName, got.CargoCapacity)
	}
	if got.BuyRadius != 7 || got.SellRadius != 12 || got.MinMargin != 10 || got.SalesTaxPercent != 6 {
		t.Errorf("LoadConfig radii/margin/tax = %d %d %v %v", got.BuyRadius, got.SellRadius, got.MinMargin, got.SalesTaxPercent)
	}
	if got.WindowW != 1024 || got.WindowH != 768 {
		t.Errorf("LoadConfig window = %dx%d", got.WindowW, got.WindowH)
	}
}

func TestDB_DemandRegionRoundTrip(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	r := &DemandRegion{
		RegionID:      10000033,
		RegionName:    "Tash-Murkon",
		HotScore:      1.5,
		Status:        "hot",
		KillsToday:    100,
		KillsBaseline: 50,
		ISKDestroyed:  2e11,
		ActivePlayers: 200,
		TopShips:      []string{"Ship A", "Ship B"},
	}
	if err := d.SaveDemandRegion(r); err != nil {
		t.Fatalf("SaveDemandRegion: %v", err)
	}
	list, err := d.GetDemandRegions()
	if err != nil {
		t.Fatalf("GetDemandRegions: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("GetDemandRegions len = %d, want 1", len(list))
	}
	if list[0].RegionID != r.RegionID || list[0].RegionName != r.RegionName || list[0].HotScore != r.HotScore {
		t.Errorf("GetDemandRegions[0] = %+v", list[0])
	}
	if len(list[0].TopShips) != 2 || list[0].TopShips[0] != "Ship A" {
		t.Errorf("TopShips = %v", list[0].TopShips)
	}

	one, err := d.GetDemandRegion(10000033)
	if err != nil || one == nil {
		t.Fatalf("GetDemandRegion(10000033): %v", err)
	}
	if one.RegionName != "Tash-Murkon" || one.ISKDestroyed != 2e11 {
		t.Errorf("GetDemandRegion = %+v", one)
	}
}

func TestDB_DemandRegion_NotFound(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()
	got, err := d.GetDemandRegion(99999)
	if err != nil {
		t.Errorf("GetDemandRegion(99999) err = %v (API returns nil,nil for no rows)", err)
	}
	if got != nil {
		t.Error("GetDemandRegion(99999) should return nil region")
	}
}
