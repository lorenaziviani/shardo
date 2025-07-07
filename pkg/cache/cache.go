package cache

import (
	"container/list"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type entry struct {
	key     string
	value   []byte
	expires time.Time
}

type Cache struct {
	capacity int
	items    map[string]*list.Element
	ll       *list.List
	lock     sync.Mutex

	hits       int32
	misses     int32
	ttlExpired int32

	hitsMetric       prometheus.Counter
	missesMetric     prometheus.Counter
	ttlExpiredMetric prometheus.Counter
	sizeMetric       prometheus.Gauge

	registry prometheus.Registerer
}

type cacheItem struct {
	entry *entry
}

func NewWithRegistry(capacity int, reg prometheus.Registerer) *Cache {
	c := &Cache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		ll:       list.New(),
		registry: reg,
	}
	c.initMetrics()
	return c
}

func New(capacity int) *Cache {
	return NewWithRegistry(capacity, prometheus.DefaultRegisterer)
}

func (c *Cache) initMetrics() {
	c.hitsMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "Total cache hits",
	})
	c.missesMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "Total cache misses",
	})
	c.ttlExpiredMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_ttl_expired_total",
		Help: "Total TTL expired",
	})
	c.sizeMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cache_size",
		Help: "Current cache size",
	})
	c.registry.MustRegister(c.hitsMetric, c.missesMetric, c.ttlExpiredMetric, c.sizeMetric)
}

func (c *Cache) Set(key string, value []byte, ttl time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if ele, ok := c.items[key]; ok {
		item := ele.Value.(*cacheItem)
		item.entry.value = value
		item.entry.expires = time.Now().Add(ttl)
		c.ll.MoveToFront(ele)
		return
	}
	ent := &entry{key: key, value: value, expires: time.Now().Add(ttl)}
	item := &cacheItem{entry: ent}
	ele := c.ll.PushFront(item)
	c.items[key] = ele
	if c.ll.Len() > c.capacity {
		c.removeOldest()
	}
	c.sizeMetric.Set(float64(c.ll.Len()))
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if ele, ok := c.items[key]; ok {
		item := ele.Value.(*cacheItem)
		if time.Now().After(item.entry.expires) {
			c.ll.Remove(ele)
			delete(c.items, key)
			c.misses++
			c.ttlExpired++
			c.missesMetric.Inc()
			c.ttlExpiredMetric.Inc()
			c.sizeMetric.Set(float64(c.ll.Len()))
			return nil, false
		}
		c.ll.MoveToFront(ele)
		c.hits++
		c.hitsMetric.Inc()
		return item.entry.value, true
	}
	c.misses++
	c.missesMetric.Inc()
	return nil, false
}

func (c *Cache) Delete(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if ele, ok := c.items[key]; ok {
		c.ll.Remove(ele)
		delete(c.items, key)
		c.sizeMetric.Set(float64(c.ll.Len()))
	}
}

func (c *Cache) removeOldest() {
	ele := c.ll.Back()
	if ele != nil {
		item := ele.Value.(*cacheItem)
		delete(c.items, item.entry.key)
		c.ll.Remove(ele)
		c.sizeMetric.Set(float64(c.ll.Len()))
	}
}

func (c *Cache) Len() int {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.ll.Len()
}

func (c *Cache) Metrics() (hits, misses, size int) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return int(c.hits), int(c.misses), c.ll.Len()
}
