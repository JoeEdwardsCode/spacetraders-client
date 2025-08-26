package unit

import (
	"context"
	"spacetraders-client/internal/ratelimit"
	"testing"
	"time"
)

func TestTokenBucket(t *testing.T) {
	t.Run("Basic Allow", func(t *testing.T) {
		bucket := ratelimit.NewTokenBucket()
		
		// Should allow immediately
		if !bucket.Allow() {
			t.Error("New bucket should allow request")
		}
		
		// Should still have tokens left
		state := bucket.GetState()
		if state.Tokens <= 0 {
			t.Error("Bucket should have tokens remaining")
		}
	})

	t.Run("Burst Limit", func(t *testing.T) {
		bucket := ratelimit.NewTokenBucket()
		state := bucket.GetState()
		capacity := state.Capacity
		
		// Consume all tokens
		for i := 0; i < capacity; i++ {
			if !bucket.Allow() {
				t.Errorf("Should allow request %d/%d", i+1, capacity)
			}
		}
		
		// Should be empty now
		if bucket.Allow() {
			t.Error("Bucket should be empty after consuming all tokens")
		}
		
		state = bucket.GetState()
		if !state.IsEmpty() {
			t.Errorf("Bucket should report as empty, but has %d tokens", state.Tokens)
		}
	})

	t.Run("Refill Rate", func(t *testing.T) {
		// Create bucket with fast refill for testing
		bucket := ratelimit.NewCustomTokenBucket(2, 100*time.Millisecond)
		
		// Consume all tokens
		bucket.Allow()
		bucket.Allow()
		
		if bucket.Allow() {
			t.Error("Should be empty")
		}
		
		// Wait for refill
		time.Sleep(150 * time.Millisecond)
		
		// Should have one token now
		if !bucket.Allow() {
			t.Error("Should have refilled at least one token")
		}
	})

	t.Run("Wait Method", func(t *testing.T) {
		bucket := ratelimit.NewCustomTokenBucket(1, 200*time.Millisecond)
		
		// Consume the only token
		bucket.Allow()
		
		// Wait should block and then succeed
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		
		start := time.Now()
		err := bucket.Wait(ctx)
		elapsed := time.Since(start)
		
		if err != nil {
			t.Errorf("Wait should succeed: %v", err)
		}
		
		if elapsed < 150*time.Millisecond {
			t.Errorf("Wait should take at least 150ms, took %v", elapsed)
		}
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		bucket := ratelimit.NewCustomTokenBucket(1, 10*time.Second)
		
		// Consume the only token
		bucket.Allow()
		
		// Cancel context immediately
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		
		err := bucket.Wait(ctx)
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	})

	t.Run("Reset", func(t *testing.T) {
		bucket := ratelimit.NewTokenBucket()
		
		// Consume some tokens
		for i := 0; i < 20; i++ {
			bucket.Allow()
		}
		
		state := bucket.GetState()
		tokensBefore := state.Tokens
		
		// Reset should restore full capacity
		bucket.Reset()
		
		state = bucket.GetState()
		if state.Tokens <= tokensBefore {
			t.Errorf("Reset should increase tokens from %d to %d", tokensBefore, state.Tokens)
		}
		
		if !state.IsFull() {
			t.Error("Bucket should be full after reset")
		}
	})

	t.Run("State Information", func(t *testing.T) {
		bucket := ratelimit.NewCustomTokenBucket(10, 100*time.Millisecond)
		
		state := bucket.GetState()
		
		// Test utilization
		if state.Utilization() != 1.0 {
			t.Errorf("New bucket should be 100%% utilized, got %f", state.Utilization())
		}
		
		// Consume half the tokens
		for i := 0; i < 5; i++ {
			bucket.Allow()
		}
		
		state = bucket.GetState()
		expectedUtil := 0.5
		if state.Utilization() != expectedUtil {
			t.Errorf("Half-consumed bucket should be 50%% utilized, got %f", state.Utilization())
		}
		
		// Test AvailableIn
		availableIn := state.AvailableIn()
		if state.Tokens > 0 && availableIn != 0 {
			t.Errorf("Should be available immediately when tokens exist, got %v", availableIn)
		}
	})

	t.Run("Thread Safety", func(t *testing.T) {
		bucket := ratelimit.NewCustomTokenBucket(100, 10*time.Millisecond)
		
		// Start multiple goroutines consuming tokens
		results := make(chan bool, 200)
		
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 20; j++ {
					results <- bucket.Allow()
				}
			}()
		}
		
		// Collect results
		allowed := 0
		denied := 0
		
		for i := 0; i < 200; i++ {
			if <-results {
				allowed++
			} else {
				denied++
			}
		}
		
		// Should have consumed exactly the initial capacity
		if allowed > 100 {
			t.Errorf("Should not allow more than capacity (100), allowed %d", allowed)
		}
		
		if allowed < 90 {
			t.Errorf("Should allow most requests initially, only allowed %d", allowed)
		}
		
		t.Logf("Thread safety test: %d allowed, %d denied", allowed, denied)
	})
}

func BenchmarkTokenBucket(b *testing.B) {
	bucket := ratelimit.NewTokenBucket()
	
	b.ResetTimer()
	
	b.Run("Allow", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bucket.TryAllow()
		}
	})
	
	b.Run("GetState", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bucket.GetState()
		}
	})
	
	b.Run("Concurrent Allow", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				bucket.TryAllow()
			}
		})
	})
}

func TestTokenBucketEdgeCases(t *testing.T) {
	t.Run("Zero Capacity", func(t *testing.T) {
		bucket := ratelimit.NewCustomTokenBucket(0, 100*time.Millisecond)
		
		if bucket.Allow() {
			t.Error("Zero capacity bucket should never allow")
		}
		
		state := bucket.GetState()
		if state.Capacity != 0 {
			t.Errorf("Expected zero capacity, got %d", state.Capacity)
		}
		
		if state.Utilization() != 0.0 {
			t.Errorf("Zero capacity utilization should be 0.0, got %f", state.Utilization())
		}
	})
	
	t.Run("Very Fast Refill", func(t *testing.T) {
		bucket := ratelimit.NewCustomTokenBucket(1, 1*time.Nanosecond)
		
		// Should refill very quickly
		bucket.Allow() // Consume token
		
		time.Sleep(1 * time.Millisecond) // Wait much longer than refill rate
		
		if !bucket.Allow() {
			t.Error("Should have refilled with very fast refill rate")
		}
	})
	
	t.Run("Very Slow Refill", func(t *testing.T) {
		bucket := ratelimit.NewCustomTokenBucket(1, 1*time.Hour)
		
		bucket.Allow() // Consume token
		
		// Should not refill quickly
		time.Sleep(10 * time.Millisecond)
		
		if bucket.Allow() {
			t.Error("Should not have refilled with very slow refill rate")
		}
	})
}