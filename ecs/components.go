package ecs

import (
	"fmt"
	"iter"
	"strings"
)

const initialEntityArraySize = 512

type Signature uint64
type EntityID uint64

func (s Signature) Next() (Signature, bool) {
	next := s << 1
	if next == 0 {
		return 0, false
	}
	return next, true
}

func newComponentManager() *componentManager {
	return &componentManager{
		componentArray:   make([]componentArrayType, 0, 64),
		componentToIdx:   make(map[uint64]int, 64),
		entitySignatures: make(map[uint64]Signature, initialEntityArraySize),
		deadEntities:     make(map[uint64]struct{}, initialEntityArraySize),
		sig:              1,
	}
}

type componentManager struct {
	componentArray      []componentArrayType
	componentSignatures []Signature
	componentToIdx      map[uint64]int
	entitySignatures    map[uint64]Signature
	deadEntities        map[uint64]struct{}
	sig                 Signature
}

func (cm *componentManager) iterateEntities(frame uint64, sig Signature) iter.Seq[EntityID] {
	return func(yield func(EntityID) bool) {
		for i := range cm.componentArray {
			if cm.componentSignatures[i]&sig == 0 {
				continue
			}
			for e := range cm.componentArray[i].iter(frame) {
				if !yield(e) {
					break
				}
			}
		}
	}
}

func (cm *componentManager) prune() {
	for dead := range cm.deadEntities {
		delete(cm.entitySignatures, dead)
	}
	clear(cm.deadEntities)
	for _, ca := range cm.componentArray {
		ca.prune()
	}
}

func (cm *componentManager) removeEntity(frame uint64, e EntityID) {
	for _, ca := range cm.componentArray {
		cm.deadEntities[uint64(e)] = struct{}{}
		ca.remove(frame, e)
	}
}

func hasComponentArray[T any](cm *componentManager) bool {
	var t T
	name := TypeID(t)
	_, ok := cm.componentToIdx[name]
	return ok
}

func getComponentIdx[T any](cm *componentManager) int {
	var t T
	name := TypeID(t)
	idx, ok := cm.componentToIdx[name]
	if !ok {
		panic(fmt.Sprintf("component %v  is not registered with ComponentManager", name))
	}
	return idx
}

func addComponent[T any](frame uint64, cm *componentManager, e EntityID, c T) {
	idx := getComponentIdx[T](cm)
	ca := cm.componentArray[idx].(*componentArray[T])
	sig := cm.entitySignatures[uint64(e)]
	sig |= cm.componentSignatures[idx]
	cm.entitySignatures[uint64(e)] = sig
	ca.add(frame, e, c)
}

func debugPrintComponent[T any](frame uint64, cm *componentManager, e EntityID) string {
	idx := getComponentIdx[T](cm)
	return cm.componentArray[idx].debug(frame, e)
}

func debugPrintEntity(frame uint64, cm *componentManager, e EntityID) string {
	var debugs []string
	for _, ca := range cm.componentArray {
		if ca.has(frame, e) {
			debugs = append(debugs, ca.debug(frame, e))
		}
	}
	return strings.Join(debugs, "\n")
}

func removeComponent[T any](frame uint64, cm *componentManager, e EntityID) {
	idx := getComponentIdx[T](cm)
	ca := cm.componentArray[idx].(*componentArray[T])
	sig := cm.entitySignatures[uint64(e)]
	sig &= ^cm.componentSignatures[idx]
	cm.entitySignatures[uint64(e)] = sig
	ca.remove(frame, e)
}

func registerComponent[T any](cm *componentManager) {
	var c T
	var (
		ok bool
	)
	if hasComponentArray[T](cm) {
		panic(fmt.Sprintf("component %T is already registered", c))
	}

	cm.componentArray = append(cm.componentArray, newComponentArray[T]())
	cm.componentSignatures = append(cm.componentSignatures, cm.sig)
	cm.componentToIdx[TypeID(c)] = len(cm.componentArray) - 1
	cm.sig, ok = cm.sig.Next()
	if !ok {
		panic("signature overflowed, too many registerd components")
	}
}

func getComponent[T any](frame uint64, cm *componentManager, e EntityID) *T {
	idx := getComponentIdx[T](cm)
	ca := cm.componentArray[idx].(*componentArray[T])
	return ca.Get(frame, e)
}

func hasComponent[T any](frame uint64, cm *componentManager, e EntityID) bool {
	idx := getComponentIdx[T](cm)
	ca := cm.componentArray[idx].(*componentArray[T])
	return ca.has(frame, e)
}

func getComponentSignature[T any](cm *componentManager) Signature {
	idx := getComponentIdx[T](cm)
	return cm.componentSignatures[idx]
}

type entityManager struct {
	idx uint64
}

func (em *entityManager) newEntity() EntityID {
	e := em.idx
	em.idx++
	return EntityID(e)
}
