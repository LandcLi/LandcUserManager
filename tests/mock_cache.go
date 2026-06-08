package tests

import (
	"context"
	"sync"
	"time"
)

// MockCache is a simple in-memory cache for testing
type MockCache struct {
	mu   sync.Mutex
	data map[string]mockCacheEntry
}

type mockCacheEntry struct {
	value    string
	expireAt time.Time
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]mockCacheEntry),
	}
}

func (c *MockCache) Set(ctx context.Context, key, value string, expire time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = mockCacheEntry{
		value:    value,
		expireAt: time.Now().Add(expire),
	}
	return nil
}

func (c *MockCache) Get(ctx context.Context, key string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.data[key]
	if !ok {
		return "", nil
	}
	if time.Now().After(entry.expireAt) {
		delete(c.data, key)
		return "", nil
	}
	return entry.value, nil
}

func (c *MockCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
	return nil
}

func (c *MockCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.data[key]
	if !ok {
		return false, nil
	}
	if time.Now().After(entry.expireAt) {
		delete(c.data, key)
		return false, nil
	}
	return true, nil
}

func (c *MockCache) Close() error {
	return nil
}
