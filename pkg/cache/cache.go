package cache

import (
	"container/list"
	"sync"
	"time"
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
	hits     int
	misses   int
}

type cacheItem struct {
	entry *entry
}

func New(capacity int) *Cache {
	return &Cache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		ll:       list.New(),
	}
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
			return nil, false
		}
		c.ll.MoveToFront(ele)
		c.hits++
		return item.entry.value, true
	}
	c.misses++
	return nil, false
}

func (c *Cache) Delete(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if ele, ok := c.items[key]; ok {
		c.ll.Remove(ele)
		delete(c.items, key)
	}
}

func (c *Cache) removeOldest() {
	ele := c.ll.Back()
	if ele != nil {
		item := ele.Value.(*cacheItem)
		delete(c.items, item.entry.key)
		c.ll.Remove(ele)
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
	return c.hits, c.misses, c.ll.Len()
}
