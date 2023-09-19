package uof

import (
	"context"
	"sync"
	"time"
)

type MapCache[K string | int, V any] struct {
	sync.RWMutex
	ctx                context.Context
	data               map[K]*V
	expiration         map[K]time.Time
	timeToLive         time.Duration
	cacheCleanInterval time.Duration
}

func NewMapCache[K string | int, V any](ctx context.Context, timeToLive time.Duration, cacheCleanInterval time.Duration) *MapCache[K, V] {
	cache := &MapCache[K, V]{
		ctx:                ctx,
		data:               make(map[K]*V),
		expiration:         make(map[K]time.Time),
		timeToLive:         timeToLive,
		cacheCleanInterval: cacheCleanInterval,
	}

	go cache.cleanUpExpiredEntries()

	return cache
}

func (m *MapCache[K, V]) Set(key K, value V) {
	m.Lock()
	defer m.Unlock()
	m.data[key] = &value
	m.expiration[key] = time.Now().Add(m.timeToLive)
}

func (m *MapCache[K, V]) Get(key K) (*V, bool) {
	m.RLock()
	defer m.RUnlock()
	val, ok := m.data[key]
	return val, ok
}

func (m *MapCache[K, V]) Size() int {
	return len(m.data)
}

func (m *MapCache[K, V]) cleanUpExpiredEntries() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-time.After(m.cacheCleanInterval):
			m.Lock()
			currentTime := time.Now()
			for key, expirationTime := range m.expiration {
				if currentTime.After(expirationTime) {
					delete(m.data, key)
					delete(m.expiration, key)
				}
			}
			m.Unlock()
		}
	}
}
