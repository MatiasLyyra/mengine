package ecs

import (
	"fmt"
	"iter"
)

type componentArrayType interface {
	iter(uint64) iter.Seq[EntityID]
	debug(uint64, EntityID) string
	has(uint64, EntityID) bool
	prune()
	remove(uint64, EntityID)
}

type componentArray[T any] struct {
	entities    []entityComponent[T]
	entityToIdx map[EntityID]int
	idx         int
	deads       int
}

func (ca *componentArray[T]) iter(frame uint64) iter.Seq[EntityID] {
	return func(yield func(EntityID) bool) {
		for _, e := range ca.entities {
			if e.isDead(frame) {
				continue
			}
			if !yield(e.Id) {
				break
			}
		}
	}
}

func (ca *componentArray[T]) debug(frame uint64, e EntityID) string {
	idx, ok := ca.entityToIdx[e]
	if !ok || ca.entities[idx].isDead(frame) {
		return "<nil>"
	}
	return fmt.Sprintf("%+v", ca.entities[idx].Component)
}

func newComponentArray[T any]() *componentArray[T] {
	return &componentArray[T]{
		entities:    make([]entityComponent[T], initialEntityArraySize),
		entityToIdx: make(map[EntityID]int, initialEntityArraySize),
	}
}

func (ca *componentArray[T]) add(frame uint64, e EntityID, c T) {
	if idx, ok := ca.entityToIdx[e]; ok && ca.entities[idx].Alive {
		panic(fmt.Sprintf("entity %d already contains component %T", e, c))
	}
	if length := len(ca.entities); ca.idx >= length {
		newEntities := make([]entityComponent[T], 2*length)
		copy(newEntities[:length], ca.entities)
		ca.entities = newEntities
	}
	ca.entities[ca.idx] = entityComponent[T]{
		Id:        e,
		Component: c,
		Alive:     true,
		ChangedAt: frame,
	}
	ca.entityToIdx[e] = ca.idx
	ca.idx++
}

func (ca *componentArray[T]) remove(frame uint64, e EntityID) {
	eIdx, ok := ca.entityToIdx[e]
	if !ok || !ca.entities[eIdx].Alive {
		var t T
		panic(fmt.Sprintf("entity %d does not have component %T", e, t))
	}
	ca.entities[eIdx].Alive = false
	ca.entities[eIdx].ChangedAt = frame
	ca.deads++
}

func (ca *componentArray[T]) prune() {
	if ca.deads >= len(ca.entities)/2 && ca.idx > initialEntityArraySize {
		var i = 0
		var j = len(ca.entities) - 1
		clear(ca.entityToIdx)
		for {
			for idx := j; i < j; idx-- {
				if ca.entities[idx].Alive {
					j = idx
					break
				}
			}
			for idx := i; i < j; idx++ {
				if !ca.entities[idx].Alive {
					i = idx
					break
				}
			}
			if i >= j {
				break
			}
			ca.entities[i], ca.entities[j] = ca.entities[j], ca.entities[i]
		}
		ca.idx = i + 1
		ca.deads = 0
		for idx := 0; idx < ca.idx; idx++ {
			ca.entityToIdx[ca.entities[idx].Id] = idx
		}
	}
}

func (ca *componentArray[T]) has(frame uint64, e EntityID) bool {
	idx, ok := ca.entityToIdx[e]
	return ok && ca.entities[idx].isAlive(frame)
}

func (ca *componentArray[T]) Get(frame uint64, e EntityID) *T {
	idx, ok := ca.entityToIdx[e]
	if !ok || ca.entities[idx].isDead(frame) {
		return nil
	}
	return &ca.entities[idx].Component
}

type entityComponent[T any] struct {
	Id        EntityID
	Component T
	Alive     bool
	ChangedAt uint64
}

func (e entityComponent[T]) isAlive(frame uint64) bool {
	return e.Alive || e.ChangedAt == frame
}
func (e entityComponent[T]) isDead(frame uint64) bool {
	return !e.Alive || e.ChangedAt == frame
}
