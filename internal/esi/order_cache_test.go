package esi

import (
	"testing"
	"time"
)

func TestOrderCacheWindowForRegionsAll(t *testing.T) {
	oc := NewOrderCache()
	now := time.Now().UTC()

	oc.Put(10000002, "sell", nil, "s1", now.Add(5*time.Minute))
	oc.Put(10000002, "buy", nil, "b1", now.Add(2*time.Minute))
	oc.Put(10000043, "sell", nil, "s2", now.Add(9*time.Minute))

	window := oc.WindowForRegions([]int32{10000002, 10000043, 10000002}, "all")
	if window.Regions != 3 {
		t.Fatalf("Regions=%d, want 3", window.Regions)
	}
	if window.Entries != 3 {
		t.Fatalf("Entries=%d, want 3", window.Entries)
	}
	if window.CurrentRevision <= 0 {
		t.Fatalf("CurrentRevision=%d, want > 0", window.CurrentRevision)
	}
	if window.MinTTLSeconds <= 0 {
		t.Fatalf("MinTTLSeconds=%d, want > 0", window.MinTTLSeconds)
	}
	if window.MaxTTLSeconds < window.MinTTLSeconds {
		t.Fatalf("MaxTTLSeconds=%d < MinTTLSeconds=%d", window.MaxTTLSeconds, window.MinTTLSeconds)
	}
	if window.NextExpiryAt.IsZero() {
		t.Fatal("NextExpiryAt is zero")
	}
	if window.LastRefreshAt.IsZero() {
		t.Fatal("LastRefreshAt is zero")
	}
}

func TestOrderCacheWindowForRegionsSellOnly(t *testing.T) {
	oc := NewOrderCache()
	now := time.Now().UTC()

	oc.Put(10000002, "sell", nil, "s1", now.Add(5*time.Minute))
	oc.Put(10000002, "buy", nil, "b1", now.Add(2*time.Minute))
	oc.Put(10000043, "sell", nil, "s2", now.Add(9*time.Minute))

	window := oc.WindowForRegions([]int32{10000002, 10000043}, "sell")
	if window.Entries != 2 {
		t.Fatalf("Entries=%d, want 2", window.Entries)
	}
	if window.MinTTLSeconds <= 0 {
		t.Fatalf("MinTTLSeconds=%d, want > 0", window.MinTTLSeconds)
	}
	if window.MaxTTLSeconds <= 0 {
		t.Fatalf("MaxTTLSeconds=%d, want > 0", window.MaxTTLSeconds)
	}
}

func TestOrderCacheClear(t *testing.T) {
	oc := NewOrderCache()
	now := time.Now().UTC()
	oc.Put(10000002, "sell", nil, "s1", now.Add(5*time.Minute))
	oc.Put(10000002, "buy", nil, "b1", now.Add(5*time.Minute))

	removed := oc.Clear()
	if removed != 2 {
		t.Fatalf("removed=%d, want 2", removed)
	}

	window := oc.WindowForRegions([]int32{10000002}, "all")
	if window.Entries != 0 {
		t.Fatalf("Entries=%d, want 0", window.Entries)
	}
}
