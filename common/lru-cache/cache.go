package cache

import (
	lru "github.com/hashicorp/golang-lru"
)

type Cache struct {
	*lru.Cache
}

func (c *Cache) Purge() {
	c.Cache.Purge()
}

func (c *Cache) Put(key interface{}, payload interface{}) {
	_ = c.Cache.Add(key, payload)
}

func (c *Cache) Get(key interface{}) interface{} {
	item, ok := c.Cache.Get(key)
	if !ok {
		return nil
	}
	return item
}

func New(size int) *Cache {
	c, _ := lru.New(size)
	return &Cache{c}
}
