package middleware

import (
	"testing"
	"time"
)

func TestTokenBucket_Allow(t *testing.T) {
	b := &tokenBucket{rate: 10, capacity: 5, count: 5, last: time.Now()}
	for i := 0; i < 5; i++ {
		if !b.allow() {
			t.Fatalf("allow() should return true on call %d", i+1)
		}
	}
	if b.allow() {
		t.Fatal("allow() should return false when bucket is empty")
	}
}

func TestTokenBucket_Refill(t *testing.T) {
	b := &tokenBucket{rate: 100, capacity: 5, count: 0, last: time.Now().Add(-1 * time.Second)}
	if !b.allow() {
		t.Fatal("allow() should refill tokens after 1 second at 100 rps")
	}
}

func TestTokenBucket_Capacity(t *testing.T) {
	b := &tokenBucket{rate: 5, capacity: 3, count: 3, last: time.Now()}
	if b.tokenCount() != 3 {
		t.Fatalf("expected 3 tokens, got %d", b.tokenCount())
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{100, "100"},
		{-1, "0"},
	}
	for _, tt := range tests {
		if got := itoa(tt.n); got != tt.want {
			t.Errorf("itoa(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}
