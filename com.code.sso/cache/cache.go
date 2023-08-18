package cache

import (
	"sync"
)

type ExpiringCache struct {
	syncMap  sync.Map
	order    []interface{}
	capacity int
	mu       sync.Mutex
}

var (
	expiringCacheInstance *ExpiringCache
	once                  sync.Once
)

func (ec *ExpiringCache) Store(key interface{}, value interface{}) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if len(ec.order) >= ec.capacity {
		oldestKey := ec.order[0]
		ec.order = ec.order[1:]
		ec.syncMap.Delete(oldestKey)
	}

	ec.syncMap.Store(key, value)
	ec.order = append(ec.order, key)
}

func (ec *ExpiringCache) Load(key interface{}) (interface{}, bool) {
	value, found := ec.syncMap.Load(key)
	return value, found
}

func GetInstance() *ExpiringCache {
	once.Do(func() {
		expiringCacheInstance = &ExpiringCache{}
	})
	return expiringCacheInstance
}
