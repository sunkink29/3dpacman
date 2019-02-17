package textures

const (
	WallUp      = 0x1
	WallDown    = 0x2
	WallLeft    = 0x4
	WallRight   = 0x8
	WallAll     = 0xF
	WallAuto    = 0x10
	WallAllAuto = 0x1F
	Dot         = 0x20
	DotBig      = 0x40
	PlayerTex   = 0x80
)

const TextureDir = "assets/textures/"

var TextureFilenames = []string{"wallUp", "wallDown", "wallLeft", "wallRight", "wallAuto", "dot", "bigDot", "pacman"}

