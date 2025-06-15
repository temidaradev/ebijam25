package src

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/temidaradev/ebijam25/assets"
	"github.com/temidaradev/esset/v2"
	"image/color"
)

type Game struct {
}

func init() {
	assets.FontFaceS, _ = esset.GetFont(assets.Font, 16)
	assets.FontFaceM, _ = esset.GetFont(assets.Font, 32)
}

func NewGame() *Game {
	return &Game{}
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	esset.DrawText(screen, "Esset\nBasic Asset Implementer\nFor Ebitengine!", 0, 75, assets.FontFaceM, color.White)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
