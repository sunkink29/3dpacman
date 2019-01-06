package main

import (
	"fmt"
	"math"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func init() {
	registeredKeyBinding = make(map[glfw.Key]keyBinding)
}

func setMouseButtonCallback(window *glfw.Window, camera Camera, curMap *Map, testTile *Tile) {
	window.SetMouseButtonCallback(
		func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
			onMouseButton(w, button, action, mod, camera.projectionMatrix.Mul4(*camera.viewMatrix).Inv(), curMap, testTile)
		})
}

func onMouseButton(w *glfw.Window, button glfw.MouseButton, action glfw.Action,
	mod glfw.ModifierKey, matProjection mgl32.Mat4, curMap *Map, testTile *Tile) {
	if action == glfw.Release {
		mouseX, mouseY := w.GetCursorPos()
		worldPointf := screenToWorldSpace(w, [2]float64{mouseX, mouseY}, matProjection)
		worldPointf = worldPointf.Mul(1.04)
		worldPoint := []int{int(math.Floor(float64(worldPointf[0]))), int(math.Floor(float64(worldPointf[2])))}
		if worldPoint[0] >= 0 && worldPoint[1] >= 0 && worldPoint[0] < curMap.size[0] && worldPoint[1] < curMap.size[1] {
			tile := &curMap.tMap[worldPoint[0]][worldPoint[1]]
			if button == glfw.MouseButton1 {
				tile.tileOptions = testTile.tileOptions
			} else if button == glfw.MouseButton2 {
				tile.tileOptions = 0x0
			}
		}
	}
}

type KeyCallback func(w *glfw.Window, mods glfw.ModifierKey)
type keyBinding struct {
	name     string
	callback func(w *glfw.Window, mods glfw.ModifierKey)
}

var registeredKeyBinding map[glfw.Key]keyBinding

func RegisterKeyBinding(key glfw.Key, name string, callback KeyCallback) {
	binding, ok := registeredKeyBinding[key]
	if !ok {
		registeredKeyBinding[key] = keyBinding{name, callback}
	} else {
		fmt.Printf("Error binding key %v: key %v is bound to %v\n", name, key, binding.name)
	}
}

func OnKeyPress(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	binding, ok := registeredKeyBinding[key]
	if ok && action == glfw.Release {
		binding.callback(w, mods)
	}
}

func screenToWorldSpace(window *glfw.Window, point [2]float64, matProjection mgl32.Mat4) mgl32.Vec3 {
	winZ := float32(0.52)

	var input [4]float32
	input[0] = (2.0 * (float32(point[0]-0) / (windowWidth - 0))) - 1.0
	input[1] = 1.0 - (2.0 * (float32(point[1]-0) / (windowHeight - 0)))
	input[2] = 2.0*winZ - 1.0
	input[3] = 1

	inputV := mgl32.Vec4{input[0], input[1], input[2], input[3]}
	pos := matProjection.Mul4x1(inputV)
	pos[3] = 1.0 / pos[3]
	pos[0] *= pos[3]
	pos[1] *= pos[3]
	pos[2] *= pos[3]
	return pos.Vec3()
}

func moveCamera(window *glfw.Window) mgl32.Vec3 {
	movement := mgl32.Vec3{0, 0, 0}
	// Move forward
	if window.GetKey(glfw.KeyUp) == glfw.Press {
		movement = movement.Add(mgl32.Vec3{0, 0, 1})
	}
	// Move backward
	if window.GetKey(glfw.KeyDown) == glfw.Press {
		movement = movement.Add(mgl32.Vec3{0, 0, -1})
	}
	// Move right
	if window.GetKey(glfw.KeyRight) == glfw.Press {
		movement = movement.Add(mgl32.Vec3{1, 0, 0})
	}
	// Move left
	if window.GetKey(glfw.KeyLeft) == glfw.Press {
		movement = movement.Add(mgl32.Vec3{-1, 0, 0})
	}
	return movement
}
