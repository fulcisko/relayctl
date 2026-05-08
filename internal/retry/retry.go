package retry

import (
	"errors"
	"net/http"
	"time"
)

// Policy defines retry behaviour for failed upstream requests.
type Policy struct {
	MaxAttempts int
	Delay       time.Duration
	RetryOn     []int // HTTP status codes that trigger a retry
}

// DefaultPolicy returns a sensible retry policy.
func DefaultPolicy() Policy {
	return Policy{
		MaxAttempts: 3,
		Delay:       100 * time.Millisecond,
		RetryOn:     []int{502, 503, 504},
	}
}

// ErrMaxAttempts is returned when all retry attempts are exhausted.
var ErrMaxAttempts = errors.New("retry: max attempts reached")

// DoFunc is the signature of the function passed to Do.
type DoFunc func() (*http.Response, error)

// Do executes fn up to p.MaxAttempts times, retrying when fn returns an error
// or a response whose status code is in p.RetryOn.
// The caller is responsible for closing the returned response body.
func Do(p Policy, fn DoFunc) (*http.Response, error) {
	if p.MaxAttempts < 1 {
		p.MaxAttempts = 1
	}

	var (
		resp *http.Response
		err  error
	)

	for attempt := 0; attempt < p.MaxAttempts; attempt++ {
		if attempt > 0 && p.Delay > 0 {
			time.Sleep(p.Delay)
		}

		resp, err = fn()
		if err == nil && !shouldRetry(p.RetryOn, resp.StatusCode) {
			return resp, nil
		}

		// Close body before retrying to avoid leaking connections.
		if resp != nil && err == nil {
			resp.Body.Close()
		}
	}

	if err != nil {
		return nil, err
	}
	return nil, ErrMaxAttempts
}

func shouldRetry(codes []int, status int) bool {
	for _, c := range codes {
		if c == status {
			return true
		}
	}
	return false
}
