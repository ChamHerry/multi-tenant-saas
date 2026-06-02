package redis

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
)

// CacheManager manages Redis adapter instances as a global singleton.
type CacheManager struct {
	mu       sync.RWMutex
	adapters map[string]*RedisAdapter
}

var (
	instance *CacheManager
	once     sync.Once
)

// GetCacheManager returns the global CacheManager singleton.
func GetCacheManager() *CacheManager {
	once.Do(func() {
		instance = &CacheManager{
			adapters: make(map[string]*RedisAdapter),
		}
	})
	return instance
}

// GetAdapter returns a previously initialized adapter by name.
// Returns nil if not initialized.
func (m *CacheManager) GetAdapter(name string) *RedisAdapter {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.adapters[name]
}

// InitAdapter creates and registers a Redis adapter.
// Safe to call multiple times — returns early if already initialized.
func (m *CacheManager) InitAdapter(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.adapters[name]; ok {
		return nil // already initialized
	}

	adapter, err := NewRedisAdapter(ctx, name)
	if err != nil {
		return err
	}

	m.adapters[name] = adapter
	g.Log().Infof(ctx, "[Redis] adapter[%s] initialized successfully", name)
	return nil
}

// Shutdown gracefully closes all adapters.
func (m *CacheManager) Shutdown(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, adapter := range m.adapters {
		if err := adapter.Close(); err != nil {
			g.Log().Warningf(ctx, "[Redis] adapter[%s] close error: %v", name, err)
		} else {
			g.Log().Infof(ctx, "[Redis] adapter[%s] closed", name)
		}
	}
	m.adapters = make(map[string]*RedisAdapter)
}
