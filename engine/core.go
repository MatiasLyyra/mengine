package engine

import (
	"image/color"
	"log"

	"github.com/MatiasLyyra/mengine/ecs"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Engine struct {
	World *ecs.World
}

type WindowSystem struct{}

func (ws *WindowSystem) Init(w *ecs.World) {
	settings := ecs.GetSingleton[WindowSettings](w)
	if settings.ConfigFlags != 0 {
		rl.SetConfigFlags(settings.ConfigFlags)
	}
	rl.SetTraceLogLevel(settings.LogLevel)
	rl.InitWindow(settings.Width, settings.Height, settings.Title)
	if settings.WindowState != 0 {
		rl.SetWindowState(settings.WindowState)
	}
	log.Printf("Initialized window")
}

type WindowSettings struct {
	Width       int32
	Height      int32
	Title       string
	ClearColor  color.RGBA
	ConfigFlags uint32
	WindowState uint32
	LogLevel    rl.TraceLogLevel
}

type Window struct {
	Width  float32
	Height float32
}

var window Window

func New() *Engine {
	e := &Engine{
		World: ecs.New(),
	}
	ecs.RegisterSingleton(e.World, &window)
	ecs.RegisterSingleton(e.World, &WindowSettings{
		LogLevel: rl.LogTrace,
	})
	ecs.RegisterInitSystem(e.World, &WindowSystem{})
	return e
}

func (e *Engine) Run() {
	e.World.Init()
	defer rl.CloseWindow()
	for !rl.WindowShouldClose() {
		window.Width = float32(rl.GetScreenWidth())
		window.Height = float32(rl.GetScreenHeight())
		dt := rl.GetFrameTime()
		rl.ClearBackground(rl.Black)
		rl.DrawFPS(10, 10)
		rl.BeginDrawing()
		e.World.RunUpdate(dt)
		rl.EndDrawing()
	}
}
