package main

import (
	"fmt"
	"go/build"
	_ "image/png"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const windowWidth = 800
const windowHeight = 600
const speed = 5
const frameRate float64 = 60

var projectionMat mgl32.Mat4
var mapWidth = 28
var mapHeight = 31

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
	window, err := glfw.CreateWindow(windowWidth, windowHeight, "Cube", nil, nil)
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

	projectionMat = mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/windowHeight, 30.0, 50.0)

	initTileRendering()

	curMap := createEmptyMap([2]int{mapWidth, mapHeight})

	testTile := newTile([2]int{-2, 0}, 0)

	// Configure global settings
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(0.5, 0.5, 0.5, 1.0)

	// angle := 0.0
	previousTime := time.Now()

	cameraPos := mgl32.Vec3{14, 0, 15.5}
	viewMat := mgl32.LookAtV(cameraPos.Add(mgl32.Vec3{0, 40, 0}), cameraPos, mgl32.Vec3{0, 1, 0})

	setMouseButtonCallback(window, projectionMat, &viewMat, &curMap, &testTile)
	setKeyCallback(window, &testTile, &curMap)

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Update
		curTime := time.Now()
		deltaTime := curTime.Sub(previousTime).Seconds()
		previousTime = curTime

		//angle += deltaTime
		// model := mgl32.HomogRotate3D(float32(angle), mgl32.Vec3{0, 1, 0})

		cameraPos = cameraPos.Add(moveCamera(window).Mul(speed).Mul(float32(deltaTime)))
		viewMat = mgl32.LookAtV(cameraPos.Add(mgl32.Vec3{0, 40, 0.1}), cameraPos, mgl32.Vec3{0, 1, 0})

		// Render

		curMap.Render(viewMat)
		testTile.Render(viewMat)

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()

		curFrameTime := previousTime.Sub(time.Now())
		if curFrameTime.Seconds() < 1/frameRate {
			time.Sleep(time.Duration((1.0/(frameRate) - curFrameTime.Seconds()) * float64(time.Second)))
		}
	}
}

// Set the working directory to the root of Go package, so that its assets can be accessed.
func init() {
	dir, err := importPathToDir("github.com/sunkink29/3dpacman/")
	if err != nil {
		log.Fatalln("Unable to find Go package in your GOPATH, it's needed to load assets:", err)
	}
	err = os.Chdir(dir)
	if err != nil {
		log.Panicln("os.Chdir:", err)
	}
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
