package engine

import "sync"

// Heuristic is a DOM path to a field value, built from a prior extraction.
// Used by the future SQLite-backed caching layer: once a page skeleton is known,
// heuristics let us skip the LLM entirely on repeat runs.
type Heuristic struct {
	Path string `json:"path"`
}

type CacheItem struct {
	Hash       string
	Heuristics map[string]Heuristic
}

type Cache struct {
	mu    sync.RWMutex
	items map[string]CacheItem
}

func NewCache() *Cache {
	return &Cache{items: make(map[string]CacheItem)}
}

func (c *Cache) Add(hash string, heuristics map[string]Heuristic) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[hash] = CacheItem{Hash: hash, Heuristics: heuristics}
}

func (c *Cache) Get(hash string) (map[string]Heuristic, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.items[hash]
	return item.Heuristics, ok
}
