package tmdb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTokenBucket_RefillAndConsume(t *testing.T) {
	// 5 tokens capacity, 10 tokens/sec refill
	tb := NewTokenBucket(5, 10)
	assert.NotNil(t, tb)

	// Consume 5 initial tokens
	for i := 0; i < 5; i++ {
		start := time.Now()
		tb.Wait()
		assert.Less(t, time.Since(start), 50*time.Millisecond)
	}

	// Next token requires refill (approx 100ms for 1 token at 10/sec)
	start := time.Now()
	tb.Wait()
	elapsed := time.Since(start)
	assert.GreaterOrEqual(t, elapsed, 70*time.Millisecond)
}

func TestTokenBucket_NilSafe(t *testing.T) {
	var tb *TokenBucket
	assert.NotPanics(t, func() {
		tb.Wait()
	})
}
