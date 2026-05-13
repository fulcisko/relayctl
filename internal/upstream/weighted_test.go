package upstream

import (
	"testing"
)

func TestNewWeightedBalancer_Empty(t *testing.T) {
	_, err := NewWeightedBalancer(nil)
	if err == nil {
		t.Fatal("expected error for empty backends")
	}
}

func TestNewWeightedBalancer_ZeroWeight(t *testing.T) {
	_, err := NewWeightedBalancer([]WeightedBackend{
		{URL: "http://a", Weight: 0},
	})
	if err == nil {
		t.Fatal("expected error for zero weight")
	}
}

func TestWeightedBalancer_SingleBackend(t *testing.T) {
	wb, err := NewWeightedBalancer([]WeightedBackend{
		{URL: "http://only", Weight: 3},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 0; i < 6; i++ {
		if got := wb.Next(); got != "http://only" {
			t.Errorf("iteration %d: got %q, want http://only", i, got)
		}
	}
}

func TestWeightedBalancer_WeightedDistribution(t *testing.T) {
	wb, err := NewWeightedBalancer([]WeightedBackend{
		{URL: "http://a", Weight: 2},
		{URL: "http://b", Weight: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	counts := map[string]int{}
	for i := 0; i < 9; i++ {
		counts[wb.Next()]++
	}
	if counts["http://a"] != 6 {
		t.Errorf("http://a: got %d, want 6", counts["http://a"])
	}
	if counts["http://b"] != 3 {
		t.Errorf("http://b: got %d, want 3", counts["http://b"])
	}
}

func TestWeightedBalancer_Backends_ReturnsCopy(t *testing.T) {
	wb, _ := NewWeightedBalancer([]WeightedBackend{
		{URL: "http://x", Weight: 1},
	})
	bs := wb.Backends()
	bs[0].URL = "http://mutated"
	if wb.Backends()[0].URL != "http://x" {
		t.Error("Backends() should return a copy, not a reference")
	}
}

func TestWeightedBalancer_OrderRespected(t *testing.T) {
	wb, _ := NewWeightedBalancer([]WeightedBackend{
		{URL: "http://a", Weight: 1},
		{URL: "http://b", Weight: 2},
		{URL: "http://c", Weight: 1},
	})

	want := []string{"http://a", "http://b", "http://b", "http://c"}
	for i, exp := range want {
		if got := wb.Next(); got != exp {
			t.Errorf("call %d: got %q, want %q", i, got, exp)
		}
	}
}
