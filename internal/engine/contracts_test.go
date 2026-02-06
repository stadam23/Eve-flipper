package engine

import (
	"testing"
)

// getContractFilters returns effective minPrice, maxMargin, minPricedRatio (defaults when params are 0).

func TestGetContractFilters_Defaults(t *testing.T) {
	var params ScanParams
	minPrice, maxMargin, minPricedRatio := getContractFilters(params)
	if minPrice != DefaultMinContractPrice {
		t.Errorf("minPrice = %v, want DefaultMinContractPrice %v", minPrice, DefaultMinContractPrice)
	}
	if maxMargin != DefaultMaxContractMargin {
		t.Errorf("maxMargin = %v, want DefaultMaxContractMargin %v", maxMargin, DefaultMaxContractMargin)
	}
	if minPricedRatio != DefaultMinPricedRatio {
		t.Errorf("minPricedRatio = %v, want DefaultMinPricedRatio %v", minPricedRatio, DefaultMinPricedRatio)
	}
}

func TestGetContractFilters_Explicit(t *testing.T) {
	params := ScanParams{
		MinContractPrice:  50_000_000,
		MaxContractMargin: 80,
		MinPricedRatio:    0.9,
	}
	minPrice, maxMargin, minPricedRatio := getContractFilters(params)
	if minPrice != 50_000_000 {
		t.Errorf("minPrice = %v, want 50_000_000", minPrice)
	}
	if maxMargin != 80 {
		t.Errorf("maxMargin = %v, want 80", maxMargin)
	}
	if minPricedRatio != 0.9 {
		t.Errorf("minPricedRatio = %v, want 0.9", minPricedRatio)
	}
}

func TestGetContractFilters_PartialDefaults(t *testing.T) {
	// Only MinContractPrice set; others use defaults
	params := ScanParams{MinContractPrice: 1_000_000}
	minPrice, maxMargin, minPricedRatio := getContractFilters(params)
	if minPrice != 1_000_000 {
		t.Errorf("minPrice = %v, want 1_000_000", minPrice)
	}
	if maxMargin != DefaultMaxContractMargin {
		t.Errorf("maxMargin = %v, want default %v", maxMargin, DefaultMaxContractMargin)
	}
	if minPricedRatio != DefaultMinPricedRatio {
		t.Errorf("minPricedRatio = %v, want default %v", minPricedRatio, DefaultMinPricedRatio)
	}
}

func TestGetContractFilters_ZeroMaxMarginUsesDefault(t *testing.T) {
	params := ScanParams{MaxContractMargin: 0}
	_, maxMargin, _ := getContractFilters(params)
	if maxMargin != DefaultMaxContractMargin {
		t.Errorf("maxMargin when 0 = %v, want default %v", maxMargin, DefaultMaxContractMargin)
	}
}
