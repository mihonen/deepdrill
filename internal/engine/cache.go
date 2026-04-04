package engine

import (
	"sync"
)

type CacheItem struct {
	Hash       string
	Heuristics map[string]FieldHeuristic
}

type Cache struct {
	mu    sync.RWMutex
	items map[string]CacheItem
}

func NewCache() *Cache {
	return &Cache{
		items: make(map[string]CacheItem),
	}
}

func (c *Cache) Add(hash string, heuristics map[string]FieldHeuristic) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[hash] = CacheItem{
		Hash:       hash,
		Heuristics: heuristics,
	}
}

func (c *Cache) Get(hash string) (map[string]FieldHeuristic, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.items[hash]
	return item.Heuristics, ok
}

