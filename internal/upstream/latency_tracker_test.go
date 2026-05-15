package upstream

import (
	"testing"
	"time"
)

func TestLatencyRecord_AvgEmpty(t *testing.T) {
	r := newLatencyRecord(10)
	if got := r.Avg(); got != 0 {
		t.Fatalf("expected 0 for empty record, got %v", got)
	}
}

func TestLatencyRecord_AvgSingle(t *testing.T) {
	r := newLatencyRecord(10)
	r.Record(20 * time.Millisecond)
	if got := r.Avg(); got != 20*time.Millisecond {
		t.Fatalf("expected 20ms, got %v", got)
	}
}

func TestLatencyRecord_AvgMultiple(t *testing.T) {
	r := newLatencyRecord(10)
	r.Record(10 * time.Millisecond)
	r.Record(30 * time.Millisecond)
	if got := r.Avg(); got != 20*time.Millisecond {
		t.Fatalf("expected 20ms avg, got %v", got)
	}
}

func TestLatencyRecord_Eviction(t *testing.T) {
	r := newLatencyRecord(3)
	r.Record(100 * time.Millisecond)
	r.Record(100 * time.Millisecond)
	r.Record(100 * time.Millisecond)
	// fourth sample evicts first; new avg should reflect only last 3
	r.Record(10 * time.Millisecond)
	if r.Count() != 3 {
		t.Fatalf("expected 3 samples after eviction, got %d", r.Count())
	}
}

func TestLatencyTracker_RecordAndAvg(t *testing.T) {
	tr := NewLatencyTracker(50)
	tr.Record("http://backend1", 40*time.Millisecond)
	tr.Record("http://backend1", 60*time.Millisecond)
	if got := tr.Avg("http://backend1"); got != 50*time.Millisecond {
		t.Fatalf("expected 50ms, got %v", got)
	}
}

func TestLatencyTracker_UnknownBackend(t *testing.T) {
	tr := NewLatencyTracker(10)
	if got := tr.Avg("http://unknown"); got != 0 {
		t.Fatalf("expected 0 for unknown backend, got %v", got)
	}
}

func TestLatencyTracker_Snapshot(t *testing.T) {
	tr := NewLatencyTracker(10)
	tr.Record("http://a", 10*time.Millisecond)
	tr.Record("http://b", 20*time.Millisecond)
	snap := tr.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries in snapshot, got %d", len(snap))
	}
	if snap["http://a"] != 10*time.Millisecond {
		t.Errorf("unexpected avg for http://a: %v", snap["http://a"])
	}
	if snap["http://b"] != 20*time.Millisecond {
		t.Errorf("unexpected avg for http://b: %v", snap["http://b"])
	}
}

func TestLatencyTracker_DefaultMaxSamples(t *testing.T) {
	// maxSamples <= 0 should default to 100
	tr := NewLatencyTracker(0)
	if tr.maxSamples != 100 {
		t.Fatalf("expected default maxSamples=100, got %d", tr.maxSamples)
	}
}
