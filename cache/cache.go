package cache

import (
	"context"
	"log"
	"time"

	"github.com/karlseguin/ccache"
)

// MemoryCache is a CCache implementation.
type MemoryCache struct {
	name          string
	cCache        *ccache.Cache
	defaultTTL    time.Duration
	returnExpired bool
}

// NewMemoryCache creates a new MemoryCache instance.
// name: name of the cache
// size: max number of items in the cache
// ttl: default ttl for items in the cache (in seconds)
// returnExpired: if true, expired items will be returned. if false nil will be returned
func NewMemoryCache(name string, size int, ttl time.Duration, returnExpired bool) Spec {
	return &MemoryCache{
		name:          name,
		defaultTTL:    ttl,
		cCache:        ccache.New(ccache.Configure().MaxSize(int64(size)).ItemsToPrune(100)),
		returnExpired: returnExpired,
	}
}

func (impl *MemoryCache) Get(ctx context.Context, key string) (expired bool, value interface{}) {
	if ctx != nil {
		log.Printf("cache.%s.Get: key=%s", impl.name, key)
	}
	item := impl.cCache.Get(key)
	if item != nil && !item.Expired() {
		res := item.Value()
		return false, res
	}
	if impl.returnExpired && item != nil {
		res := item.Value()
		return item.Expired(), res
	}

	return false, nil
}

func (impl *MemoryCache) Save(ctx context.Context, key string, item interface{}) {
	if ctx != nil {
		log.Printf("cache.%s.Save: key=%s", impl.name, key)
	}
	impl.cCache.Set(key, item, impl.defaultTTL)
}

func (impl *MemoryCache) SaveWithTTL(ctx context.Context, key string, item interface{}, ttl time.Duration) {
	if ctx != nil {
		log.Printf("cache.%s.SaveWithTTL: key=%s, ttl=%v", impl.name, key, ttl)
	}
	impl.cCache.Set(key, item, ttl)
}

// Delete key value pair from the cache
func (impl *MemoryCache) Delete(ctx context.Context, key string) {
	if ctx != nil {
		log.Printf("cache.%s.Delete: key=%s", impl.name, key)
	}
	impl.cCache.Delete(key)
}
