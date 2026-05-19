package monitor

import (
	_ "embed"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	//go:embed emoji/monster.png
	emojiMonsterPNG []byte
	//go:embed emoji/boss.png
	emojiBossPNG []byte
	//go:embed emoji/player.png
	emojiPlayerPNG []byte
	//go:embed emoji/dead.png
	emojiDeadPNG []byte
)

func loadEmojiTextures() {
	load := func(data []byte) rl.Texture2D {
		img := rl.LoadImageFromMemory(".png", data, int32(len(data)))
		if img == nil {
			return rl.Texture2D{}
		}
		t := rl.LoadTextureFromImage(img)
		rl.UnloadImage(img)
		return t
	}
	texMonster = load(emojiMonsterPNG)
	texBoss = load(emojiBossPNG)
	texPlayer = load(emojiPlayerPNG)
	texDead = load(emojiDeadPNG)
}