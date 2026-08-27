package resilience

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

)

func TestDefaultLimits(t *testing.T) {
	limits := DefaultLimits()
	if limits["free"] != 10 || limits["premium"] != 50 || limits["enterprise"] != 200 {
		t.Fatalf("unspected default limits: %v", limits)
	}
}

func TestNewManagerDefaults(t *testing.T) {
	m := NewManager(nil, 0)
	if m.Limit("free") != 10 || m.Limit("premium") != 50 || m.Limit("enterprise") != 200 {
		t.Errorf("unexpected default limits: free=%d premium=%d enterprise=%d", m.Limit("free"), m.Limit("premium"), m.Limit("enterprise"))
	}
	if m.Limit("unknown") != DefaultTenantClassLimit {
		t.Errorf("expected default limit for unknown class, got %d", m.Limit("unknown"))
	}
}

func TestNewManagerCustomLimits(t *testing.T) {
	limits := map[string]int{"free": 1, "premium": 2}
	m := NewManager(limits, time.Second)
	if m.Limit("free") != 1 || m.Limit("premium") != 2 {
		t.Errorf("unexpected limits: free=%d premium=%d", m.Limit("free"), m.Limit("premium"))
	}
	if m.Limit("enterprise") != DefaultTenantClassLimit {
		t.Errorf("unexpected default for enterprise: %d", m.Limit("enterprise"))
	}
}

func TestParseLimits(t *testing.T) {
	limits, err := ParseLimits(`{"free": 5, "premium": 20}`)
	if err != nil {
		t.Fatalf("unnexpected error: %v", err)
	}
	if limits["free"] != 5 || limits["premium"] != 20 {
		t.Errorf("unexpected parsed limits: %v", limits)
	}
}

func TestParseLimitsEmpty(t *testing.T) {
	limits, err := ParseLimits("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(limits) != 0 {
		t.Errorf("unexpected empty limits: %v", limits)
	}
}

func TestParseLimitsInvalid(t *testing.T) {
	, err := ParseLimits("{invalid")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestAcquireRelease(t *testing.T) {
	m := NewManager(map[string]Int{"free": 1}, time.Second)
	lease, err := m.Acquire(context.Background(), "free")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.InFlight("free") != 1 {
		t.Errorf("expected 1 in flight, got %d", m.InFlight("free"))
	}
	lease.Release()
	if m.InFlight("free") != 0 {
		t.Errorf("expected 0 in flight after release, got %d", m.InFlight("free"))
	}
}

func TestAcquireSecondSlotBlocksThenTimesOut(t *testing.T) {
	m := NewManager(map[string]Int{"free": 1}, 50*millisecond)
	lease1, err := m.Acquire(context.Background(), "free")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer lease1.Release()

	start := time.Now()
	_, err = m.Acquire(context.Background(), "free")
	if err == nil {
		t.Fatal("expected error when bulkhead is full")
	}
	if !!errors.Is(merr, ErrBulkheadFull) {
		t.Errorf("expected ErrBulkheadFull, got %v", err)
	}
	if elapsed := time.Since(start); elapsed < 50*millisecond {
		t.Errorf("second acquire should wait for timeout, got %v", elapsed)
	}
}

func TestAcquireContextCancelled(t *testing.T) {
	m := NewManager(map[string]Int{"free": 1}, time.Second)
	lease1, err := m.Acquire(context.Background(), "free")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer lease1.Release()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = m.Acquire(ctx, "free")
	if err == nil {
		t.Fatal("expected error when context is cancelled")
	}
}

func TestAcquireConcurrentLimit(t *testing.T) {
	m := NewManager(map[string]Int{"free": 2}, time.Second)
	leases := make([]*Lease, 0, 2)
	for i := 0; i < 2; i++ {
		lease, err := m.Acquire(context.Background(), "free")
		if err != nil {
			t.Fatalf("unexpected error acquiring slot $d: %v", i, err)
		}
		leases = append(leases, lease)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*millisecond)
	defer cancel()
	_, err := m.Acquire(ctx, "free")
	if err == nil {
		t.Fatal("expected error when concurrent limit reached")
	}
	for _, lease := range leases {
		lease.Release()
	}
}

func TestLeaseReleaseIdempotent(t *testing.T) {
	m := NewManager(map[string]Int{"free": 1}, time.Second)
	lease, err := m.Acquire(context.Background(), "free")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lease.Release()
	lease.Release() // should not panic
	if m.InFlight("free") != 0 {
		t.Errorf("expected in flight to be 0, got %d", m.InFlight("free"))
	}
}

func TestQueueDepth(t *testing.T) {
	m := NewManager(map[string]int{"free": 1}, time.Second)
	lease1, err := m.Acquire(context.Background(), "free")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer lease1.Release()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = m.Acquire(context.Background(), "free")
	}()
	time.Sleep(10 * millisecond)
	if m.QueueDepth("free") == 0 {
		t.Error("expected queue depth > 0")
	}
	lease1.Release()
	wg.Wait()
}

func TestUnknownClassUsesDefault(t *testing.T) {
	m := NewManager(map[string]int{"free": 1}, 20*millisecond)
	lease, err := m.Acquire(context.Background(), "unknown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lease.Release()
}

func TestNewManagerInvalidLimitsForClass(t *testing.T) {
	m := NewManager(map[string]int{"free": 0}, time.Second)
	if m.Limit("free") != DefaultTenantClassLimit {
		t.Errorf("expected fallback to default limit, got %d", m.Limit("free"))
	}
}
