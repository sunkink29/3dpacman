package maps

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"strings"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/sqweek/dialog"

	"github.com/sunkink29/3dpacman/input"
	"github.com/sunkink29/3dpacman/player"
	"github.com/sunkink29/3dpacman/rendering"
	"github.com/sunkink29/3dpacman/tile"
)

// in the movement map each point is stored as binary where the first is up, the second is down
// third is left and the forth is right
// ex: 0110 is a point where you can move down and left
type Map struct {
	size      [2]int32
	tMap      [][]tile.Tile // tile map: array that holds the tile position and texture options
	playerObj player.Player
}

func (curMap *Map) Render(deltaTime float64) {
	for _, col := range curMap.tMap {
		for _, row := range col {
			row.Render()
		}
	}
	curMap.playerObj.Render(deltaTime)
}

func CreateEmptyMap(size [2]int) Map {
	tiles := make([][]tile.Tile, size[0])
	mMap := make([][]int, size[0])

	toggle := false
	for colIndex := range tiles {
		col := make([]tile.Tile, size[1])
		for rowIndex := range col {
			col[rowIndex] = tile.NewTile([2]int{colIndex, rowIndex}, 0, 0, 0)
			toggle = !toggle
		}
		tiles[colIndex] = col
		mMap[colIndex] = make([]int, size[1])
	}

	size32 := [2]int32{int32(size[0]), int32(size[1])}
	return Map{size32, tiles, player.New([2]int{2, 1})}
}

func (curMap *Map) GetMapTile(pos [2]int) tile.Tile {
	return curMap.tMap[pos[0]][pos[1]]
}

func (curMap *Map) GetSize() [2]int {
	return [2]int{int(curMap.size[0]), int(curMap.size[1])}
}

func (curMap *Map) ChangeMapTile(cTile *tile.Tile, tileType tile.TileType, flags tile.TileFlag) {
	cTile.Type = tileType
	cTile.Flags = flags
	if cTile.Type == tile.Wall && cTile.Flags&tile.All == 0 || cTile.Type != tile.Wall {
		curMap.updateNearbyWall(cTile)
	}
}

func (curMap *Map) GetPlayerSpawn() [2]int {
	for i, col := range curMap.tMap {
		for j, curTile := range col {
			if curTile.Type == tile.PlayerSpawn {
				return [2]int{i, j}
			}
		}
	}
	return [2]int{2, 2}
}

func (curMap *Map) updateNearbyWall(cTile *tile.Tile) {
	deleteWall := tile.TileFlag(0)
	if cTile.Type == tile.Wall && cTile.Flags&tile.All == 0 {
		deleteWall = tile.All
	}
	if cTile.Pos[1] > 0 {
		top := &curMap.tMap[int(cTile.Pos[0])][int(cTile.Pos[1])-1]
		if top.Type == tile.Wall {
			cTile.Flags |= tile.Up & deleteWall
			top.Flags |= tile.Down
			top.Flags &= tile.All ^ (tile.Down & (deleteWall ^ tile.All))
		}
	}
	if int32(cTile.Pos[1]) < curMap.size[1]-1 {
		bottem := &curMap.tMap[int(cTile.Pos[0])][int(cTile.Pos[1])+1]
		if bottem.Type == tile.Wall {
			cTile.Flags |= tile.Down & deleteWall
			bottem.Flags |= tile.Up
			bottem.Flags &= tile.All ^ (tile.Up & (deleteWall ^ tile.All))
		}
	}
	if cTile.Pos[0] > 0 {
		left := &curMap.tMap[int(cTile.Pos[0])-1][int(cTile.Pos[1])]
		if left.Type == tile.Wall {
			cTile.Flags |= tile.Left & deleteWall
			left.Flags |= tile.Right
			left.Flags &= tile.All ^ (tile.Right & (deleteWall ^ tile.All))
		}
	}
	if int32(cTile.Pos[0]) < curMap.size[0]-1 {
		right := &curMap.tMap[int(cTile.Pos[0])+1][int(cTile.Pos[1])]
		if right.Type == tile.Wall {
			cTile.Flags |= tile.Right & deleteWall
			right.Flags |= tile.Left
			right.Flags &= tile.All ^ (tile.Left & (deleteWall ^ tile.All))
		}
	}
}

func (curMap *Map) SaveToFile(filename string) error {
	if !strings.HasSuffix(filename, ".tmap") {
		filename += ".tmap"
	}

	data := []byte("tmap")
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(curMap.size[0]))
	data = append(data, bs...)

	binary.LittleEndian.PutUint32(bs, uint32(curMap.size[1]))
	data = append(data, bs...)

	for _, col := range curMap.tMap {
		for _, cTile := range col {
			bs := make([]byte, 2)
			binary.LittleEndian.PutUint16(bs, uint16(cTile.Type))
			data = append(data, bs...)

			binary.LittleEndian.PutUint16(bs, uint16(cTile.Flags))
			data = append(data, bs...)
		}
	}
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return errors.New(fmt.Sprint("Error Saving Map to file:", err))
	}
	return nil
}

func LoadMapFromFile(filename string) (*Map, error) {
	const SIZEOF_INT32 = 4
	const SIZEOF_INT16 = 2
	const SIZEOF_TYPESTRING = len("tmap")

	mapBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.New(fmt.Sprint("Error Reading Map:", err))
	}
	mapBytes = mapBytes[SIZEOF_TYPESTRING:]

	var mapSize [2]int
	mapSize[0] = int(binary.LittleEndian.Uint32(mapBytes[0:SIZEOF_INT32]))
	mapSize[1] = int(binary.LittleEndian.Uint32(mapBytes[SIZEOF_INT32 : SIZEOF_INT32*2]))

	mapBytes = mapBytes[SIZEOF_INT32*2:]

	if len(mapBytes) == mapSize[0]*mapSize[1]*SIZEOF_INT32 {
		newMap := CreateEmptyMap(mapSize)
		for i, col := range newMap.tMap {
			for j := range col {
				curIndex := (i*mapSize[1] + j)
				newMap.tMap[i][j].Type = tile.TileType(binary.LittleEndian.Uint16(mapBytes[curIndex*SIZEOF_INT32 : curIndex*SIZEOF_INT32+SIZEOF_INT16]))
				newMap.tMap[i][j].Flags = tile.TileFlag(binary.LittleEndian.Uint16(mapBytes[curIndex*SIZEOF_INT32+SIZEOF_INT16 : (curIndex+1)*SIZEOF_INT32]))
			}
		}
		newMap.playerObj.SetPos(newMap.GetPlayerSpawn())
		return &newMap, nil
	}
	return nil, errors.New("Error loading map: given map size and given map data do not match")
}

func (curMap *Map) Update() {
	curMap.playerObj.UpdatePlayerPos(curMap.GetSize(), func(pos [2]int) tile.TileType { return curMap.GetMapTile(pos).Type })
}

func RegisterMapBindings(curMap *Map, tTile *tile.Tile, camera *rendering.Camera) {
	input.RegisterKeyBinding(glfw.KeyW, "Toggle Up Wall Tile", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			tTile.Type = tile.Wall
			tTile.Flags ^= tile.Up
			if tTile.Flags&tile.All == 0 {
				tTile.Type = tile.Blank
			}
		}
	})
	input.RegisterKeyBinding(glfw.KeyS, "Toggle Down Wall Tile", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			tTile.Type = tile.Wall
			tTile.Flags ^= tile.Down
			if tTile.Flags&tile.All == 0 {
				tTile.Type = tile.Blank
			}
		}
	})
	input.RegisterKeyBinding(glfw.KeyA, "Toggle Left Wall Tile", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			tTile.Type = tile.Wall
			tTile.Flags ^= tile.Left
			if tTile.Flags&tile.All == 0 {
				tTile.Type = tile.Blank
			}
		}
	})
	input.RegisterKeyBinding(glfw.KeyD, "Toggle Right Wall Tile", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			tTile.Type = tile.Wall
			tTile.Flags ^= tile.Right
			if tTile.Flags&tile.All == 0 {
				tTile.Type = tile.Blank
			}
		}
	})
	input.RegisterKeyBinding(glfw.KeyR, "Toggle Auto Wall Tile", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			if tTile.Type != tile.Wall {
				tTile.Type = tile.Wall
			} else {
				if tTile.Flags&tile.All != 0 {
					tTile.Flags &= 0xFFF0
				} else {
					tTile.Type = tile.Blank
				}
			}
		}
	})
	input.RegisterKeyBinding(glfw.KeyE, "Toggle Dot Tile", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			if tTile.Type != tile.Dot {
				tTile.Type = tile.Dot
			} else {
				tTile.Type = tile.Blank
			}

		}
	})
	input.RegisterKeyBinding(glfw.KeyQ, "Toggle Big Dot Tile", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			if tTile.Type != tile.DotBig {
				tTile.Type = tile.DotBig
			} else {
				tTile.Type = tile.Blank
			}
		}
	})
	input.RegisterKeyBinding(glfw.KeyT, "Toggle Player Spawn Tile", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			if tTile.Type != tile.PlayerSpawn {
				tTile.Type = tile.PlayerSpawn
			} else {
				tTile.Type = tile.Blank
			}
		}
	})
	input.RegisterKeyBinding(glfw.KeyZ, "Clear Tile", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			tTile.Type = tile.Blank
			tTile.Flags = 0x0
		}
	})
	input.RegisterKeyBinding(glfw.KeyX, "Toggle WireFrame", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			rendering.RenderWireframe ^= 1
		}

	})
	input.RegisterKeyBinding(glfw.KeyC, "Load Map", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Release {
			filename, err := dialog.File().Filter("*.tmap", "tmap").Load()
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
			filename, err := dialog.File().Filter("*.tmap", "tmap").Title("Save Map").Save()
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
			newMap, err := LoadMapFromFile("assets/maps/smallTestMap.tmap")
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
		if worldPoint[0] >= 0 && worldPoint[1] >= 0 && worldPoint[0] < int(curMap.size[0]) && worldPoint[1] < int(curMap.size[1]) {
			cTile := &curMap.tMap[worldPoint[0]][worldPoint[1]]
			tileChange := *tTile
			if button == glfw.MouseButton2 {
				tileChange.Type = tile.Blank
				tileChange.Flags = 0x0
			}
			curMap.ChangeMapTile(cTile, tileChange.Type, tileChange.Flags)
		}
	})
}
