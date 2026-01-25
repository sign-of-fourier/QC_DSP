// internal/core/metrics.go
// Lightweight metrics helper for both prod & sim.
// No Prometheus dependency yet—just atomic counters and snapshots.
package core

import (
	"sync/atomic"
	"time"
)

// Metrics holds simple in-memory counters.
// You can later expose these to Prometheus or logs.
type Metrics struct {
	// Total number of auctions received (prod) or simulated (sim)
	totalAuctions uint64

	// How many times the engine produced a bid (ShouldBid == true)
	totalBids uint64

	// How many errors happened in evaluation
	totalErrors uint64

	// Sum of latencies in microseconds (so you can compute avg)
	totalLatencyMicros uint64
}

// NewMetrics returns a fresh metrics struct.
func NewMetrics() *Metrics {
	return &Metrics{}
}

// IncAuctions increments the auction counter.
func (m *Metrics) IncAuctions() {
	atomic.AddUint64(&m.totalAuctions, 1)
}

// IncBids increments the bid counter.
func (m *Metrics) IncBids() {
	atomic.AddUint64(&m.totalBids, 1)
}

// IncErrors increments the error counter.
func (m *Metrics) IncErrors() {
	atomic.AddUint64(&m.totalErrors, 1)
}

// ObserveLatency records a latency sample.
func (m *Metrics) ObserveLatency(d time.Duration) {
	us := uint64(d.Microseconds())
	atomic.AddUint64(&m.totalLatencyMicros, us)
}

// Snapshot returns a consistent read of all counters.
type MetricsSnapshot struct {
	TotalAuctions      uint64
	TotalBids          uint64
	TotalErrors        uint64
	AverageLatencyMs   float64
	TotalLatencyMicros uint64
}

func (m *Metrics) Snapshot() MetricsSnapshot {
	auctions := atomic.LoadUint64(&m.totalAuctions)
	bids := atomic.LoadUint64(&m.totalBids)
	errs := atomic.LoadUint64(&m.totalErrors)
	lat := atomic.LoadUint64(&m.totalLatencyMicros)

	avgMs := 0.0
	if auctions > 0 {
		avgMs = float64(lat) / float64(auctions) / 1000.0
	}

	return MetricsSnapshot{
		TotalAuctions:      auctions,
		TotalBids:          bids,
		TotalErrors:        errs,
		AverageLatencyMs:   avgMs,
		TotalLatencyMicros: lat,
	}
}
