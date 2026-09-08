package cache

import (
	"testing"
	"time"
)

func TestMemoryCache_SetGet(t *testing.T) {
	c := NewMemoryCache()

	c.Set("test-key", "hello world", 100*time.Millisecond)

	val, ok := c.Get("test-key")
	if !ok || val != "hello world" {
		t.Fatalf("expected 'hello world', got %v, ok=%v", val, ok)
	}

	time.Sleep(150 * time.Millisecond)

	_, ok = c.Get("test-key")
	if ok {
		t.Fatalf("expected key to be expired, but found it")
	}
}
