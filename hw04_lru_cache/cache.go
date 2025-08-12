package hw04lrucache

import (
	"sync"
)

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
	mutex    sync.RWMutex
}
type cacheItem struct {
	key   Key
	value interface{}
}

func (c *lruCache) Set(key Key, value interface{}) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	val, wasInCache := c.items[key]
	if wasInCache {
		val.Value = cacheItem{key, value}
		c.queue.MoveToFront(val)
	} else {
		if c.queue.Len() >= c.capacity {
			c.removeLatest()
		}
		c.items[key] = c.queue.PushFront(cacheItem{key, value})
	}
	return wasInCache
}
func (c *lruCache) removeLatest() {
	latest := c.queue.Back()
	if latest != nil {
		if ci, ok := latest.Value.(cacheItem); ok {
			delete(c.items, ci.key)
			c.queue.Remove(latest)
		}
	}
}

func (c *lruCache) Get(key Key) (interface{}, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	val, ok := c.items[key]
	if ok {
		c.queue.MoveToFront(val)
		if ci, ok := val.Value.(cacheItem); ok {
			return ci.value, true
		}
	}
	return nil, false
}

func (c *lruCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for k := range c.items {
		delete(c.items, k)
	}
	c.queue = NewList()
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}
