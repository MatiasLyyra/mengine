// Package ecs provides ECS (Entity Component System) implementation.
//
// The package provides effiecient way for systems to subscribe to relevant component combinations
// for entities and run update loops on only those entities.
//
// # Typesafety
//
// ECS package provides typesafe GetComponent, AddComponent and RemoveComponent functions with
// golang generics. These should work with any user provided types as long as they are first registered.
//
// See GetComponent, AddComponent and RemoveComponent examples for more information.
//
// # Changes during updates
//
// Additions and removals of components and entities are done lazily. This also provides ecs system
// more performant way to handle deletions and do cleanup on larger batches.
//
// First one is to provide consistent view of entities' components during update cycle of multiple systems.
// All RemoveComponent, AddComponent and RemoveEntity calls will have their effect visible only on the next
// update cycle. Changes to the component values themselves are always immediately visble to the other systems.
//
// Please see RegisterSystem (UpdateBarrier) example.
package ecs

import "fmt"

type World struct {
	em   *entityManager
	cm   *componentManager
	sm   *systemManager
	ism  *initSystemManger
	sing *singletonManager

	oplog []oplogEntry
	frame uint64
}

type oplogKind int

const (
	Delete oplogKind = iota
	Add
)

type oplogEntry struct {
	Kind   oplogKind
	Entity EntityID
	Sig    Signature
}

func New() *World {
	return &World{
		em:    &entityManager{},
		cm:    newComponentManager(),
		sm:    newSystemManager(),
		ism:   newInitSystemManager(),
		sing:  newSingletonManager(),
		frame: 0,
	}
}

// Init prepares ecs world and runs all of the init systems.
//
// This should be called after all the initial components have been created,
// and before the first RunUpdate call.
func (w *World) Init() {
	w.cleanup()
	w.frame++
	for _, init := range w.ism.systems {
		init.Init(w)
	}
}

func (w *World) cleanup() {
	for _, op := range w.oplog {
		switch op.Kind {
		case Add:
			w.sm.addEntity(op.Entity, op.Sig)
		case Delete:
			w.sm.removeEntity(op.Entity, op.Sig)
		}
	}
	w.oplog = w.oplog[:0]
	w.cm.prune()
}

// RunUpdate runs all of the Update systems and flushes all component additions or removals.
func (w *World) RunUpdate(dt float32) {
	us := UpdateState{
		World:     w,
		DeltaTime: dt,
	}
	for i, s := range w.sm.systems {
		entities := w.sm.systemEntities[i]
		us.Entities = entities
		s.Update(us)
	}
	w.cleanup()
	w.frame++
}

// NewEntity returns new EntityID handle
func NewEntity(w *World) EntityID {
	return w.em.newEntity()
}

// RegisterComponent registers new component of type T.
//
// This needs to be called before component can be used.
func RegisterComponent[T any](w *World) {
	registerComponent[T](w.cm)
}

// AddComponent adds component of type T to the Entity.
//
// Calling this on non-existent entity or on entity that already
// has the same component will panic.
func AddComponent[T any](w *World, e EntityID, c T) {
	addComponent(w.frame, w.cm, e, c)
	sig := w.cm.entitySignatures[uint64(e)]
	w.oplog = append(w.oplog, oplogEntry{Kind: Add, Entity: e, Sig: sig})
}

// RemoveComponent removes component of type T from the Entity.
//
// Calling this on non-existent entity or on entity that does not
// have the component will panic.
func RemoveComponent[T any](w *World, e EntityID) {
	sig := w.cm.entitySignatures[uint64(e)]
	removeComponent[T](w.frame, w.cm, e)
	w.oplog = append(w.oplog, oplogEntry{Kind: Delete, Entity: e, Sig: sig})
}

// GetComponent returns component of type T attached to the entity.
//
// If the entity does not have the component, GetComponent will return nil.
func GetComponent[T any](w *World, e EntityID) *T {
	return getComponent[T](w.frame, w.cm, e)
}

// MustGetComponent will return non-nil component of type T attached to the entity.
//
// Call to this will panic, if the the component does not exist on the entity.
func MustGetComponent[T any](w *World, e EntityID) *T {
	c := getComponent[T](w.frame, w.cm, e)
	if c == nil {
		var t T
		panic(fmt.Sprintf("entity %d does not have component of type %T", e, t))
	}
	return c
}

// RemoveEntity removes the entity and all of its associated components.
func RemoveEntity(w *World, e EntityID) {
	sig := w.cm.entitySignatures[uint64(e)]
	w.cm.removeEntity(w.frame, e)
	w.oplog = append(w.oplog, oplogEntry{Kind: Delete, Entity: e, Sig: sig})
}

// DebugComponent returns debug print of component T on entity.
func DebugComponent[T any](w *World, e EntityID) string {
	return debugPrintComponent[T](w.frame, w.cm, e)
}

// RegisterSystem registers new Update system the ecs world.
func RegisterSystem[T SystemType](w *World, s T, sig Signature) {
	registerSystem(w.sm, s, sig)
	idx := getSystemIdx[T](w.sm)
	for e := range w.cm.iterateEntities(w.frame, sig) {
		w.sm.systemEntities[idx] = append(w.sm.systemEntities[idx], e)
	}
}

func HasComponent[T any](w *World, e EntityID) bool {
	return hasComponent[T](w.frame, w.cm, e)
}

// RegisterInitSystem register new InitSystem to the ecs world.
func RegisterInitSystem(w *World, s InitSystem) {
	w.ism.registerSystem(s)
}

// RegisterSingleton registers singleton value to the ecs world.
//
// Calling this multiple times with same T will overwrite the previous value.
func RegisterSingleton[T any](w *World, singleton *T) {
	registerSingleton(w.sing, singleton)
}

// GetSingleton returns previously registered singleton value.
//
// Calling this on type T that has not been registered, will panic.
func GetSingleton[T any](w *World) *T {
	return getSingleton[T](w.sing)
}

// Sig will return Signature associated with component T.
func Sig[T any](w *World) Signature {
	return getComponentSignature[T](w.cm)
}
