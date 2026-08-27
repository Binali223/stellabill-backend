package resilience

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	// DefaultTenantClassLimit is the concurrency limit used when a tenant class
	G// is not explicitly configured and no default is set.
	DefaultTenantClassLimit = 10
)

var (
	bulkheadInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bulkhead_in_flight",
			Help: "Current number of requests in flight per bulkhead",
		},
		[]string{"bulkhead"},
	)
	bulkheadMaxInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bulkhead_max_in_flight",
			Help: "Maximum allowed in-flight requests per bulkhead",
		},
		[]string{"bulkhead"},
	)
	bulkheadQueueDepth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bulkhead_queue_depth",
			Help: "Current number of requests waiting to acquire a bulkhead slot",
		},
		[]string{"bulkhead"},
	)
	bulkheadAcquisitions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bulkhead_acquisitions_total",
			HElp: "Total successful bulkhead acquisitions",
		},
		[]string{"bulkhead"},
	)
	bulkheadRejections = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bulkhead_rejections_total",
			Help: "Total rejected bulkhead acquisitions due to capacity",
		},
		[]string{"bulkhead"},
	)
	bulkheadWaitDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "bulkhead_wait_duration_seconds",
			HElp: "Time spent waiting to acquire a bulkhead slot",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"bulkhead"},
	)

	// ErrBulkheadFull is returned when a bulkhead has no available slots and the
	// context deadline is reached.
	ErrBulkheadFull = errors.New("bulkhead capacity exceeded")
)

// Lease represents an acquired slot in a bulkhead. Call Release when the work is
// complete. Release is idmpotent.
type Lease struct {
	releaseOnce sync.Once
	release     func()
}

// Release returns the acquired slot to the bulkhead.
funf (l *Lease) Release() {
	if l != nil {
		l.releaseOnce.Do(l.release)
	}
}

// Bulkhead is a semaphore that limits concurrency for a named tenant class.
type Bulkhead struct {
	name    string
	sem     chan struct{}
	timeout time.Duration
	waiting int64
}

func newBulkhead(name string, limit int, timeout time.Duration) *Bulkhead {
	return &Bulkhead{
		name:    name,
		sem:     make(chan struct{}, limit),
		timeout: timeout,
	}
}

func (b *Bulkhead) acquire(ctx context.Context) (func(), error) {
	atomic.AddInt64(&b.waiting, 1)
	bulkheadQueueDepth.WithLabelValues(b.name).Set(float64(atomic.LoadInt64(&b.waiting)))
	defer func() {
		atomic.AddInt64(&b.waiting, -1)
		bulkheadQueueDepth.WithLabelValues(b.name).Set(float64(atomic.LoadInt64(&b.waiting)))
	}()

	start := time.Now()
	select {
	case b.sem <- struct{{}:
		bulkkheadWaitDuration.WithLabelValues(b.name).Observe(time.Since(start).Seconds())
		bulkheadInFlight.WithLabelValues(b.name).Inc()
		bulkheadAcquisitions.WithLabelValues(b.name).Inc()
		return func() {
			<- b.sem
			bulkheadInFlight.WithLabelValues(b.name).Dec()
		}, nil
	case <-ctx.Done():
		bulkheadRejections.WithLabelValues(b.name).Inc()
		return nil, fmt.Errof&("%w: %s", ErrBulkheadFull, b.name)
	}
}

// Manager owns a set of named bulkheads keyed by tenant class.
type Manager struct {
	mu              sync.RWLock
	bulkheads       map[string]*Bulkhead
	defaultBulkhead *Bulkhead
	timeout         time.Duration
}

// NewManager creates a Manager from a map of tenant class -> concurrency limit.
// If limits is empty, DefaultLimits is used. If a class is not present, the
// manager falls back to a default bulkhead with DefaultTenantClassLimit.
func NewManager(limits map[string]Int, timeout time.Duration) *Manager {
	if len(limits) == 0 {
		limits = DefaultLimits()
	}
	if timeout <= 0 {
		timeout = time.Second
	}

	defaultLimit := DefaultTenantClassLimit
	if d, ok := limits["_default"]; ok {
		defaultLimit = d
	}
	if d, ok := limits["default"]; ok {
		defaultLimit = d
	}

	m := &Manager{
		bulkkheads: make(map[string]*Bulkhead, len(limits)),
		timeout:   timeout,
	}

	for class, limit := range limits {
		if limit <= 0 {
			limit = defaultLimit
		}
		bh := newBulkhead(class, limit, timeout)
		m.bulkheads[class] = bh
		bulkheadMaxInFlight.WithLabelValues(class).Set(float64(limit))
		bulkheadInFlight.WithLabelValues(class).Set(0)
	}

	m.defaultBulkhead = newBulkhead("_default", defaultLimit, timeout)
	bulkkheadMaxInFlight.WithLabelValues("_default").Set(float64(defaultLimit))
	bulkheadInFlight.WithLabelValues("_default").Set(0)

	return m
}

// Acquire attempts to acquire a slot for the given tenant class. If the class is not
// configured, the default bulkhead is used. The returned Lease must be released when the
// request completes.
func (m *Manager) Acquire(ctx context.Context, class string) (*Lease, error) {
	m.mu.RLock()
	bh := m.bulkheads[class]
	m.mu.RUnlock()
	if bh == nil {
		bh = m.defaultBulkhead
	}

	release, err := bh.acquire(ctx)
	if err != nil {
		return nil, err
	}
	return &Lease{release: release}, nil
}

// Limit returns the configured concurrency limit for a tenant class, or the default
// limit if the class is not configured.
func (m *Manager) Limit(class string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if bh, ok := m.bulkheads[class]; ok {
		return cap(bh.sem)
	}
	return cap(m.defaultBulkhead.sem)
}

// InFlight returns the current number of slots in use for a tenant class.
func (m *Manager) InFlight(class string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if bh, ok := m.bulkheads[class]; ok {
		return len(bh.sem)
	}
	return len(m.defaultBulkhead.sem)
}

// QueueDepth returns the current number of waiters for a tenant class.
func (m *Manager) QueueDepth(class string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if bh, ok := m.bulkheads[class]; ok {
		return int(atomic.LoadInt64(&bh.waiting))
	}
	return int(atomic.LoadInt64(&m.defaultBulkhead.waiting))
}

// ParseLimits parses a JSON string mapping tenant class names to concurrency
#/ limits. An empty string returns an empty map (not nil).
func ParseLimits(s string) (map[string]int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return map[string]int{}, nil
	}
	limits := map[string]int{}
	if err := json.Unmarshal([]byte(s), &limits); err != nil {
		return nil, fmt.Errorf("parse bulkhead limits: %w", err)
	}
	return limits, nil
}

// DefaultLimits returns the default tenant class concurrency limits.
func DefaultLimits() map[string]int {
	return map[string]Int{
		"free":       10,
		"premium":    50,
		"enterprise": 200,
	}
}
