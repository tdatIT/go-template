package httpclient

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when a request is rejected by an open circuit.
var ErrCircuitOpen = errors.New("httpclient: circuit breaker open")

type cbState int

const (
	cbClosed   cbState = iota // normal — all requests pass through
	cbOpen                    // tripped — requests are rejected immediately
	cbHalfOpen                // probing — one request allowed to test recovery
)

// CBConfig configures circuit breaker behaviour.
type CBConfig struct {
	// MaxFailures is the number of consecutive failures that open the circuit.
	MaxFailures int
	// HalfOpenProbes is the number of consecutive successes in half-open state
	// needed to close the circuit again.
	HalfOpenProbes int
	// OpenTimeout is how long the circuit stays open before moving to half-open.
	OpenTimeout time.Duration
	// ShouldTrip reports whether a round-trip result counts as a failure.
	// Default (nil): only transport errors (non-nil err) count.
	// Override to also trip on HTTP 5xx:
	//   func(r *http.Response, err error) bool { return err != nil || r.StatusCode >= 500 }
	ShouldTrip func(resp *http.Response, err error) bool
}

type circuitBreaker struct {
	mu        sync.Mutex
	state     cbState
	failures  int
	successes int
	openedAt  time.Time
	cfg       CBConfig
}

func newCircuitBreaker(cfg CBConfig) *circuitBreaker {
	if cfg.ShouldTrip == nil {
		cfg.ShouldTrip = func(_ *http.Response, err error) bool { return err != nil }
	}
	return &circuitBreaker{cfg: cfg}
}

func (cb *circuitBreaker) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case cbClosed:
		return true
	case cbOpen:
		if time.Since(cb.openedAt) >= cb.cfg.OpenTimeout {
			cb.state = cbHalfOpen
			cb.successes = 0
			return true
		}
		return false
	case cbHalfOpen:
		return true
	}
	return false
}

func (cb *circuitBreaker) onSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	if cb.state == cbHalfOpen {
		cb.successes++
		if cb.successes >= cb.cfg.HalfOpenProbes {
			cb.state = cbClosed
		}
	}
}

func (cb *circuitBreaker) onFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.successes = 0
	if cb.state == cbHalfOpen {
		cb.state = cbOpen
		cb.openedAt = time.Now()
		return
	}
	cb.failures++
	if cb.failures >= cb.cfg.MaxFailures {
		cb.state = cbOpen
		cb.openedAt = time.Now()
	}
}

// cbTransport wraps an http.RoundTripper with circuit breaker logic.
type cbTransport struct {
	cb   *circuitBreaker
	next http.RoundTripper
}

func (t *cbTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !t.cb.allow() {
		return nil, ErrCircuitOpen
	}
	resp, err := t.next.RoundTrip(req)
	if t.cb.cfg.ShouldTrip(resp, err) {
		t.cb.onFailure()
	} else {
		t.cb.onSuccess()
	}
	return resp, err
}
