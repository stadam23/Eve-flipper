package zkillboard

import (
	"encoding/json"
	"testing"
)

func TestNewClient_NonNil(t *testing.T) {
	c := NewClient()
	if c == nil {
		t.Fatal("NewClient() returned nil")
	}
}

func TestRegionStats_UnmarshalJSON(t *testing.T) {
	raw := `{"id":10000002,"type":"region","shipsDestroyed":5000,"iskDestroyed":1.5e12}`
	var s RegionStats
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if s.ID != 10000002 || s.Type != "region" {
		t.Errorf("ID/Type = %v/%q", s.ID, s.Type)
	}
	if s.ShipsDestroyed != 5000 || s.ISKDestroyed != 1.5e12 {
		t.Errorf("ShipsDestroyed/ISKDestroyed = %v/%v", s.ShipsDestroyed, s.ISKDestroyed)
	}
}

func TestZKBInfo_UnmarshalJSON(t *testing.T) {
	raw := `{"locationID":60003760,"hash":"abc","fittedValue":1e9,"destroyedValue":5e8,"points":100,"npc":false}`
	var z ZKBInfo
	if err := json.Unmarshal([]byte(raw), &z); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if z.LocationID != 60003760 || z.FittedValue != 1e9 || z.DestroyedValue != 5e8 {
		t.Errorf("ZKBInfo = %+v", z)
	}
	if z.Points != 100 || z.NPC != false {
		t.Errorf("Points/NPC = %v/%v", z.Points, z.NPC)
	}
}

func TestKillmail_UnmarshalJSON(t *testing.T) {
	raw := `{"killmail_id":90000001,"zkb.hash":"h123","zkb":{"locationID":60003760,"totalValue":2e9}}`
	var k Killmail
	if err := json.Unmarshal([]byte(raw), &k); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if k.KillmailID != 90000001 || k.KillmailHash != "h123" {
		t.Errorf("KillmailID/Hash = %v/%q", k.KillmailID, k.KillmailHash)
	}
	if k.ZKB == nil || k.ZKB.LocationID != 60003760 || k.ZKB.TotalValue != 2e9 {
		t.Errorf("ZKB = %+v", k.ZKB)
	}
}
