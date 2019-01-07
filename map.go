package main

import (
	"errors"
	"fmt"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/scorpheus/dialog"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	wallUp      = 0x1
	wallDown    = 0x2
	wallLeft    = 0x4
	wallRight   = 0x8
	wallAll     = 0xF
	wallAuto    = 0x10
	wallAllAuto = 0x1F
	dot         = 0x20
	dotBig      = 0x40
)

// in the movement map each point is stored as binary where the first is up, the second is down
// third is left and the forth is right
// ex: 0110 is a point where you can move down and left
type Map struct {
	size [2]int
	tMap [][]Tile // tile map: array that holds the tile vao and program
	mMap [][]int  // movement map: array that holds what directions the player can go at any point
}

// Tile holds the vao and program pointers
type Tile struct {
	vao, program uint32
	pos          [2]int
	tileOptions  int32
}

var tVao, tProgram uint32
var renderWireframe int32

func (curMap Map) Render(viewMatrix mgl32.Mat4) {
	gl.UseProgram(tProgram)

	cameraUniform := gl.GetUniformLocation(tProgram, gl.Str("camera\x00"))
	gl.UniformMatrix4fv(cameraUniform, 1, false, &viewMatrix[0])

	textures := []int32{1, 2, 3, 4, 5, 6, 7}

	textureUniform := gl.GetUniformLocation(tProgram, gl.Str("tex\x00"))
	gl.Uniform1iv(textureUniform, int32(len(textures)), &textures[0])

	wireframeUniform := gl.GetUniformLocation(tProgram, gl.Str("renderWireframe\x00"))
	gl.Uniform1i(wireframeUniform, renderWireframe)

	for _, col := range curMap.tMap {
		for _, row := range col {
			row.Render()
		}
	}
}

// can not render only a tile without rendering a map
func (tile Tile) Render() {
	model := mgl32.Translate3D(float32(tile.pos[0]), 0, float32(tile.pos[1]))
	modelUniform := gl.GetUniformLocation(tile.program, gl.Str("model\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

	tileOptionsUniform := gl.GetUniformLocation(tile.program, gl.Str("tileOptions\x00"))
	gl.Uniform1i(tileOptionsUniform, tile.tileOptions)

	var color mgl32.Vec4
	if tile.tileOptions&wallAllAuto != 0 {
		color = mgl32.Vec4{0, 0, 1, 1}
	} else if tile.tileOptions&(dot|dotBig) != 0 {
		color = mgl32.Vec4{1, 1, 0, 1}
	}
	// fmt.Println("color", color)

	colorUniform := gl.GetUniformLocation(tile.program, gl.Str("inputColor\x00"))
	gl.Uniform4fv(colorUniform, 1, &color[0])

	gl.BindVertexArray(tile.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, 2*3)
}

func createEmptyMap(size [2]int) Map {
	tiles := make([][]Tile, size[0])
	mMap := make([][]int, size[0])

	toggle := false
	for colIndex := range tiles {
		col := make([]Tile, size[1])
		for rowIndex := range col {
			col[rowIndex] = newTile([2]int{colIndex, rowIndex}, 0)
			toggle = !toggle
		}
		tiles[colIndex] = col
		mMap[colIndex] = make([]int, size[1])
	}

	return Map{size, tiles, mMap}
}

func newTile(pos [2]int, texture int32) Tile {

	return Tile{tVao, tProgram, pos, texture}
}

func (curMap *Map) ChangeTileOptions(tile *Tile, tileOptions int32) {
	tile.tileOptions = tileOptions
	if tileOptions&wallAll == 0 {
		curMap.updateNearbyWall(tile, tileOptions)
	}
}

func (curMap *Map) updateNearbyWall(tile *Tile, tileOptions int32) {
	deleteWall := int32(0)
	if tileOptions&wallAuto != 0 {
		deleteWall = wallAll
	}
	if tile.pos[1] > 0 {
		top := &curMap.tMap[tile.pos[0]][tile.pos[1]-1]
		if top.tileOptions&wallAllAuto != 0 {
			tile.tileOptions |= wallUp & deleteWall
			top.tileOptions |= wallDown
			top.tileOptions &= wallAll ^ (wallDown & (deleteWall ^ wallAll))
		}
	}
	if tile.pos[1] < curMap.size[1]-1 {
		bottem := &curMap.tMap[tile.pos[0]][tile.pos[1]+1]
		if bottem.tileOptions&wallAllAuto != 0 {
			tile.tileOptions |= wallDown & deleteWall
			bottem.tileOptions |= wallUp
			bottem.tileOptions &= wallAll ^ (wallUp & (deleteWall ^ wallAll))
		}
	}
	if tile.pos[0] > 0 {
		left := &curMap.tMap[tile.pos[0]-1][tile.pos[1]]
		if left.tileOptions&wallAllAuto != 0 {
			tile.tileOptions |= wallLeft & deleteWall
			left.tileOptions |= wallRight
			left.tileOptions &= wallAll ^ (wallRight & (deleteWall ^ wallAll))
		}
	}
	if tile.pos[0] < curMap.size[0]-1 {
		right := &curMap.tMap[tile.pos[0]+1][tile.pos[1]]
		if right.tileOptions&wallAllAuto != 0 {
			tile.tileOptions |= wallRight & deleteWall
			right.tileOptions |= wallLeft
			right.tileOptions &= wallAll ^ (wallLeft & (deleteWall ^ wallAll))
		}
	}
}

func (curMap Map) getSaveableMap() string {
	var mapString string

	mapString += fmt.Sprintf("%0*X%0*X", 2, curMap.size[0], 2, curMap.size[1])
	for _, col := range curMap.tMap {
		for _, tile := range col {
			mapString += fmt.Sprintf("%0*X", 2, tile.tileOptions)
		}
	}
	for _, col := range curMap.mMap {
		for _, move := range col {
			mapString += fmt.Sprintf("%0*X", 2, move)
		}
	}
	return mapString
}

func (curMap Map) SaveToFile(filename string) error {
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
		newMap := createEmptyMap([2]int{mapWidth, mapHeight})
		for i, col := range newMap.tMap {
			for j := range col {
				curStrIndex := (i*mapHeight+j)*2 + 4
				tileOptions, _ := strconv.ParseInt(strings.Join(sliceMap[curStrIndex:curStrIndex+2], ""), 16, 32)
				newMap.tMap[i][j].tileOptions = int32(tileOptions)
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

func initTileRendering(camera Camera) {
	// Configure the vertex and fragment shaders
	program, err := newProgram(vertexShader, tileFragShader)
	if err != nil {
		panic(err)
	}
	gl.UseProgram(program)

	projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &camera.projectionMatrix[0])

	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	// Configure the vertex data
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(planeVertices)*4, gl.Ptr(planeVertices), gl.STATIC_DRAW)

	vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))

	texCoordAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vertTexCoord\x00")))
	gl.EnableVertexAttribArray(texCoordAttrib)
	gl.VertexAttribPointer(texCoordAttrib, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))

	borderWidthUniform := gl.GetUniformLocation(program, gl.Str("borderWidth\x00"))
	gl.Uniform1f(borderWidthUniform, 0.03)

	aspectUniform := gl.GetUniformLocation(program, gl.Str("aspect\x00"))
	gl.Uniform1f(aspectUniform, 1)

	// Load the textures
	textureFileNames := []string{"wallUp", "wallDown", "wallLeft", "wallRight", "wallAuto", "dot", "bigDot"}

	for i, fileName := range textureFileNames {
		texture, err := newTexture("textures/" + fileName + ".png")
		if err != nil {
			log.Fatalln(err)
		}
		gl.ActiveTexture(gl.TEXTURE1 + uint32(i))
		gl.BindTexture(gl.TEXTURE_2D, texture)
	}

	tVao = vao
	tProgram = program
}

func RegisterMapBindings(curMap *Map, tTile *Tile) {
	RegisterKeyBinding(glfw.KeyW, "Toggle Up Wall Tile", func(w *glfw.Window, mods glfw.ModifierKey) {
		tTile.tileOptions = (tTile.tileOptions ^ wallUp) & wallAll
	})
	RegisterKeyBinding(glfw.KeyS, "Toggle Down Wall Tile", func(w *glfw.Window, mods glfw.ModifierKey) {
		tTile.tileOptions = (tTile.tileOptions ^ wallDown) & wallAll
	})
	RegisterKeyBinding(glfw.KeyA, "Toggle Left Wall Tile", func(w *glfw.Window, mods glfw.ModifierKey) {
		tTile.tileOptions = (tTile.tileOptions ^ wallLeft) & wallAll
	})
	RegisterKeyBinding(glfw.KeyD, "Toggle Right Wall Tile", func(w *glfw.Window, mods glfw.ModifierKey) {
		tTile.tileOptions = (tTile.tileOptions ^ wallRight) & wallAll
	})
	RegisterKeyBinding(glfw.KeyR, "Toggle Auto Wall Tile", func(w *glfw.Window, mods glfw.ModifierKey) {
		tTile.tileOptions = (tTile.tileOptions ^ wallAuto) & wallAuto
	})
	RegisterKeyBinding(glfw.KeyE, "Toggle Dot Tile", func(w *glfw.Window, mods glfw.ModifierKey) {
		tTile.tileOptions = (tTile.tileOptions ^ dot) & dot
	})
	RegisterKeyBinding(glfw.KeyQ, "Toggle Big Dot Tile", func(w *glfw.Window, mods glfw.ModifierKey) {
		tTile.tileOptions = (tTile.tileOptions ^ dotBig) & dotBig
	})
	RegisterKeyBinding(glfw.KeyZ, "Clear Tile", func(w *glfw.Window, mods glfw.ModifierKey) {
		tTile.tileOptions = 0x0
	})
	RegisterKeyBinding(glfw.KeyX, "Toggle WireFrame", func(w *glfw.Window, mods glfw.ModifierKey) {
		renderWireframe ^= 1
	})
	RegisterKeyBinding(glfw.KeyC, "Load Map", func(w *glfw.Window, mods glfw.ModifierKey) {
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
	})
	RegisterKeyBinding(glfw.KeyV, "Save Map", func(w *glfw.Window, mods glfw.ModifierKey) {
		filename, err := dialog.File().Filter("*.pmap", "pmap").Title("Save Map").Save()
		if err != nil {
			fmt.Println("Error getting map filename:", err)
			return
		}
		err = curMap.SaveToFile(filename)
		if err != nil {
			fmt.Println(err)
		}
	})
	RegisterKeyBinding(glfw.KeyF, "Load Test Map", func(w *glfw.Window, mods glfw.ModifierKey) {
		newMap, err := LoadMapFromFile("maps/smallTestMap.pmap")
		if err != nil {
			fmt.Println(err)
			return
		}
		*curMap = *newMap
	})
}

var planeVertices = []float32{
	//  X, Y, Z, U, V
	-0.5, 0.5, -0.5, 0.0, 0.0,
	-0.5, 0.5, 0.5, 0.0, 1.0,
	0.5, 0.5, -0.5, 1.0, 0.0,
	0.5, 0.5, -0.5, 1.0, 0.0,
	-0.5, 0.5, 0.5, 0.0, 1.0,
	0.5, 0.5, 0.5, 1.0, 1.0,
}
