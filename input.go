package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"strings"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sqweek/dialog"
)

func setMouseButtonCallback(window *glfw.Window, matProjection mgl32.Mat4, viewMat *mgl32.Mat4, curMap *Map, testTile *Tile) {
	window.SetMouseButtonCallback(
		func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
			onMouseButton(w, button, action, mod, matProjection.Mul4(*viewMat).Inv(), curMap, testTile)
		})
}

func onMouseButton(w *glfw.Window, button glfw.MouseButton, action glfw.Action,
	mod glfw.ModifierKey, matProjection mgl32.Mat4, curMap *Map, testTile *Tile) {
	if action == glfw.Release {
		mouseX, mouseY := w.GetCursorPos()
		worldPointf := screenToWorldSpace(w, [2]float64{mouseX, mouseY}, matProjection)
		worldPointf = worldPointf.Mul(1.04)
		worldPoint := []int{int(math.Floor(float64(worldPointf[0]))), int(math.Floor(float64(worldPointf[2])))}
		if worldPoint[0] >= 0 && worldPoint[1] >= 0 && worldPoint[0] < mapWidth && worldPoint[1] < mapHeight {
			tile := &curMap.tMap[worldPoint[0]][worldPoint[1]]
			if button == glfw.MouseButton1 {
				tile.tileOptions = testTile.tileOptions
			} else if button == glfw.MouseButton2 {
				tile.tileOptions = 0x0
			}
		}
	}
}

func setKeyCallback(window *glfw.Window, testTile *Tile, curMap *Map) {
	window.SetKeyCallback(
		func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			onKeyPress(w, key, scancode, action, mods, testTile, curMap)
		})
}

func onKeyPress(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey, testTile *Tile, curMap *Map) {
	if action == glfw.Release {
		curTileOptions := testTile.tileOptions
		if key == glfw.KeyW {
			curTileOptions ^= 0x1
			curTileOptions &= 0xF
		} else if key == glfw.KeyS {
			curTileOptions ^= 0x2
			curTileOptions &= 0xF
		} else if key == glfw.KeyA {
			curTileOptions ^= 0x4
			curTileOptions &= 0xF
		} else if key == glfw.KeyD {
			curTileOptions ^= 0x8
			curTileOptions &= 0xF
		} else if key == glfw.KeyE {
			curTileOptions ^= 0x10
			curTileOptions &= 0x10
		} else if key == glfw.KeyQ {
			curTileOptions ^= 0x20
			curTileOptions &= 0x20
		} else if key == glfw.KeyZ {
			curTileOptions = 0x0
		} else if key == glfw.KeyX {
			renderWireframe ^= 1
		} else if key == glfw.KeyC {
			curMapString := curMap.GetSaveableMap()
			filename, err := dialog.File().Filter("*.pmap", "pmap").Title("Save Map").Save()
			if err != nil {
				fmt.Println("Error: dialog box canceled/closed")
				return
			}
			// fmt.Println(strings.Contains(filename, ".pmap"))
			if !strings.Contains(filename, ".pmap") {
				filename += ".pmap"
			}
			err = ioutil.WriteFile(filename, []byte(curMapString), 0644)
			if err != nil {
				fmt.Println(err)
			}
			// fmt.Println("map:", curMap.GetSaveableMap())
		} else if key == glfw.KeyV {
			mapDir, err := dialog.File().Filter("*.pmap", "pmap").Load()
			if err != nil {
				fmt.Println("Error: dialog box canceled/closed")
				return
			}
			mapBytes, err := ioutil.ReadFile(mapDir)
			if err != nil {
				fmt.Println("Error Reading Map:", err)
				return
			}
			mapString := string(mapBytes)

			newMap := LoadMap(mapString)
			if newMap != nil {
				*curMap = *newMap
				mapWidth = newMap.size[0]
				mapHeight = newMap.size[1]
			}
		}
		testTile.tileOptions = curTileOptions
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
