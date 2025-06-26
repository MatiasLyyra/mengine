package ecs_test

import (
	"fmt"

	"github.com/MatiasLyyra/mengine/ecs"
)

type Player struct {
	X, Y int
}

type Inventory struct {
	Food int
}

func ExampleAddComponent() {
	w := ecs.New()
	ecs.RegisterComponent[Player](w)
	player := ecs.NewEntity(w)
	ecs.AddComponent(w, player, Player{X: 200, Y: 400})
	// Note: Init must be called before added components can be queries
	w.Init()
	fmt.Printf("Player: %v\n", ecs.DebugComponent[Player](w, player))
	// Output:
	// Player: {X:200 Y:400}
}

func ExampleRemoveComponent() {
	w := ecs.New()
	ecs.RegisterComponent[Player](w)
	ecs.RegisterComponent[Inventory](w)
	player := ecs.NewEntity(w)
	ecs.AddComponent(w, player, Player{X: 200, Y: 400})
	ecs.AddComponent(w, player, Inventory{Food: 2})
	w.Init()
	fmt.Printf("Player: %v\n", ecs.DebugComponent[Player](w, player))
	fmt.Printf("Inventory: %v\n", ecs.DebugComponent[Inventory](w, player))
	ecs.RemoveComponent[Inventory](w, player)
	fmt.Printf("Inventory: %v\n", ecs.DebugComponent[Inventory](w, player))
	// Output:
	// Player: {X:200 Y:400}
	// Inventory: {Food:2}
	// Inventory: <nil>
}

func ExampleGetComponent() {
	w := ecs.New()
	ecs.RegisterComponent[Player](w)
	ecs.RegisterComponent[Inventory](w)
	player := ecs.NewEntity(w)
	ecs.AddComponent(w, player, Player{X: 200, Y: 400})
	w.Init()

	playerComponent := ecs.GetComponent[Player](w, player)
	playerComponent.X += 20
	playerComponent.Y /= 2
	fmt.Printf("Player: %v\n", ecs.DebugComponent[Player](w, player))
	// Output:
	// Player: {X:220 Y:200}
}

type PlayerUpdateSystem struct{}

func (PlayerUpdateSystem) Update(us ecs.UpdateState) {
	for _, e := range us.Entities {
		player := ecs.GetComponent[Player](us.World, e)
		inventory := ecs.GetComponent[Inventory](us.World, e)
		fmt.Printf("Player: %+v\n", player)
		fmt.Printf("Inventory: %+v\n", inventory)
	}
}

func ExampleRegisterSystem() {
	w := ecs.New()
	ecs.RegisterComponent[Player](w)
	ecs.RegisterComponent[Inventory](w)
	// player1 entity will not get printed, as it lacks the necessary Inventory component for the system
	player1 := ecs.NewEntity(w)
	ecs.AddComponent(w, player1, Player{X: 200, Y: 400})
	player2 := ecs.NewEntity(w)
	ecs.AddComponent(w, player2, Player{X: 300, Y: 500})
	ecs.AddComponent(w, player2, Inventory{Food: 6})
	ecs.RegisterSystem(w, PlayerUpdateSystem{}, ecs.Sig[Player](w)|ecs.Sig[Inventory](w))
	w.Init()
	w.RunUpdate(0)
	// Output:
	// Player: &{X:300 Y:500}
	// Inventory: &{Food:6}
}

type ModifyPlayerSystem struct{}

func (ModifyPlayerSystem) Update(us ecs.UpdateState) {
	for _, e := range us.Entities {
		if !ecs.HasComponent[Inventory](us.World, e) {
			ecs.AddComponent(us.World, e, Inventory{Food: 10})
		}
		fmt.Printf("ModifyPlayerSystem.Update Player: %+v\n", ecs.GetComponent[Player](us.World, e))
		fmt.Printf("ModifyPlayerSystem.Update Inventory: %+v\n", ecs.GetComponent[Inventory](us.World, e))
	}
}

type PrintPlayerSystem struct{}

func (PrintPlayerSystem) Update(us ecs.UpdateState) {
	for _, e := range us.Entities {
		fmt.Printf("PrintPlayerSystem.Update Player: %+v\n", ecs.GetComponent[Player](us.World, e))
		fmt.Printf("PrintPlayerSystem.Update Inventory: %+v\n", ecs.GetComponent[Inventory](us.World, e))
	}
}

// This demonstrates more in-depth interactions with multiple
// systems that add and delete components during an update.
//
// In the example, PrintPlayerSystem will only trigger on the second update cycle.
// This is because component changes are only applied on the next update cycle.
func ExampleRegisterSystem_updateBarrier() {
	w := ecs.New()
	ecs.RegisterComponent[Player](w)
	ecs.RegisterComponent[Inventory](w)
	player1 := ecs.NewEntity(w)
	ecs.AddComponent(w, player1, Player{X: 200, Y: 400})
	ecs.RegisterSystem(w, ModifyPlayerSystem{}, ecs.Sig[Player](w))
	ecs.RegisterSystem(w, PrintPlayerSystem{}, ecs.Sig[Player](w)|ecs.Sig[Inventory](w))
	w.Init()
	// First update cycle
	w.RunUpdate(0)
	// Second update cycle
	w.RunUpdate(0)
	// Output:
	// ModifyPlayerSystem.Update Player: &{X:200 Y:400}
	// ModifyPlayerSystem.Update Inventory: <nil>
	// ModifyPlayerSystem.Update Player: &{X:200 Y:400}
	// ModifyPlayerSystem.Update Inventory: &{Food:10}
	// PrintPlayerSystem.Update Player: &{X:200 Y:400}
	// PrintPlayerSystem.Update Inventory: &{Food:10}
}
