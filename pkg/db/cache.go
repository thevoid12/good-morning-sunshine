// This is a cache layer which has all the caching code
// today I am using sync map and In future will be replaced by redis or anyother alternative based on use case
package dbpkg

import (
	"context"
	constants "gms/constant"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Cache struct holds the sync.Map for concurrent access
type Cache struct {
	CacheStore sync.Map
}

// CacheEntry represents a single cache entry
type CacheEntry struct {
	RecordID      uuid.UUID
	EmailID       string
	RandomNumbers string
	ExpiryDate    time.Time
}

func NewCache() *Cache {
	return &Cache{}
}

// Set stores a key-value pair in the cache
func (c *Cache) Set(key string, value *CacheEntry) {
	val := c.Get(key)
	if val == nil {
		ce := []*CacheEntry{}
		ce = append(ce, value)
		c.CacheStore.Store(key, ce)
	} else {
		val = append(val, value)
		c.CacheStore.Store(key, val)
	}
}

// Get retrieves a value from the cache if it hasn't expired
func (c *Cache) Get(key string) []*CacheEntry {
	value, ok := c.CacheStore.Load(key)
	if !ok {
		return nil
	}

	entry := value.([]*CacheEntry) //type assersion
	return entry
}

// Update complely replaces existing key If if exists
func (c *Cache) Update(key string, value []*CacheEntry) {
	val := c.Get(key)
	if val == nil {
		return
	} else {
		c.CacheStore.Store(key, value)
	}
}

// Delete removes a key-value pair from the cache
func (c *Cache) Delete(key string) {
	c.CacheStore.Delete(key)
}

// set cache struct instance in context
func SetCachectx(ctx context.Context, c *Cache) context.Context {
	return context.WithValue(ctx, constants.CONTEXT_KEY_CACHE, c)
}

// get the cache object from context
func GetCacheFromctx(ctx context.Context) (c *Cache) {
	val := ctx.Value(constants.CONTEXT_KEY_CACHE)
	if val == nil {
		return nil
	}
	c, ok := val.(*Cache)
	if !ok {
		return nil
	}
	return c
}
