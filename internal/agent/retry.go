package agent

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk"
)

var DefaultRetry = adk.ModelRetryConfig{
	MaxRetries:  3,
	IsRetryAble: isRetryableError,
	BackoffFunc: rateLimitBackoff,
}

// isRetryableError returns true for 429 rate-limit and 5xx server errors.
// Hard errors (auth failures, invalid parameters) are not retried.
func isRetryableError(_ context.Context, err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "429") ||
		strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "too many requests") ||
		strings.Contains(msg, "server error") ||
		strings.Contains(msg, "502") ||
		strings.Contains(msg, "503") ||
		strings.Contains(msg, "504")
}

// rateLimitBackoff returns an exponential delay starting at 5 s (max 60 s, ±25% jitter).
// The doubling strategy gives rate-limit quotas time to replenish between retries.
func rateLimitBackoff(_ context.Context, attempt int) time.Duration {
	const base = 5 * time.Second
	const maxDelay = 60 * time.Second

	delay := base * time.Duration(1<<uint(attempt-1))
	if delay > maxDelay {
		delay = maxDelay
	}

	// Add ±25% jitter to avoid thundering-herd retries across goroutines.
	jitter := time.Duration(rand.Int63n(int64(delay / 2)))
	if rand.Intn(2) == 0 {
		delay += jitter / 2
	} else {
		delay -= jitter / 2
	}

	if delay < time.Second {
		delay = time.Second
	}
	return delay
}
