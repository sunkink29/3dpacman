package main

import (
	"fmt"
	"go/build"
	_ "image/png"
	"log"
	"runtime"
	"strconv"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/sunkink29/3dpacman/input"
	"github.com/sunkink29/3dpacman/maps"
	"github.com/sunkink29/3dpacman/player"
	"github.com/sunkink29/3dpacman/rendering"
	"github.com/sunkink29/3dpacman/rendering/text"
	"github.com/sunkink29/3dpacman/tile"
)

const speed = 5
const frameRate float64 = 60

var startMapSize = [2]int{28, 31}

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Samples, 4)
	window, err := glfw.CreateWindow(rendering.WindowWidth, rendering.WindowHeight, "PacMan", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	gl.Enable(gl.MULTISAMPLE)

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)

	window.SetKeyCallback(input.OnKeyPress)
	window.SetMouseButtonCallback(input.OnMouseButtonPress)

	cameraPos := mgl32.Vec3{14, 0, 15.5}
	// projectionMat := mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/windowHeight, 30.0, 50)
	projectionMat := mgl32.Ortho(-50/2, 50/2, -50*.75/2, 50*.75/2, 30, 50)
	viewMat := mgl32.LookAtV(cameraPos.Add(mgl32.Vec3{0, 40, 0}), cameraPos, mgl32.Vec3{0, 1, 0})
	camera := rendering.Camera{&cameraPos, &projectionMat, &viewMat}

	text.Init(window)

	frameRateText := text.New("Test", text.GetFont("8bitmadness", 30), mgl32.Vec2{380, 280}, mgl32.Vec3{1, 1, 1})
	defer frameRateText.Release()

	var frameRateEnable = false
	frameRateText.Hide()
	input.RegisterKeyBinding(glfw.KeyGraveAccent, "Toggle Test Text", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			if frameRateEnable {
				frameRateText.Hide()
			} else {
				frameRateText.Show()
			}
			frameRateEnable = !frameRateEnable
		}
	})

	tile.InitTileRendering(camera)

	curMap := maps.CreateEmptyMap(startMapSize)

	// move the managment of floating tiles to the map package

	testTile := tile.NewTile([2]int{-2, 0}, 0, 0, 0)

	// Configure global settings
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(0.5, 0.5, 0.5, 1.0)

	// angle := 0.0
	previousTime := time.Now()
	averageFrameRate := 0
	frameCount := 0

	rendering.RegisterMapBindings(&camera)
	maps.RegisterMapBindings(&curMap, &testTile, &camera)
	player.RegisterPlayerBindings()
	input.RegisterKeyBinding(glfw.KeyEscape, "quit", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		w.SetShouldClose(true)
	})

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Update
		curTime := time.Now()
		deltaTime := curTime.Sub(previousTime).Seconds()
		previousTime = curTime
		averageFrameRate += int(1 / deltaTime)
		if frameCount%5 == 0 {
			frameRateText.SetString(strconv.Itoa(averageFrameRate / 5))
			averageFrameRate = 0

		}
		frameCount++

		//angle += deltaTime
		// model := mgl32.HomogRotate3D(float32(angle), mgl32.Vec3{0, 1, 0})

		rendering.UpdateCameraPosition(&camera, speed, deltaTime)
		viewMat = mgl32.LookAtV(cameraPos.Add(mgl32.Vec3{0, 40, 0.1}), cameraPos, mgl32.Vec3{0, 1, 0})
		curMap.Update()

		// Render
		tile.SetTileUniforms(viewMat)
		testTile.Render()
		curMap.Render(deltaTime)
		frameRateText.Draw()

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()

		curFrameTime := time.Now().Sub(curTime)
		if curFrameTime.Seconds() < 1/frameRate {
			time.Sleep(time.Duration((1.0/(frameRate) - curFrameTime.Seconds()) * float64(time.Second)))
		}
	}
}

// Set the working directory to the root of Go package, so that its assets can be accessed.
func init() {
	// dir, err := importPathToDir("github.com/sunkink29/3dpacman/")
	// if err != nil {
	// 	log.Fatalln("Unable to find Go package in your GOPATH, it's needed to load assets:", err)
	// }
	// err = os.Chdir(dir)
	// if err != nil {
	// 	log.Panicln("os.Chdir:", err)
	// }
}

// importPathToDir resolves the absolute path from importPath.
// There doesn't need to be a valid Go package inside that import path,
// but the directory must exist.
func importPathToDir(importPath string) (string, error) {
	p, err := build.Import(importPath, "", build.FindOnly)
	if err != nil {
		return "", err
	}
	return p.Dir, nil
}
