package ecs

import (
	"reflect"
	"sync"
	"sync/atomic"
)

var (
	typeIDCounter uint64
	typeIDCache   = make(map[reflect.Type]uint64)
	cacheMu       sync.RWMutex
)

func TypeID[T any](v T) uint64 {
	typ := reflect.TypeOf(v)
	cacheMu.RLock()
	if id, ok := typeIDCache[typ]; ok {
		cacheMu.RUnlock()
		return id
	}
	cacheMu.RUnlock()

	cacheMu.Lock()
	if id, ok := typeIDCache[typ]; ok {
		cacheMu.Unlock()
		return id
	}
	id := atomic.AddUint64(&typeIDCounter, 1)
	typeIDCache[typ] = id
	cacheMu.Unlock()
	return id
}
