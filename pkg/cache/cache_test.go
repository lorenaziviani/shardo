package cache

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func newTestCache(cap int) *Cache {
	reg := prometheus.NewRegistry()
	return NewWithRegistry(cap, reg)
}

func TestCacheSetGetDelete(t *testing.T) {
	c := newTestCache(10)
	c.Set("foo", []byte("bar"), time.Second)
	val, ok := c.Get("foo")
	if !ok || string(val) != "bar" {
		t.Fatalf("expected bar, got %s", val)
	}
	c.Delete("foo")
	_, ok = c.Get("foo")
	if ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestCacheTTL(t *testing.T) {
	c := newTestCache(10)
	c.Set("foo", []byte("bar"), 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	_, ok := c.Get("foo")
	if ok {
		t.Fatal("expected key to expire")
	}
}

func TestCacheLRU(t *testing.T) {
	c := newTestCache(2)
	c.Set("a", []byte("1"), time.Second)
	c.Set("b", []byte("2"), time.Second)
	c.Set("c", []byte("3"), time.Second)
	if _, ok := c.Get("a"); ok {
		t.Fatal("expected a to be evicted")
	}
}
