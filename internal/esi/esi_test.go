package esi

import (
	"encoding/json"
	"testing"
)

func TestMarketOrder_UnmarshalJSON(t *testing.T) {
	raw := `{"order_id":1,"type_id":34,"location_id":60003760,"system_id":30000142,"price":4.5,"volume_remain":100000,"is_buy_order":false}`
	var o MarketOrder
	if err := json.Unmarshal([]byte(raw), &o); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if o.OrderID != 1 || o.TypeID != 34 || o.LocationID != 60003760 || o.SystemID != 30000142 {
		t.Errorf("MarketOrder = %+v", o)
	}
	if o.Price != 4.5 || o.VolumeRemain != 100000 {
		t.Errorf("Price/VolumeRemain = %v/%v", o.Price, o.VolumeRemain)
	}
	if o.IsBuyOrder != false {
		t.Error("IsBuyOrder want false")
	}
}

func TestHistoryEntry_UnmarshalJSON(t *testing.T) {
	raw := `{"date":"2025-01-15","average":100.5,"highest":105,"lowest":98,"volume":50000,"order_count":12}`
	var h HistoryEntry
	if err := json.Unmarshal([]byte(raw), &h); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if h.Date != "2025-01-15" || h.Average != 100.5 || h.Highest != 105 || h.Lowest != 98 {
		t.Errorf("HistoryEntry = %+v", h)
	}
	if h.Volume != 50000 || h.OrderCount != 12 {
		t.Errorf("Volume/OrderCount = %v/%v", h.Volume, h.OrderCount)
	}
}

func TestNewClient_NonNil(t *testing.T) {
	c := NewClient(nil)
	if c == nil {
		t.Fatal("NewClient(nil) returned nil")
	}
}
