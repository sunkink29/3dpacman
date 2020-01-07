package maps

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"strconv"
	"strings"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sqweek/dialog"

	"github.com/sunkink29/3dpacman/input"
	"github.com/sunkink29/3dpacman/rendering"
	"github.com/sunkink29/3dpacman/textures"
	"github.com/sunkink29/3dpacman/tile"
)

// in the movement map each point is stored as binary where the first is up, the second is down
// third is left and the forth is right
// ex: 0110 is a point where you can move down and left
type Map struct {
	size [2]int
	tMap [][]tile.Tile // tile map: array that holds the tile vao and program
	mMap [][]int       // movement map: array that holds what directions the player can go at any point
}

func (curMap Map) Render() {
	for _, col := range curMap.tMap {
		for _, row := range col {
			row.Render()
		}
	}
}

func CreateEmptyMap(size [2]int) Map {
	tiles := make([][]tile.Tile, size[0])
	mMap := make([][]int, size[0])

	toggle := false
	for colIndex := range tiles {
		col := make([]tile.Tile, size[1])
		for rowIndex := range col {
			col[rowIndex] = tile.NewTile([2]int{colIndex, rowIndex}, 0, 0)
			toggle = !toggle
		}
		tiles[colIndex] = col
		mMap[colIndex] = make([]int, size[1])
	}

	return Map{size, tiles, mMap}
}

func (curMap *Map) GetMapTile(pos [2]int) int32 {
	return curMap.tMap[pos[0]][pos[1]].TileOptions
}

func (curMap *Map) ChangeMapTile(tile *tile.Tile, tileOptions int32) {
	tile.TileOptions = tileOptions
	if tileOptions&textures.WallAll == 0 {
		curMap.updateNearbyWall(tile, tileOptions)
	}
}

func (curMap *Map) GetSize() [2]int {
	return curMap.size
}

func (curMap *Map) updateNearbyWall(tile *tile.Tile, tileOptions int32) {
	deleteWall := int32(0)
	if tileOptions&textures.WallAuto != 0 {
		deleteWall = textures.WallAllAuto
	}
	if tile.Pos[1] > 0 {
		top := &curMap.tMap[int(tile.Pos[0])][int(tile.Pos[1])-1]
		if top.TileOptions&textures.WallAllAuto != 0 {
			tile.TileOptions |= textures.WallUp & deleteWall
			top.TileOptions |= textures.WallDown | textures.WallAuto
			top.TileOptions &= textures.WallAllAuto ^ (textures.WallDown & (deleteWall ^ textures.WallAll))
		}
	}
	if int(tile.Pos[1]) < curMap.size[1]-1 {
		bottem := &curMap.tMap[int(tile.Pos[0])][int(tile.Pos[1])+1]
		if bottem.TileOptions&textures.WallAllAuto != 0 {
			tile.TileOptions |= textures.WallDown & deleteWall
			bottem.TileOptions |= textures.WallUp | textures.WallAuto
			bottem.TileOptions &= textures.WallAllAuto ^ (textures.WallUp & (deleteWall ^ textures.WallAll))
		}
	}
	if tile.Pos[0] > 0 {
		left := &curMap.tMap[int(tile.Pos[0])-1][int(tile.Pos[1])]
		if left.TileOptions&textures.WallAllAuto != 0 {
			tile.TileOptions |= textures.WallLeft & deleteWall
			left.TileOptions |= textures.WallRight | textures.WallAuto
			left.TileOptions &= textures.WallAllAuto ^ (textures.WallRight & (deleteWall ^ textures.WallAll))
		}
	}
	if int(tile.Pos[0]) < curMap.size[0]-1 {
		right := &curMap.tMap[int(tile.Pos[0])+1][int(tile.Pos[1])]
		if right.TileOptions&textures.WallAllAuto != 0 {
			tile.TileOptions |= textures.WallRight & deleteWall
			right.TileOptions |= textures.WallLeft | textures.WallAuto
			right.TileOptions &= textures.WallAllAuto ^ (textures.WallLeft & (deleteWall ^ textures.WallAll))
		}
	}
}

func (curMap *Map) getSaveableMap() string {
	var mapString string

	mapString += fmt.Sprintf("%0*X%0*X", 2, curMap.size[0], 2, curMap.size[1])
	for _, col := range curMap.tMap {
		for _, tile := range col {
			mapString += fmt.Sprintf("%0*X", 2, tile.TileOptions)
		}
	}
	for _, col := range curMap.mMap {
		for _, move := range col {
			mapString += fmt.Sprintf("%0*X", 2, move)
		}
	}
	return mapString
}

func (curMap *Map) SaveToFile(filename string) error {
	curMapString := curMap.getSaveableMap()
	if !strings.HasSuffix(filename, ".pmap") {
		filename += ".pmap"
	}
	err := ioutil.WriteFile(filename, []byte(curMapString), 0644)
	if err != nil {
		return errors.New(fmt.Sprint("Error Saving Map to file:", err))
	}
	return nil
}

func LoadMapFromString(stringMap string) *Map {
	sliceMap := strings.Split(stringMap, "")
	mapWidth64, _ := strconv.ParseInt(strings.Join(sliceMap[0:2], ""), 16, 64)
	mapHeight64, _ := strconv.ParseInt(strings.Join(sliceMap[2:4], ""), 16, 64)
	mapWidth := int(mapWidth64)
	mapHeight := int(mapHeight64)
	if len(sliceMap) == mapWidth*mapHeight*4+4 {
		newMap := CreateEmptyMap([2]int{mapWidth, mapHeight})
		for i, col := range newMap.tMap {
			for j := range col {
				curStrIndex := (i*mapHeight+j)*2 + 4
				tileOptions, _ := strconv.ParseInt(strings.Join(sliceMap[curStrIndex:curStrIndex+2], ""), 16, 32)
				newMap.tMap[i][j].TileOptions = int32(tileOptions)
			}
		}
		for i, col := range newMap.tMap {
			for j := range col {
				curStrIndex := mapWidth*mapHeight*2 + (i*mapHeight+j)*2 + 4
				movement, _ := strconv.ParseInt(strings.Join(sliceMap[curStrIndex:curStrIndex+1], ""), 16, 0)
				newMap.mMap[i][j] = int(movement)
			}
		}

		return &newMap
	} else {
		fmt.Println("Error loading map: given map size and given map data do not match")
		return nil
	}
}

func LoadMapFromFile(filename string) (*Map, error) {
	mapBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.New(fmt.Sprint("Error Reading Map:", err))
	}
	mapString := string(mapBytes)

	return LoadMapFromString(mapString), nil
}

// stores the camera movement. each movement is stored as 4 bits: up, down, left right
var movement uint8 = 0x0

func UpdateCameraPosition(camera *rendering.Camera, speed float32, deltaTime float64) {
	camMovement := mgl32.Vec3{0, 0, 0}
	// Move up
	if movement&1 != 0 {
		camMovement = camMovement.Add(mgl32.Vec3{0, 0, 1})
	}
	// Move down
	if movement&2 != 0 {
		camMovement = camMovement.Add(mgl32.Vec3{0, 0, -1})
	}
	// Move right
	if movement&(1<<2) != 0 {
		camMovement = camMovement.Add(mgl32.Vec3{1, 0, 0})
	}
	// Move left
	if movement&(2<<2) != 0 {
		camMovement = camMovement.Add(mgl32.Vec3{-1, 0, 0})
	}
	*camera.CameraPos = camera.CameraPos.Add(camMovement.Mul(speed).Mul(float32(deltaTime)))
}

func RegisterMapBindings(curMap *Map, tTile *tile.Tile, camera *rendering.Camera) {
	input.RegisterKeyBinding(glfw.KeyI, "Move Camera Up", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press {
			movement |= 2
		} else if action == glfw.Release {
			movement &= 2 ^ 0xFF
		}
	})
	input.RegisterKeyBinding(glfw.KeyK, "Move Camera Down", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press {
			movement |= 1
		} else if action == glfw.Release {
			movement &= 1 ^ 0xFF
		}
	})
	input.RegisterKeyBinding(glfw.KeyJ, "Move Camera Left", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press {
			movement |= 2 << 2
		} else if action == glfw.Release {
			movement &= (2 << 2) ^ 0xFF
		}
	})
	input.RegisterKeyBinding(glfw.KeyL, "Move Camera right", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press {
			movement |= 1 << 2
		} else if action == glfw.Release {
			movement &= (1 << 2) ^ 0xFF
		}
	})
	input.RegisterKeyBinding(glfw.KeyW, "Toggle Up Wall Tile", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			tTile.TileOptions = (tTile.TileOptions ^ textures.WallUp) & textures.WallAll
		}
	})
	input.RegisterKeyBinding(glfw.KeyS, "Toggle Down Wall Tile", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			tTile.TileOptions = (tTile.TileOptions ^ textures.WallDown) & textures.WallAll
		}
	})
	input.RegisterKeyBinding(glfw.KeyA, "Toggle Left Wall Tile", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			tTile.TileOptions = (tTile.TileOptions ^ textures.WallLeft) & textures.WallAll
		}
	})
	input.RegisterKeyBinding(glfw.KeyD, "Toggle Right Wall Tile", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			tTile.TileOptions = (tTile.TileOptions ^ textures.WallRight) & textures.WallAll
		}
	})
	input.RegisterKeyBinding(glfw.KeyR, "Toggle Auto Wall Tile", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			tTile.TileOptions = (tTile.TileOptions ^ textures.WallAuto) & textures.WallAuto
		}
	})
	input.RegisterKeyBinding(glfw.KeyE, "Toggle Dot Tile", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			tTile.TileOptions = (tTile.TileOptions ^ textures.Dot) & textures.Dot
		}
	})
	input.RegisterKeyBinding(glfw.KeyQ, "Toggle Big Dot Tile", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			tTile.TileOptions = (tTile.TileOptions ^ textures.DotBig) & textures.DotBig
		}
	})
	input.RegisterKeyBinding(glfw.KeyZ, "Clear Tile", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			tTile.TileOptions = 0x0
		}
	})
	input.RegisterKeyBinding(glfw.KeyX, "Toggle WireFrame", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			rendering.RenderWireframe ^= 1
		}

	})
	input.RegisterKeyBinding(glfw.KeyC, "Load Map", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			filename, err := dialog.File().Filter("*.pmap", "pmap").Load()
			if err != nil {
				fmt.Println("Error getting map filename:", err)
				return
			}
			newMap, err := LoadMapFromFile(filename)
			if err != nil {
				fmt.Println(err)
				return
			}
			*curMap = *newMap
		}
	})
	input.RegisterKeyBinding(glfw.KeyV, "Save Map", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			filename, err := dialog.File().Filter("*.pmap", "pmap").Title("Save Map").Save()
			if err != nil {
				fmt.Println("Error getting map filename:", err)
				return
			}
			err = curMap.SaveToFile(filename)
			if err != nil {
				fmt.Println(err)
			}
		}
	})
	input.RegisterKeyBinding(glfw.KeyF, "Load Test Map", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			newMap, err := LoadMapFromFile("assets/maps/smallTestMap.pmap")
			if err != nil {
				fmt.Println(err)
				return
			}
			*curMap = *newMap
		}
	})
	input.RegisterMouseButtonBinding("map editor click", func(w *glfw.Window, button glfw.MouseButton, mod glfw.ModifierKey) {
		mouseX, mouseY := w.GetCursorPos()
		matProjection := camera.ProjectionMatrix.Mul4(*camera.ViewMatrix).Inv()
		worldPointf := rendering.ScreenToWorldSpace(w, [2]float64{mouseX, mouseY}, matProjection)
		worldPoint := []int{int(math.Floor(float64(worldPointf[0] + 0.5))), int(math.Floor(float64(worldPointf[2] + 0.5)))}
		if worldPoint[0] >= 0 && worldPoint[1] >= 0 && worldPoint[0] < curMap.size[0] && worldPoint[1] < curMap.size[1] {
			tile := &curMap.tMap[worldPoint[0]][worldPoint[1]]
			tileChange := int32(0)
			if button == glfw.MouseButton1 {
				tileChange = tTile.TileOptions
			} else if button == glfw.MouseButton2 {
				tileChange = 0x0 //tile.TileOptions & (tTile.TileOptions ^ 0xFF)
			}
			curMap.ChangeMapTile(tile, tileChange)
		}
	})
}

