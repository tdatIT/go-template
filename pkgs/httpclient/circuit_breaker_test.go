package httpclient

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// --- state machine unit tests (white-box, same package) ---

func TestCircuitBreaker_closedAllowsRequests(t *testing.T) {
	cb := newCircuitBreaker(CBConfig{MaxFailures: 2, HalfOpenProbes: 1, OpenTimeout: time.Hour})
	require.True(t, cb.allow())
}

func TestCircuitBreaker_opensAfterMaxFailures(t *testing.T) {
	cb := newCircuitBreaker(CBConfig{MaxFailures: 2, HalfOpenProbes: 1, OpenTimeout: time.Hour})

	cb.onFailure()
	require.True(t, cb.allow()) // still closed after 1

	cb.onFailure()
	require.False(t, cb.allow()) // open after 2
}

func TestCircuitBreaker_successResetsFailureCount(t *testing.T) {
	cb := newCircuitBreaker(CBConfig{MaxFailures: 2, HalfOpenProbes: 1, OpenTimeout: time.Hour})

	cb.onFailure()
	cb.onSuccess() // reset
	cb.onFailure()
	require.True(t, cb.allow()) // 1 failure (not 2) — still closed
}

func TestCircuitBreaker_halfOpenAfterTimeout(t *testing.T) {
	cb := newCircuitBreaker(CBConfig{MaxFailures: 1, HalfOpenProbes: 1, OpenTimeout: 30 * time.Millisecond})

	cb.onFailure()
	require.False(t, cb.allow()) // open

	time.Sleep(40 * time.Millisecond)
	require.True(t, cb.allow()) // half-open: one probe allowed
}

func TestCircuitBreaker_closesAfterHalfOpenProbes(t *testing.T) {
	cb := newCircuitBreaker(CBConfig{MaxFailures: 1, HalfOpenProbes: 2, OpenTimeout: 10 * time.Millisecond})

	cb.onFailure()
	time.Sleep(15 * time.Millisecond)

	// allow() triggers the Open→HalfOpen transition; onSuccess() follows each probe.
	require.True(t, cb.allow())
	cb.onSuccess() // probe 1 of 2 — still half-open

	require.True(t, cb.allow())
	cb.onSuccess() // probe 2 of 2 — closes

	require.Equal(t, cbClosed, cb.state)
	require.True(t, cb.allow())
}

func TestCircuitBreaker_failureInHalfOpenReopens(t *testing.T) {
	cb := newCircuitBreaker(CBConfig{MaxFailures: 1, HalfOpenProbes: 2, OpenTimeout: 10 * time.Millisecond})

	cb.onFailure()
	time.Sleep(15 * time.Millisecond)

	cb.onFailure() // failure while half-open → back to open
	require.False(t, cb.allow())
	require.Equal(t, cbOpen, cb.state)
}

// --- integration tests through the full Caller ---

func TestCaller_circuitBreaker_blocksAfterFailures(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New(Config{
		BaseURL: srv.URL,
		CircuitBreaker: &CBConfig{
			MaxFailures:    2,
			HalfOpenProbes: 1,
			OpenTimeout:    time.Hour,
			// trip on 5xx too
			ShouldTrip: func(r *http.Response, err error) bool {
				return err != nil || (r != nil && r.StatusCode >= 500)
			},
		},
	})

	_, _ = c.MakeRequest().Get("/")
	_, _ = c.MakeRequest().Get("/")

	_, err := c.MakeRequest().Get("/")
	require.ErrorIs(t, err, ErrCircuitOpen)
}

func TestCaller_circuitBreaker_recoversThroughHalfOpen(t *testing.T) {
	var healthy atomic.Bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if healthy.Load() {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	trip := func(r *http.Response, err error) bool {
		return err != nil || (r != nil && r.StatusCode >= 500)
	}

	c := New(Config{
		BaseURL: srv.URL,
		CircuitBreaker: &CBConfig{
			MaxFailures:    1,
			HalfOpenProbes: 1,
			OpenTimeout:    30 * time.Millisecond,
			ShouldTrip:     trip,
		},
	})

	// open the circuit
	_, _ = c.MakeRequest().Get("/")
	_, err := c.MakeRequest().Get("/")
	require.ErrorIs(t, err, ErrCircuitOpen)

	// wait for half-open, then fix the server
	time.Sleep(40 * time.Millisecond)
	healthy.Store(true)

	// probe succeeds → circuit closes
	resp, err := c.MakeRequest().Get("/")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	// circuit is now closed — request goes through normally
	resp, err = c.MakeRequest().Get("/")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestCaller_noCircuitBreaker_behavesNormally(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL}) // no CircuitBreaker

	for range 3 {
		resp, err := c.MakeRequest().Get("/")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
	}
	require.Equal(t, int32(3), hits.Load())
}

func TestCircuitBreaker_defaultShouldTrip_ignoresHTTPErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New(Config{
		BaseURL: srv.URL,
		CircuitBreaker: &CBConfig{
			MaxFailures:    2,
			HalfOpenProbes: 1,
			OpenTimeout:    time.Hour,
			// ShouldTrip nil → default: only transport errors trip the circuit
		},
	})

	// 5xx responses do NOT count as failures with the default ShouldTrip
	for range 5 {
		resp, err := c.MakeRequest().Get("/")
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode())
	}
	// circuit never opens — no ErrCircuitOpen
	_, err := c.MakeRequest().Get("/")
	require.False(t, errors.Is(err, ErrCircuitOpen))
}
