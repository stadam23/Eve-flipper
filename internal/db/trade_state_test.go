package db

import "testing"

func TestTradeStateCRUD(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	state := TradeState{
		Tab:           "station",
		TypeID:        34,
		StationID:     60003760,
		RegionID:      10000002,
		Mode:          TradeStateModeIgnored,
		UntilRevision: 0,
	}
	if err := d.UpsertTradeStateForUser("user-a", state); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	items, err := d.ListTradeStatesForUser("user-a", "station")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Mode != TradeStateModeIgnored {
		t.Fatalf("mode mismatch: %q", items[0].Mode)
	}

	if err := d.DeleteTradeStatesForUser("user-a", "station", []TradeStateKey{
		{TypeID: 34, StationID: 60003760, RegionID: 10000002},
	}); err != nil {
		t.Fatalf("delete: %v", err)
	}

	items, err = d.ListTradeStatesForUser("user-a", "station")
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty list, got %d", len(items))
	}
}

func TestDeleteExpiredDoneTradeStatesForUser(t *testing.T) {
	d := openTestDB(t)
	defer d.Close()

	if err := d.UpsertTradeStateForUser("user-a", TradeState{
		Tab:           "station",
		TypeID:        34,
		StationID:     60003760,
		RegionID:      10000002,
		Mode:          TradeStateModeDone,
		UntilRevision: 100,
	}); err != nil {
		t.Fatalf("upsert done: %v", err)
	}
	if err := d.UpsertTradeStateForUser("user-a", TradeState{
		Tab:           "station",
		TypeID:        35,
		StationID:     60003760,
		RegionID:      10000002,
		Mode:          TradeStateModeIgnored,
		UntilRevision: 0,
	}); err != nil {
		t.Fatalf("upsert ignored: %v", err)
	}

	deleted, err := d.DeleteExpiredDoneTradeStatesForUser("user-a", "station", 120)
	if err != nil {
		t.Fatalf("delete expired: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted=%d, want 1", deleted)
	}

	items, err := d.ListTradeStatesForUser("user-a", "station")
	if err != nil {
		t.Fatalf("list after cleanup: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 remaining item, got %d", len(items))
	}
	if items[0].Mode != TradeStateModeIgnored {
		t.Fatalf("expected ignored to remain, got %q", items[0].Mode)
	}
}
