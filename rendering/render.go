package rendering

import (
	"fmt"
	"image"
	"image/draw"
	"os"
	"strings"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

const WindowWidth = 800
const WindowHeight = 600

var RenderWireframe int32

type Camera struct {
	CameraPos        *mgl32.Vec3
	ProjectionMatrix *mgl32.Mat4
	ViewMatrix       *mgl32.Mat4
}

func NewProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := CompileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := CompileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func CompileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func NewTexture(file string) (uint32, error) {
	imgFile, err := os.Open(file)
	if err != nil {
		return 0, fmt.Errorf("texture %q not found on disk: %v", file, err)
	}
	defer imgFile.Close()
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return 0, err
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return 0, fmt.Errorf("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))

	return texture, nil
}

func ScreenToWorldSpace(window *glfw.Window, point [2]float64, matProjection mgl32.Mat4) mgl32.Vec3 {
	depth := float32(0)
	pointer := unsafe.Pointer(&depth)
	gl.ReadPixels(int32(point[0]), WindowHeight-int32(point[1]), 1, 1, gl.DEPTH_COMPONENT, gl.FLOAT, pointer)
	winZ := depth

	var input [4]float32
	input[0] = (2.0 * (float32(point[0]-0) / (WindowWidth - 0))) - 1.0
	input[1] = 1.0 - (2.0 * (float32(point[1]-0) / (WindowHeight - 0)))
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

var VertexShader = `
#version 400
uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;
in vec3 vert;
in vec2 vertTexCoord;
out vec2 fragTexCoord;
void main() {
    fragTexCoord = vertTexCoord;
    gl_Position = projection * camera * model * vec4(vert, 1);
}
` + "\x00"

var TileFragShader = `
#version 400
uniform sampler2D tex[7];
uniform int tileOptions;
uniform vec4 inputColor;
uniform int renderWireframe;
uniform float borderWidth;
uniform float aspect;
in vec2 fragTexCoord;
out vec4 outputColor;

void renderTexture() {
	outputColor = vec4(0, 0, 0, 1);
	for (int i = 0; i < tex.length(); i++) {
		int useTex = min(tileOptions & (1 << i), 1);
		outputColor += texture(tex[i], fragTexCoord) * useTex;
	}

	outputColor = min(outputColor, 1);
	if (dot(vec3(outputColor), vec3(1)) != 0) {
		outputColor[0] = 1 - outputColor[0];
		outputColor[1] = 1 - outputColor[1];
		outputColor[2] = 1 - outputColor[2];
	}
	outputColor[2] = outputColor[0];
	outputColor *= inputColor;
}

void main() {
	if (renderWireframe == 1) {
		float maxX = 1.0 - borderWidth;
		float minX = borderWidth;
		float maxY = maxX / aspect;
		float minY = minX / aspect;

		outputColor = vec4(0);
		if (fragTexCoord.x < maxX && fragTexCoord.x > minX && fragTexCoord.y < maxY && fragTexCoord.y > minY) {
			renderTexture();
		} else {
			outputColor = vec4(1, 1, 1, 1);
		}
	} else {
		renderTexture();
	}
}
` + "\x00"
