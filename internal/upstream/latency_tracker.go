package upstream

import (
	"sync"
	"time"
)

// LatencyRecord holds rolling latency statistics for a single backend.
type LatencyRecord struct {
	mu      sync.Mutex
	samples []time.Duration
	max     int
}

func newLatencyRecord(maxSamples int) *LatencyRecord {
	return &LatencyRecord{
		samples: make([]time.Duration, 0, maxSamples),
		max:     maxSamples,
	}
}

// Record adds a new latency sample, evicting the oldest if at capacity.
func (r *LatencyRecord) Record(d time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.samples) >= r.max {
		r.samples = r.samples[1:]
	}
	r.samples = append(r.samples, d)
}

// Avg returns the mean latency across all recorded samples.
// Returns 0 if no samples have been recorded.
func (r *LatencyRecord) Avg() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.samples) == 0 {
		return 0
	}
	var total time.Duration
	for _, s := range r.samples {
		total += s
	}
	return total / time.Duration(len(r.samples))
}

// Count returns the number of recorded samples.
func (r *LatencyRecord) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.samples)
}

// LatencyTracker manages per-backend latency records.
type LatencyTracker struct {
	mu         sync.RWMutex
	records    map[string]*LatencyRecord
	maxSamples int
}

// NewLatencyTracker creates a tracker that keeps up to maxSamples per backend.
func NewLatencyTracker(maxSamples int) *LatencyTracker {
	if maxSamples <= 0 {
		maxSamples = 100
	}
	return &LatencyTracker{
		records:    make(map[string]*LatencyRecord),
		maxSamples: maxSamples,
	}
}

// Record adds a latency sample for the given backend URL.
func (t *LatencyTracker) Record(backend string, d time.Duration) {
	t.mu.Lock()
	rec, ok := t.records[backend]
	if !ok {
		rec = newLatencyRecord(t.maxSamples)
		t.records[backend] = rec
	}
	t.mu.Unlock()
	rec.Record(d)
}

// Avg returns the average latency for the given backend, or 0 if unknown.
func (t *LatencyTracker) Avg(backend string) time.Duration {
	t.mu.RLock()
	rec, ok := t.records[backend]
	t.mu.RUnlock()
	if !ok {
		return 0
	}
	return rec.Avg()
}

// Snapshot returns a map of backend URL to average latency.
func (t *LatencyTracker) Snapshot() map[string]time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make(map[string]time.Duration, len(t.records))
	for k, rec := range t.records {
		out[k] = rec.Avg()
	}
	return out
}
