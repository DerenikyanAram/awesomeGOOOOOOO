package main

import "sync"

type Cache struct {
	mu    sync.RWMutex
	data  map[string]Order
	order []string
	max   int
}

func NewCache(max int) *Cache {
	return &Cache{data: make(map[string]Order), max: max}
}

func (c *Cache) Get(id string) (Order, bool) {
	c.mu.RLock()
	o, ok := c.data[id]
	c.mu.RUnlock()
	return o, ok
}

func (c *Cache) Set(o Order) {
	c.mu.Lock()
	if c.max > 0 && len(c.data) >= c.max && len(c.order) > 0 {
		old := c.order[0]
		c.order = c.order[1:]
		delete(c.data, old)
	}
	if _, exists := c.data[o.OrderUID]; !exists {
		c.order = append(c.order, o.OrderUID)
	}
	c.data[o.OrderUID] = o
	c.mu.Unlock()
}
