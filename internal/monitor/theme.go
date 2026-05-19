package monitor

import rl "github.com/gen2brain/raylib-go/raylib"

var (
	colBgUI      = rl.NewColor(0x34, 0x2b, 0x22, 0xff)
	colBgDark    = rl.NewColor(0x1e, 0x1a, 0x15, 0xff)
	colBgFloor   = rl.NewColor(0x5b, 0x50, 0x1a, 0xff)
	colBgElement = rl.NewColor(0x4a, 0x40, 0x2c, 0xff)
	colTextMain  = rl.NewColor(0xec, 0xe1, 0xc9, 0xff)
	colTextMuted = rl.NewColor(0xa8, 0x9f, 0x8d, 0xff)
	colAccent    = rl.NewColor(0xb5, 0x7c, 0x00, 0xff)
	colDanger    = rl.NewColor(0x9d, 0x3b, 0x2b, 0xff)
	colBorder    = rl.NewColor(0x1f, 0x1a, 0x14, 0xff)

	texMonster rl.Texture2D
	texBoss    rl.Texture2D
	texPlayer  rl.Texture2D
	texDead    rl.Texture2D

	windowW int32 = 1280
	windowH int32 = 720
)

const (
	headerH int32 = 50
	sideW   int32 = 320
	entityW int32 = 50
	entityH int32 = 50
	bossW   int32 = 60
	bossH   int32 = 60
)