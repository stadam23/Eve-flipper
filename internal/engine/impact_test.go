package engine

import (
	"math"
	"testing"
	"time"

	"eve-flipper/internal/esi"
)

func TestVariance(t *testing.T) {
	// Population variance: [1,2,3,4,5] mean=3, var = (4+1+0+1+4)/5 = 2
	x := []float64{1, 2, 3, 4, 5}
	got := variance(x)
	if math.Abs(got-2.0) > 1e-9 {
		t.Errorf("variance(%v) = %v, want 2", x, got)
	}
	if variance(nil) != 0 {
		t.Errorf("variance(nil) want 0")
	}
	if variance([]float64{7}) != 0 {
		t.Errorf("variance(single) want 0")
	}
}

func TestCalibrateImpact(t *testing.T) {
	// Use dates within last 30 days so filterLastNDays keeps them
	base := time.Now().AddDate(0, 0, -10).Format("2006-01-02")
	d1 := time.Now().AddDate(0, 0, -9).Format("2006-01-02")
	d2 := time.Now().AddDate(0, 0, -8).Format("2006-01-02")
	d3 := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	d4 := time.Now().AddDate(0, 0, -6).Format("2006-01-02")
	history := []esi.HistoryEntry{
		{Date: base, Average: 100, Highest: 102, Lowest: 98, Volume: 1000, OrderCount: 10},
		{Date: d1, Average: 100, Highest: 103, Lowest: 97, Volume: 2000, OrderCount: 12},
		{Date: d2, Average: 101, Highest: 104, Lowest: 98, Volume: 1500, OrderCount: 11},
		{Date: d3, Average: 99, Highest: 102, Lowest: 96, Volume: 1200, OrderCount: 9},
		{Date: d4, Average: 100, Highest: 105, Lowest: 95, Volume: 1800, OrderCount: 14},
	}
	params := CalibrateImpact(history, 30)
	if !params.Valid {
		t.Fatal("expected Valid=true with 5 days of history")
	}
	if params.Lambda <= 0 || params.Eta <= 0 {
		t.Errorf("expected positive Lambda and Eta, got Lambda=%v Eta=%v", params.Lambda, params.Eta)
	}
	if params.DaysUsed != 5 {
		t.Errorf("DaysUsed want 5, got %d", params.DaysUsed)
	}
}

func TestImpactLinear(t *testing.T) {
	// ΔP = λ × Q
	lambda := 0.01
	q := 1000.0
	got := ImpactLinear(lambda, q)
	want := 10.0
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("ImpactLinear(%v, %v) = %v, want %v", lambda, q, got, want)
	}
}

func TestImpactSqrt(t *testing.T) {
	// ΔP = η × √Q
	eta := 1.0
	q := 10000.0
	got := ImpactSqrt(eta, q)
	want := 100.0
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("ImpactSqrt(%v, %v) = %v, want %v", eta, q, got, want)
	}
}

func TestOptimalSlicesTWAP(t *testing.T) {
	// n* = √(σ² T / (2 λ κ)), κ = 0.5λ  =>  n* = √(σ² T / λ²)
	sigmaSq := 0.0001
	lambda := 0.01
	T := 1.0
	kappaRatio := 0.5
	n := OptimalSlicesTWAP(sigmaSq, lambda, T, kappaRatio)
	if n < 1 || n > 100 {
		t.Errorf("OptimalSlicesTWAP returned %d, expected 1..100", n)
	}
}

func TestEstimateImpact(t *testing.T) {
	params := ImpactParams{Lambda: 0.01, Eta: 2.0, SigmaSq: 0.0002, Valid: true}
	est := EstimateImpact(params, 5000, 1.0)
	if est.LinearImpact <= 0 || est.SqrtImpact <= 0 {
		t.Errorf("expected positive impacts, got Linear=%v Sqrt=%v", est.LinearImpact, est.SqrtImpact)
	}
	if est.RecommendedImpact != est.SqrtImpact {
		t.Errorf("RecommendedImpact should equal SqrtImpact, got %v", est.RecommendedImpact)
	}
	if est.OptimalSlicesTWAP < 1 {
		t.Errorf("OptimalSlicesTWAP want >= 1, got %d", est.OptimalSlicesTWAP)
	}
}
