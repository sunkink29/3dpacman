package player

import (
	"github.com/go-gl/glfw/v3.2/glfw"

	"github.com/sunkink29/3dpacman/input"
	"github.com/sunkink29/3dpacman/tile"
)

const speed = 5

type Player struct {
	pos       [2]int
	tile      tile.Tile
	targetPos [2]int
	targetDir [2]int
}

func New(pos [2]int) Player {
	tile := tile.NewTile(pos, 2, tile.PlayerTex, 0)
	return Player{pos, tile, [2]int{-1, -1}, [2]int{0, 0}}
}

func (curPlayer *Player) SetPos(pos [2]int) {
	curPlayer.pos = pos
	curPlayer.tile.Pos = [2]float32{float32(curPlayer.pos[0]), float32(curPlayer.pos[1])}
}

func (curPlayer *Player) Render(deltaTime float64) {
	if curPlayer.targetPos[0] != -1 && curPlayer.targetPos[1] != -1 {
		var targetDist [2]float32
		targetDist[0] = float32(curPlayer.targetPos[0]) - curPlayer.tile.Pos[0]
		targetDist[1] = float32(curPlayer.targetPos[1]) - curPlayer.tile.Pos[1]
		if targetDist[0]*float32(curPlayer.targetDir[0]) > 0 || targetDist[1]*float32(curPlayer.targetDir[1]) > 0 {
			curPlayer.tile.Pos[0] += float32(float64(curPlayer.targetDir[0]) * deltaTime * speed)
			curPlayer.tile.Pos[1] += float32(float64(curPlayer.targetDir[1]) * deltaTime * speed)
		} else {
			curPlayer.pos = curPlayer.targetPos
			curPlayer.tile.Pos = [2]float32{float32(curPlayer.targetPos[0]), float32(curPlayer.targetPos[1])}
			curPlayer.targetPos = [2]int{-1, -1}
			curPlayer.targetDir = [2]int{0, 0}
		}
	}
	curPlayer.tile.Render()
}

type GetMapTileType func(pos [2]int) tile.TileType

func (curPlayer *Player) UpdatePlayerPos(mapSize [2]int, getTileType GetMapTileType) {
	var nextTileType tile.TileType
	var nextPos [2]int
	nextPos[0] = curPlayer.pos[0] + movement[0]
	nextPos[1] = curPlayer.pos[1] + movement[1]
	if nextPos[0] >= 0 && nextPos[0] < mapSize[0] && nextPos[1] >= 0 && nextPos[1] < mapSize[1] {
		nextTileType = getTileType(nextPos)
	}
	if nextTileType != tile.Wall && curPlayer.targetPos[0] == -1 && curPlayer.targetPos[1] == -1 &&
		(movement[0] != 0 || movement[1] != 0) {
		curPlayer.targetPos[0] = movement[0] + curPlayer.pos[0]
		curPlayer.targetPos[1] = movement[1] + curPlayer.pos[1]
		curPlayer.targetDir = movement
	}
}

var movement [2]int
var lastPress int

func RegisterPlayerBindings() {
	input.RegisterKeyBinding(glfw.KeyUp, "Move Player Up", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press {
			movement[0] = 0
			movement[1] = -1
			lastPress = 1
		} else if action == glfw.Release && lastPress == 1 {
			movement[1] = 0
		}
	})
	input.RegisterKeyBinding(glfw.KeyDown, "Move Player Down", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press {
			movement[0] = 0
			movement[1] = 1
			lastPress = 2
		} else if action == glfw.Release && lastPress == 2 {
			movement[1] = 0
		}
	})
	input.RegisterKeyBinding(glfw.KeyLeft, "Move Player Left", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press {
			movement[0] = -1
			movement[1] = 0
			lastPress = 3
		} else if action == glfw.Release && lastPress == 3 {
			movement[0] = 0
		}
	})
	input.RegisterKeyBinding(glfw.KeyRight, "Move Player Right", func(w *glfw.Window, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press {
			movement[0] = 1
			movement[1] = 0
			lastPress = 4
		} else if action == glfw.Release && lastPress == 4 {
			movement[0] = 0
		}
	})
}
