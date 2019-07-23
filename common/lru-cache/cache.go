package cache

import (
	lru "github.com/hashicorp/golang-lru"
)

type Cache struct {
	*lru.Cache
}

func (c *Cache) Add(key interface{}, payload interface{}) {
	_ = c.Cache.Add(key, payload)
}

func (c *Cache) Get(key interface{}) interface{} {
	item, ok := c.Cache.Get(key)
	if !ok {
		return nil
	}
	return item
}

func New(size int, onEvicted func(key interface{}, value interface{})) *Cache {
	c, _ := lru.NewWithEvict(size, onEvicted)
	return &Cache{c}
}
