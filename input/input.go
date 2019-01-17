package input

import (
	"fmt"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func init() {
	registeredKeyBinding = make(map[glfw.Key]keyBinding)
	registeredMouseButtonBinding = make([]mouseButtonBinding, 0)
}

type MouseButtonCallback func(w *glfw.Window, button glfw.MouseButton, mod glfw.ModifierKey)
type mouseButtonBinding struct {
	name     string
	callback MouseButtonCallback
}

var registeredMouseButtonBinding []mouseButtonBinding

func RegisterMouseButtonBinding(name string, callback MouseButtonCallback) {
	registeredMouseButtonBinding = append(registeredMouseButtonBinding, mouseButtonBinding{name, callback})
}

func OnMouseButtonPress(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	for _, binding := range registeredMouseButtonBinding {
		binding.callback(w, button, mod)
	}
}

type KeyCallback func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey)
type keyBinding struct {
	name     string
	callback KeyCallback
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
	if ok {
		binding.callback(w, action, mods)
	}
}

func MoveCamera(window *glfw.Window) mgl32.Vec3 {
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
