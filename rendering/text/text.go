package text

import (
	"fmt"
	"os"
	"strconv"

	"github.com/4ydx/gltext"
	v41 "github.com/4ydx/gltext/v4.1"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/image/math/fixed"
)

func New(str string, font *v41.Font, pos mgl32.Vec2, color mgl32.Vec3) *v41.Text {
	scaleMin, scaleMax := float32(1.0), float32(1.1)
	text := v41.NewText(font, scaleMin, scaleMax)
	text.SetString(str)
	text.SetColor(color)
	text.SetPosition(pos)
	return text
}

var loadedFonts map[string]*v41.Font

const fontLocation = "assets/fonts/"

var window *glfw.Window

func Init(win *glfw.Window) {
	window = win
	loadedFonts = make(map[string]*v41.Font)
}

func GetFont(name string, size int) *v41.Font {
	if font, ok := loadedFonts[name+"_"+strconv.Itoa(size)]; ok {
		return font
	} else {
		loadedFonts[name+"_"+strconv.Itoa(size)] = loadFont(name, size)
		return loadedFonts[name+"_"+strconv.Itoa(size)]
	}
}

func loadFont(name string, size int) *v41.Font {
	var font *v41.Font
	config, err := gltext.LoadTruetypeFontConfig(fontLocation+"fontConfigs", name+"_"+strconv.Itoa(size))
	if err == nil {
		if font, err = v41.NewFont(config); err != nil {
			panic(err)
		}
		fmt.Println("Font loaded from disk...")
	} else {
		fd, err := os.Open(fontLocation + name + ".ttf")
		if err != nil {
			panic(err)
		}
		defer fd.Close()

		runeRanges := make(gltext.RuneRanges, 0)
		runeRanges = append(runeRanges, gltext.RuneRange{Low: 32, High: 128})

		scale := fixed.Int26_6(size)
		runesPerRow := fixed.Int26_6(10)
		if config, err = gltext.NewTruetypeFontConfig(fd, scale, runeRanges, runesPerRow, scale/2); err != nil {
			panic(err)
		}
		if err = config.Save(fontLocation+"fontconfigs", name+"_"+strconv.Itoa(size)); err != nil {
			panic(err)
		}
		if font, err = v41.NewFont(config); err != nil {
			panic(err)
		}
	}
	width, height := window.GetSize()
	font.ResizeWindow(float32(width), float32(height))

	return font
}
