package main

import (
	"image/color"
	"math/rand"

	"github.com/MatiasLyyra/mengine/ecs"
	"github.com/MatiasLyyra/mengine/engine"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Transform struct {
	Position rl.Vector2
	Rotation float32
}

type PlayerGraphics struct {
	Radius          float32
	Color           color.RGBA
	OnCooldownColor color.RGBA
	OnDashColor     color.RGBA
}

type PlayerValues struct {
	DashCooldown float32
	DashSpeed    float32
	DashDuration float32
	BaseSpeed    float32
	Acceleration float32
}

type Player struct {
	Velocity rl.Vector2
	Speed    float32

	DashRemaining float32
	DashCooldown  float32
}

type DrawPlayerSystem struct{}

func (mbs DrawPlayerSystem) Update(us ecs.UpdateState) {
	for _, e := range us.Entities {
		ball := ecs.GetComponent[PlayerGraphics](us.World, e)
		transform := ecs.GetComponent[Transform](us.World, e)
		player := ecs.GetComponent[Player](us.World, e)
		playerColor := ball.Color
		if player.DashRemaining > 0 {
			playerColor = ball.OnDashColor
		} else if player.DashCooldown > 0 {
			playerColor = ball.OnCooldownColor
		}
		rl.DrawCircle(int32(transform.Position.X), int32(transform.Position.Y), ball.Radius, playerColor)
	}
}

type PlayerDashSystem struct{}

func (mbs PlayerDashSystem) Update(us ecs.UpdateState) {

	values := ecs.GetSingleton[PlayerValues](us.World)
	for _, e := range us.Entities {
		player := ecs.GetComponent[Player](us.World, e)

		if player.DashCooldown <= 0 && rl.IsKeyPressed(rl.KeySpace) {
			player.DashCooldown = values.DashCooldown
			player.Speed = values.BaseSpeed + values.DashSpeed
			player.DashRemaining = values.DashDuration
		}

		if player.DashCooldown > 0 {
			player.DashCooldown -= us.DeltaTime
		}
		if player.DashRemaining > 0 {
			player.DashRemaining -= us.DeltaTime
		} else {
			player.Speed = values.BaseSpeed
		}
	}
}

type MovePlayerSystem struct{}

func (mbs MovePlayerSystem) Update(us ecs.UpdateState) {
	values := ecs.GetSingleton[PlayerValues](us.World)
	for _, e := range us.Entities {
		var dir rl.Vector2
		transform := ecs.GetComponent[Transform](us.World, e)
		player := ecs.GetComponent[Player](us.World, e)

		up := rl.IsKeyDown(rl.KeyUp)
		down := rl.IsKeyDown(rl.KeyDown)
		left := rl.IsKeyDown(rl.KeyLeft)
		right := rl.IsKeyDown(rl.KeyRight)
		if (up || down) && !(up && down) {
			if up {
				dir.Y = -1
			} else if down {
				dir.Y = 1
			}
		}
		if (left || right) && !(left && right) {
			if left {
				dir.X = -1
			} else if right {
				dir.X = 1
			}
		}
		dir = rl.Vector2Normalize(dir)
		target := rl.Vector2Scale(dir, player.Speed)
		player.Velocity = rl.Vector2MoveTowards(player.Velocity, target, values.Acceleration*us.DeltaTime)

		if rl.Vector2LengthSqr(player.Velocity) == 0 {
			player.DashRemaining = 0
		}
		transform.Position = rl.Vector2Add(transform.Position, rl.Vector2Scale(player.Velocity, us.DeltaTime))
	}
}

type PlayerCollisionSystem struct{}

func (PlayerCollisionSystem) Update(us ecs.UpdateState) {
	w := ecs.GetSingleton[engine.Window](us.World)
	for _, e := range us.Entities {
		player := ecs.GetComponent[Player](us.World, e)
		transform := ecs.GetComponent[Transform](us.World, e)
		ball := ecs.GetComponent[PlayerGraphics](us.World, e)
		if transform.Position.X <= ball.Radius || transform.Position.X >= w.Width-ball.Radius {
			player.Velocity.X *= -0.4
			transform.Position.X = rl.Clamp(transform.Position.X, ball.Radius, w.Width-ball.Radius)
		}
		if transform.Position.Y <= ball.Radius || transform.Position.Y >= w.Height-ball.Radius {
			player.Velocity.Y *= -0.4
			transform.Position.Y = rl.Clamp(transform.Position.Y, ball.Radius, w.Height-ball.Radius)
		}
	}
}

func main() {
	e := engine.New()
	w := e.World

	ecs.RegisterSingleton(w, &PlayerValues{
		DashCooldown: 3,
		DashDuration: 0.8,
		DashSpeed:    1800,
		BaseSpeed:    600,
		Acceleration: 2000,
	})
	ecs.RegisterComponent[Transform](w)
	ecs.RegisterComponent[PlayerGraphics](w)
	ecs.RegisterComponent[Player](w)
	ecs.RegisterSystem(w, PlayerDashSystem{}, ecs.Sig[Player](w))
	ecs.RegisterSystem(w, MovePlayerSystem{}, ecs.Sig[Transform](w)|ecs.Sig[Player](w))
	ecs.RegisterSystem(w, PlayerCollisionSystem{}, ecs.Sig[Transform](w)|ecs.Sig[Player](w)|ecs.Sig[PlayerGraphics](w))
	ecs.RegisterSystem(w, DrawPlayerSystem{}, ecs.Sig[PlayerGraphics](w)|ecs.Sig[Transform](w)|ecs.Sig[Player](w))
	settings := ecs.GetSingleton[engine.WindowSettings](w)
	settings.Width = 1280
	settings.Height = 720
	settings.Title = "Bouncy Balls"
	settings.LogLevel = rl.LogError

	player := ecs.NewEntity(w)
	ecs.AddComponent(w, player, Player{})
	ecs.AddComponent(w, player, Transform{
		Position: rl.Vector2{X: 200, Y: 200},
	})
	ecs.AddComponent(w, player, PlayerGraphics{
		Radius:          20,
		Color:           rl.Green,
		OnDashColor:     rl.Red,
		OnCooldownColor: rl.DarkGreen,
	})
	e.Run()
}

// type PhysWorld struct {
// 	Objects  []rl.Vector2
// 	Velocity []rl.Vector2
// 	Radius   []int32
// }

// type Graphics struct {
// 	Colors []rl.Color
// }

// func initializeRandomPositions(w, h, count int32) *PhysWorld {
// 	world := PhysWorld{
// 		Objects:  make([]rl.Vector2, count),
// 		Radius:   make([]int32, count),
// 		Velocity: make([]rl.Vector2, count),
// 	}
// 	var i int32
// 	for i = 0; i < count; i++ {
// 		radius := rand.Int31n(25) + 10
// 		x := float32(rand.Int31n(w-2*radius) + radius)
// 		y := float32(rand.Int31n(h-2*radius) + radius)

// 		velocity := rl.Vector2Scale(
// 			rl.Vector2Normalize(
// 				rl.Vector2{
// 					X: rand.Float32()*2 - 1,
// 					Y: rand.Float32()*2 - 1,
// 				},
// 			),
// 			rand.Float32()*150+50,
// 		)
// 		world.Objects[i] = rl.Vector2{
// 			X: x,
// 			Y: y,
// 		}
// 		world.Radius[i] = radius
// 		world.Velocity[i] = velocity
// 	}
// 	return &world
// }

// func intializeRandomColors(count int32) *Graphics {
// 	graphics := Graphics{
// 		Colors: make([]rl.Color, count),
// 	}
// 	var i int32
// 	for i = 0; i < count; i++ {
// 		graphics.Colors[i] = color.RGBA{
// 			R: uint8(rand.Int31n(256)),
// 			G: uint8(rand.Int31n(256)),
// 			B: uint8(rand.Int31n(256)),
// 			A: 255,
// 		}
// 	}
// 	return &graphics
// }

// func updatePhysics(phys *PhysWorld, t float32) {
// 	for i := range phys.Objects {
// 		obj := &phys.Objects[i]
// 		velocity := &phys.Velocity[i]
// 		radius := phys.Radius[i]
// 		mass := math.Pi * float32(radius) * float32(radius)
// 		wallCollision := false
// 		if obj.X-float32(radius) <= 0 || obj.X+float32(radius) >= width {
// 			wallCollision = true
// 			velocity.X *= -1
// 			obj.X = rl.Clamp(obj.X, float32(radius), width-float32(radius))
// 		}
// 		if obj.Y-float32(radius) <= 0 || obj.Y+float32(radius) >= height {
// 			wallCollision = true
// 			velocity.Y *= -1
// 			obj.Y = rl.Clamp(obj.Y, float32(radius), height-float32(radius))
// 		}
// 		if wallCollision {
// 			*velocity = rl.Vector2Scale(*velocity, 1)
// 		}
// 		*obj = rl.Vector2Add(*obj, rl.Vector2Scale(*velocity, t))

// 		for j := range phys.Objects {
// 			if i == j {
// 				continue
// 			}
// 			otherObj := &phys.Objects[j]
// 			otherRadius := phys.Radius[j]
// 			otherVelocity := &phys.Velocity[j]
// 			otherMass := math.Pi * float32(otherRadius) * float32(otherRadius)
// 			distance := rl.Vector2Distance(*otherObj, *obj)
// 			overlap := float32(radius+otherRadius) - distance
// 			if overlap > 0 {
// 				// Compute normalized separation vector (from obj to otherObj)
// 				separation := rl.Vector2Normalize(rl.Vector2Subtract(*otherObj, *obj))

// 				// Position correction based on masses
// 				disp1 := -overlap - 1
// 				// disp2 := 0 //overlap / 2
// 				*obj = rl.Vector2Add(*obj, rl.Vector2Scale(separation, disp1))
// 				// *otherObj = rl.Vector2Add(*otherObj, rl.Vector2Scale(separation, 0))

// 				// Compute relative velocity
// 				relativeVelocity := rl.Vector2Subtract(*otherVelocity, *velocity)

// 				// Compute impulse for elastic collision (e = 1)
// 				impulse := (-2 * rl.Vector2DotProduct(relativeVelocity, separation)) / (1/mass + 1/otherMass)

// 				// Update velocities
// 				*velocity = rl.Vector2Subtract(*velocity, rl.Vector2Scale(separation, impulse/mass))
// 				*otherVelocity = rl.Vector2Add(*otherVelocity, rl.Vector2Scale(separation, impulse/otherMass))
// 			}
// 		}
// 	}
// }

// func draw(graphics *Graphics, phys *PhysWorld) {
// 	for i, col := range graphics.Colors {
// 		pos := phys.Objects[i]
// 		radius := phys.Radius[i]
// 		rl.DrawCircle(int32(pos.X), int32(pos.Y), float32(radius), col)
// 	}
// }

func randomColor() color.RGBA {
	return color.RGBA{
		R: uint8(rand.Intn(256)),
		G: uint8(rand.Intn(256)),
		B: uint8(rand.Intn(256)),
		A: 255,
	}
}
