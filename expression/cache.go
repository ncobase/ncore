package expression

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// CacheStats tracks cache statistics
type CacheStats struct {
	Hits      int64 // Number of cache hits
	Misses    int64 // Number of cache misses
	Evictions int64 // Number of cache evictions
	Size      int64 // Current size in bytes
}

// Cache implements a thread-safe LRU cache with memory limits and TTL support
type Cache struct {
	items     map[string]*list.Element // Map for O(1) lookup
	evictList *list.List               // Doubly linked list for LRU
	stats     CacheStats               // Cache statistics
	config    *CacheConfig             // Cache configuration
	mu        sync.RWMutex             // Read-write mutex for thread safety
}

// CacheConfig defines configuration options for the cache
type CacheConfig struct {
	MaxSize         int64                       // Maximum memory size in bytes
	TTL             time.Duration               // Time to live for cache entries
	CleanupInterval time.Duration               // Interval for cleanup routine
	OnEvict         func(key string, value any) // Callback when an item is evicted
}

// cacheEntry represents a single cache entry
type cacheEntry struct {
	key       string    // Cache key
	value     any       // Stored value
	size      int64     // Size in bytes
	timestamp time.Time // Last access time
	expiry    time.Time // Expiration time (zero means no expiry)
}

// NewCache creates a new cache instance with the given configuration
func NewCache(config *CacheConfig) *Cache {
	if config == nil {
		config = &CacheConfig{
			MaxSize:         1024 * 1024 * 10, // 10MB default
			TTL:             time.Hour,
			CleanupInterval: time.Minute * 5,
		}
	}

	c := &Cache{
		items:     make(map[string]*list.Element),
		evictList: list.New(),
		config:    config,
	}

	// Start cleanup routine if interval is set
	if config.CleanupInterval > 0 {
		go c.startCleanup(context.Background())
	}

	return c
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if ent, ok := c.items[key]; ok {
		entry := ent.Value.(*cacheEntry)

		// Check if expired
		if !entry.expiry.IsZero() && time.Now().After(entry.expiry) {
			c.mu.RUnlock()
			c.mu.Lock()
			c.removeElement(ent)
			c.mu.Unlock()
			c.mu.RLock()
			atomic.AddInt64(&c.stats.Misses, 1)
			return nil, false
		}

		// Update access time and move to front
		entry.timestamp = time.Now()
		c.evictList.MoveToFront(ent)
		atomic.AddInt64(&c.stats.Hits, 1)
		return entry.value, true
	}

	atomic.AddInt64(&c.stats.Misses, 1)
	return nil, false
}

// Set adds or updates a value in the cache
func (c *Cache) Set(key string, value any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	size := estimateSize(value)
	if size > c.config.MaxSize {
		return fmt.Errorf("item size %d exceeds cache max size %d", size, c.config.MaxSize)
	}

	// If key exists, remove it first
	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
	}

	// Ensure enough space
	for c.stats.Size+size > c.config.MaxSize {
		if !c.evictOldest() {
			return fmt.Errorf("cannot make room for item with size %d", size)
		}
	}

	// Create new entry
	entry := &cacheEntry{
		key:       key,
		value:     value,
		size:      size,
		timestamp: time.Now(),
	}

	if c.config.TTL > 0 {
		entry.expiry = entry.timestamp.Add(c.config.TTL)
	}

	// Add to LRU list and map
	element := c.evictList.PushFront(entry)
	c.items[key] = element
	atomic.AddInt64(&c.stats.Size, entry.size)

	return nil
}

// Remove removes a key from the cache
func (c *Cache) Remove(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
		return true
	}
	return false
}

// removeElement removes an element from the cache
func (c *Cache) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	entry := e.Value.(*cacheEntry)
	delete(c.items, entry.key)
	atomic.AddInt64(&c.stats.Size, -entry.size)
	atomic.AddInt64(&c.stats.Evictions, 1)

	if c.config.OnEvict != nil {
		c.config.OnEvict(entry.key, entry.value)
	}
}

// evictOldest removes the oldest item from the cache
func (c *Cache) evictOldest() bool {
	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent)
		return true
	}
	return false
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, ent := range c.items {
		entry := ent.Value.(*cacheEntry)
		if c.config.OnEvict != nil {
			c.config.OnEvict(entry.key, entry.value)
		}
	}

	c.items = make(map[string]*list.Element)
	c.evictList.Init()
	c.stats = CacheStats{}
}

// Size returns the current size of the cache in bytes
func (c *Cache) Size() int64 {
	return atomic.LoadInt64(&c.stats.Size)
}

// Len returns the number of items in the cache
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Stats returns cache statistics
func (c *Cache) Stats() CacheStats {
	return CacheStats{
		Hits:      atomic.LoadInt64(&c.stats.Hits),
		Misses:    atomic.LoadInt64(&c.stats.Misses),
		Evictions: atomic.LoadInt64(&c.stats.Evictions),
		Size:      atomic.LoadInt64(&c.stats.Size),
	}
}

// startCleanup starts the background cleanup routine
func (c *Cache) startCleanup(ctx context.Context) {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.cleanupExpired()
		}
	}
}

// cleanupExpired removes expired entries from the cache
func (c *Cache) cleanupExpired() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, ent := range c.items {
		entry := ent.Value.(*cacheEntry)
		if !entry.expiry.IsZero() && now.After(entry.expiry) {
			c.removeElement(ent)
		}
	}
}

// PurgeExpired manually removes all expired items from the cache
func (c *Cache) PurgeExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	now := time.Now()

	for _, ent := range c.items {
		entry := ent.Value.(*cacheEntry)
		if !entry.expiry.IsZero() && now.After(entry.expiry) {
			c.removeElement(ent)
			count++
		}
	}

	return count
}
