package ratelimit

import (
	"context"
	"sync"
	"time"
)

// TokenBucket implements a token bucket rate limiter
type TokenBucket struct {
	capacity   int           // Maximum tokens (burst capacity)
	tokens     int           // Current available tokens
	refillRate time.Duration // Time between token refills
	lastRefill time.Time     // Last time tokens were refilled
	mutex      sync.Mutex    // Thread safety
}

// NewTokenBucket creates a new token bucket rate limiter
// SpaceTraders API: 2 requests per second, 30 burst capacity, 60 second window
func NewTokenBucket() *TokenBucket {
	now := time.Now()
	return &TokenBucket{
		capacity:   30,                     // 30 request burst limit
		tokens:     30,                     // Start with full bucket
		refillRate: 500 * time.Millisecond, // 2 per second = 500ms per token
		lastRefill: now,
		mutex:      sync.Mutex{},
	}
}

// NewCustomTokenBucket creates a token bucket with custom parameters
func NewCustomTokenBucket(capacity int, refillRate time.Duration) *TokenBucket {
	now := time.Now()
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity, // Start full
		refillRate: refillRate,
		lastRefill: now,
		mutex:      sync.Mutex{},
	}
}

// Wait blocks until a token is available or context is cancelled
func (tb *TokenBucket) Wait(ctx context.Context) error {
	for {
		if tb.Allow() {
			return nil
		}

		// Calculate wait time until next token
		tb.mutex.Lock()
		waitTime := tb.refillRate - time.Since(tb.lastRefill)
		tb.mutex.Unlock()

		if waitTime <= 0 {
			waitTime = tb.refillRate // Minimum wait
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Continue to try again
		}
	}
}

// Allow checks if a token is available and consumes it if so
func (tb *TokenBucket) Allow() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	tb.refill()

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// TryAllow attempts to consume a token without blocking
func (tb *TokenBucket) TryAllow() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	tb.refill()

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// Reset resets the token bucket to full capacity
func (tb *TokenBucket) Reset() {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	tb.tokens = tb.capacity
	tb.lastRefill = time.Now()
}

// GetState returns current bucket state for monitoring
func (tb *TokenBucket) GetState() BucketState {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	tb.refill()

	return BucketState{
		Tokens:     tb.tokens,
		Capacity:   tb.capacity,
		LastRefill: tb.lastRefill,
		RefillRate: tb.refillRate,
	}
}

// refill adds tokens based on elapsed time (must be called with mutex held)
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	// Calculate how many tokens to add based on elapsed time
	tokensToAdd := int(elapsed / tb.refillRate)

	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}
}

// BucketState represents the current state of a token bucket
type BucketState struct {
	Tokens     int           `json:"tokens"`
	Capacity   int           `json:"capacity"`
	LastRefill time.Time     `json:"last_refill"`
	RefillRate time.Duration `json:"refill_rate"`
}

// AvailableIn returns the duration until the next token will be available
func (bs BucketState) AvailableIn() time.Duration {
	if bs.Tokens > 0 {
		return 0
	}

	nextRefill := bs.LastRefill.Add(bs.RefillRate)
	waitTime := time.Until(nextRefill)

	if waitTime < 0 {
		return 0
	}

	return waitTime
}

// IsEmpty returns true if no tokens are available
func (bs BucketState) IsEmpty() bool {
	return bs.Tokens <= 0
}

// IsFull returns true if the bucket is at capacity
func (bs BucketState) IsFull() bool {
	return bs.Tokens >= bs.Capacity
}

// Utilization returns the current utilization as a percentage (0.0 to 1.0)
func (bs BucketState) Utilization() float64 {
	if bs.Capacity == 0 {
		return 0.0
	}
	return float64(bs.Tokens) / float64(bs.Capacity)
}
