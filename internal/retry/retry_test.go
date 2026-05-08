package retry

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func okResponse(status int) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader("")),
	}
}

func TestDo_SuccessFirstAttempt(t *testing.T) {
	calls := 0
	resp, err := Do(DefaultPolicy(), func() (*http.Response, error) {
		calls++
		return okResponse(200), nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesOnRetryableStatus(t *testing.T) {
	p := Policy{MaxAttempts: 3, Delay: 0, RetryOn: []int{503}}
	calls := 0
	_, err := Do(p, func() (*http.Response, error) {
		calls++
		return okResponse(503), nil
	})
	if !errors.Is(err, ErrMaxAttempts) {
		t.Fatalf("expected ErrMaxAttempts, got %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestDo_SucceedsAfterRetry(t *testing.T) {
	p := Policy{MaxAttempts: 3, Delay: 0, RetryOn: []int{502}}
	calls := 0
	resp, err := Do(p, func() (*http.Response, error) {
		calls++
		if calls < 3 {
			return okResponse(502), nil
		}
		return okResponse(200), nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ReturnsErrorFromFn(t *testing.T) {
	p := Policy{MaxAttempts: 2, Delay: 0, RetryOn: []int{503}}
	sentinel := errors.New("dial error")
	_, err := Do(p, func() (*http.Response, error) {
		return nil, sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestDo_RespectsDelay(t *testing.T) {
	p := Policy{MaxAttempts: 2, Delay: 50 * time.Millisecond, RetryOn: []int{503}}
	start := time.Now()
	Do(p, func() (*http.Response, error) { //nolint:errcheck
		return okResponse(503), nil
	})
	if elapsed := time.Since(start); elapsed < 50*time.Millisecond {
		t.Errorf("expected at least 50ms delay, got %v", elapsed)
	}
}

func TestDo_ZeroMaxAttemptsTreatedAsOne(t *testing.T) {
	p := Policy{MaxAttempts: 0, Delay: 0, RetryOn: []int{503}}
	calls := 0
	Do(p, func() (*http.Response, error) { //nolint:errcheck
		calls++
		return okResponse(503), nil
	})
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}
