package tile

import (
	"log"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/sunkink29/3dpacman/rendering"
	. "github.com/sunkink29/3dpacman/textures"
)

// Tile holds the vao and program pointers
type Tile struct {
	Pos         [2]float32
	layer       int
	TileOptions int32
}

func NewTile(pos [2]int, layer int, texture int32) Tile {

	return Tile{[2]float32{float32(pos[0]), float32(pos[1])}, layer, texture}
}

var tVao, tProgram uint32

func InitTileRendering(camera rendering.Camera) {
	// Configure the vertex and fragment shaders
	program, err := rendering.NewProgram(rendering.VertexShader, rendering.TileFragShader)
	if err != nil {
		panic(err)
	}
	gl.UseProgram(program)

	projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &camera.ProjectionMatrix[0])

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
	textureFileNames := TextureFilenames

	for i, fileName := range textureFileNames {
		texture, err := rendering.NewTexture(TextureDir + fileName + ".png")
		if err != nil {
			log.Fatalln(err)
		}
		gl.ActiveTexture(gl.TEXTURE1 + uint32(i))
		gl.BindTexture(gl.TEXTURE_2D, texture)
	}

	tVao = vao
	tProgram = program
}

func SetTileUniforms(viewMatrix mgl32.Mat4) {
	gl.UseProgram(tProgram)

	cameraUniform := gl.GetUniformLocation(tProgram, gl.Str("camera\x00"))
	gl.UniformMatrix4fv(cameraUniform, 1, false, &viewMatrix[0])

	textureIndexs := make([]int32, 0)
	for index := range TextureFilenames {
		textureIndexs = append(textureIndexs, int32(index+1))
	}

	textureUniform := gl.GetUniformLocation(tProgram, gl.Str("tex\x00"))
	gl.Uniform1iv(textureUniform, int32(len(textureIndexs)), &textureIndexs[0])

	wireframeUniform := gl.GetUniformLocation(tProgram, gl.Str("renderWireframe\x00"))
	gl.Uniform1i(wireframeUniform, rendering.RenderWireframe)
}

func (tile Tile) Render() {
	model := mgl32.Translate3D(tile.Pos[0], float32(tile.layer), tile.Pos[1])
	modelUniform := gl.GetUniformLocation(tProgram, gl.Str("model\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

	tileOptionsUniform := gl.GetUniformLocation(tProgram, gl.Str("tileOptions\x00"))
	gl.Uniform1i(tileOptionsUniform, tile.TileOptions)

	var color mgl32.Vec4
	if tile.TileOptions&WallAllAuto != 0 {
		color = mgl32.Vec4{0, 0, 1, 1}
	} else if tile.TileOptions&(Dot|DotBig) != 0 {
		color = mgl32.Vec4{1, 1, 0, 1}
	} else if tile.TileOptions&PlayerTex != 0 {
		color = mgl32.Vec4{1, 1, 0, 1}
	}
	// fmt.Println("color", color)

	colorUniform := gl.GetUniformLocation(tProgram, gl.Str("inputColor\x00"))
	gl.Uniform4fv(colorUniform, 1, &color[0])

	gl.BindVertexArray(tVao)
	gl.DrawArrays(gl.TRIANGLES, 0, 2*3)
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

