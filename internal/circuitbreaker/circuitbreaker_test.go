package circuitbreaker

import (
	"testing"
	"time"
)

func TestAllow_ClosedByDefault(t *testing.T) {
	b := New(3, 100*time.Millisecond)
	if err := b.Allow(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRecordFailure_OpensCircuit(t *testing.T) {
	b := New(3, 100*time.Millisecond)
	b.RecordFailure()
	b.RecordFailure()
	if b.State() != StateClosed {
		t.Fatal("expected closed before threshold")
	}
	b.RecordFailure()
	if b.State() != StateOpen {
		t.Fatal("expected open after threshold")
	}
}

func TestAllow_ReturnsErrWhenOpen(t *testing.T) {
	b := New(1, 100*time.Millisecond)
	b.RecordFailure()
	if err := b.Allow(); err != ErrCircuitOpen {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestAllow_HalfOpenAfterTimeout(t *testing.T) {
	b := New(1, 50*time.Millisecond)
	b.RecordFailure()
	time.Sleep(60 * time.Millisecond)
	if err := b.Allow(); err != nil {
		t.Fatalf("expected nil in half-open, got %v", err)
	}
	if b.State() != StateHalfOpen {
		t.Fatal("expected half-open state")
	}
}

func TestRecordSuccess_ClosesCircuit(t *testing.T) {
	b := New(1, 50*time.Millisecond)
	b.RecordFailure()
	time.Sleep(60 * time.Millisecond)
	_ = b.Allow() // transition to half-open
	b.RecordSuccess()
	if b.State() != StateClosed {
		t.Fatal("expected closed after success")
	}
	if err := b.Allow(); err != nil {
		t.Fatalf("expected nil after reset, got %v", err)
	}
}

func TestRecordSuccess_ResetsFailureCount(t *testing.T) {
	b := New(3, 100*time.Millisecond)
	b.RecordFailure()
	b.RecordFailure()
	b.RecordSuccess()
	b.RecordFailure()
	b.RecordFailure()
	// only 2 failures after reset, should still be closed
	if b.State() != StateClosed {
		t.Fatal("expected closed, failure count should have reset")
	}
}
