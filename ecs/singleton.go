package ecs

import "fmt"

type singleton[T any] struct {
	value *T
}

type singletonType interface {
}

type singletonManager struct {
	values map[uint64]any
}

func newSingletonManager() *singletonManager {
	return &singletonManager{
		values: make(map[uint64]any),
	}
}

func registerSingleton[T any](sm *singletonManager, s *T) {
	var t T
	name := TypeID(t)
	sm.values[name] = singleton[T]{value: s}
}

func getSingleton[T any](sm *singletonManager) *T {
	var t T
	name := TypeID(t)
	v, ok := sm.values[name].(singleton[T])
	if !ok {
		panic(fmt.Sprintf("singleton of type %T has not been registed", t))
	}
	return v.value
}
