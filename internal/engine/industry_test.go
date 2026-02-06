package engine

import (
	"testing"

	"eve-flipper/internal/sde"
)

// sumJobCosts sums JobCost for all nodes where ShouldBuild && !IsBase, recursively.

func TestSumJobCosts_EmptyAndBase(t *testing.T) {
	a := &IndustryAnalyzer{}
	// Nil node would panic; we don't call with nil. Base node with ShouldBuild=false has no job cost.
	base := &MaterialNode{IsBase: true, ShouldBuild: false}
	if got := a.sumJobCosts(base); got != 0 {
		t.Errorf("sumJobCosts(base node) = %v, want 0", got)
	}
}

func TestSumJobCosts_SingleLevel(t *testing.T) {
	a := &IndustryAnalyzer{}
	root := &MaterialNode{IsBase: false, ShouldBuild: true, JobCost: 100.0, Children: nil}
	if got := a.sumJobCosts(root); got != 100 {
		t.Errorf("sumJobCosts(single node JobCost=100) = %v, want 100", got)
	}
}

func TestSumJobCosts_Tree(t *testing.T) {
	a := &IndustryAnalyzer{}
	// Root: JobCost 50, ShouldBuild true. Child1: 30, Child2: 20. Total = 50+30+20 = 100
	child1 := &MaterialNode{IsBase: false, ShouldBuild: true, JobCost: 30, Children: nil}
	child2 := &MaterialNode{IsBase: false, ShouldBuild: true, JobCost: 20, Children: nil}
	root := &MaterialNode{IsBase: false, ShouldBuild: true, JobCost: 50, Children: []*MaterialNode{child1, child2}}
	if got := a.sumJobCosts(root); got != 100 {
		t.Errorf("sumJobCosts(tree 50+30+20) = %v, want 100", got)
	}
}

func TestSumJobCosts_SkipsNonBuildAndBase(t *testing.T) {
	a := &IndustryAnalyzer{}
	// Root ShouldBuild=false -> no root JobCost. Child ShouldBuild=true -> count child only.
	child := &MaterialNode{IsBase: false, ShouldBuild: true, JobCost: 25, Children: nil}
	root := &MaterialNode{IsBase: false, ShouldBuild: false, JobCost: 100, Children: []*MaterialNode{child}}
	if got := a.sumJobCosts(root); got != 25 {
		t.Errorf("sumJobCosts(root skip, child count) = %v, want 25", got)
	}
}

func TestGetBlueprintInfo_DelegatesToSDE(t *testing.T) {
	// Minimal SDE: IndustryData with one product -> blueprint
	ind := sde.NewIndustryData()
	bp := &sde.Blueprint{ProductTypeID: 999, ProductQuantity: 2}
	ind.Blueprints[100] = bp
	ind.ProductToBlueprint[999] = 100

	a := &IndustryAnalyzer{SDE: &sde.Data{Industry: ind}}

	got, ok := a.GetBlueprintInfo(999)
	if !ok || got != bp {
		t.Errorf("GetBlueprintInfo(999) = %v, %v; want bp, true", got, ok)
	}
	_, ok = a.GetBlueprintInfo(888)
	if ok {
		t.Error("GetBlueprintInfo(888) should be false")
	}
}
