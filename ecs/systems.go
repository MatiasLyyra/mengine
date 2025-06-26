package ecs

import (
	"fmt"
)

type SystemType interface {
	Update(UpdateState)
}

type UpdateState struct {
	World     *World
	Entities  []EntityID
	DeltaTime float32
}

type systemManager struct {
	systems          []SystemType
	systemEntities   [][]EntityID
	systemEntitySet  []map[uint64]int
	systemSignatures []Signature
	systemIdx        map[uint64]int
	// idx              int
}

func newSystemManager() *systemManager {
	return &systemManager{
		systemIdx: make(map[uint64]int),
	}
}

func getSystemIdx[T SystemType](sm *systemManager) int {
	var s T
	typeName := TypeID(s)
	idx, ok := sm.systemIdx[typeName]
	if !ok {
		panic(fmt.Sprintf("component %v is not registered with ComponentManager", typeName))
	}
	return idx
}

func registerSystem[T SystemType](sm *systemManager, s T, sig Signature) {
	name := TypeID(s)
	if _, ok := sm.systemIdx[name]; ok {
		panic(fmt.Sprintf("system %T is already registered to SystemManager", s))
	}
	idx := len(sm.systems)
	sm.systems = append(sm.systems, s)
	sm.systemEntities = append(sm.systemEntities, make([]EntityID, 0, initialEntityArraySize))
	sm.systemSignatures = append(sm.systemSignatures, sig)
	sm.systemEntitySet = append(sm.systemEntitySet, make(map[uint64]int))
	sm.systemIdx[name] = idx
	// sm.idx++
}

func (sm *systemManager) addEntity(e EntityID, sig Signature) {
	for i := range sm.systems {
		// Check if the system is setup to listen for these entities
		if sm.systemSignatures[i]&sig != sm.systemSignatures[i] {
			continue
		}
		if _, ok := sm.systemEntitySet[i][uint64(e)]; ok {
			continue
		}
		idx := len(sm.systemEntities[i])
		sm.systemEntities[i] = append(sm.systemEntities[i], e)
		sm.systemEntitySet[i][uint64(e)] = idx
	}
}

func (sm *systemManager) removeEntity(e EntityID, sig Signature) {
	for i := range sm.systems {
		// Check if the system is setup to listen for these entities
		if sm.systemSignatures[i]&sig != sm.systemSignatures[i] {
			continue
		}
		idx, ok := sm.systemEntitySet[i][uint64(e)]
		if !ok {
			continue
		}
		end := len(sm.systemEntities[i]) - 1
		lastEntity := sm.systemEntities[i][end]
		sm.systemEntitySet[i][uint64(lastEntity)] = idx
		sm.systemEntities[i] = sm.systemEntities[i][:end]
	}
}
