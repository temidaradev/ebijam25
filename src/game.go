package src

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/temidaradev/ebijam25/assets"
	"github.com/temidaradev/ebijam25/assets/images/start"
	"github.com/temidaradev/esset/v2"
)

// GameState represents the current state of the game
type GameState int

const (
	GameStateMenu GameState = iota
	GameStatePlaying
	GameStatePaused
)

type Game struct {
	state          GameState
	menu           *Menu
	parallaxOffset float64
}

func init() {
	assets.FontFaceS, _ = esset.GetFont(assets.Font, 16)
	assets.FontFaceM, _ = esset.GetFont(assets.Font, 32)
}

func NewGame() *Game {
	return &Game{
		state: GameStateMenu,
		menu:  NewMenu(),
	}
}

func (g *Game) Update() error {
	switch g.state {
	case GameStateMenu:
		g.menu.Update()

		if g.menu.IsStartSelected() {
			g.state = GameStatePlaying
		}

		if g.menu.IsExitSelected() {
			return ebiten.Termination
		}

	case GameStatePlaying:
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.state = GameStatePaused
		}
		g.parallaxOffset += 0.5

	case GameStatePaused:
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.state = GameStatePlaying
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case GameStateMenu:
		start.DrawStart(screen, g.parallaxOffset*0.1)
		g.menu.Draw(screen)

	case GameStatePlaying:
		start.DrawStart(screen, g.parallaxOffset)

	case GameStatePaused:
		start.DrawStart(screen, g.parallaxOffset)

		screenWidth, screenHeight := screen.Bounds().Dx(), screen.Bounds().Dy()

		vector.DrawFilledRect(screen, 0, 0, float32(screenWidth), float32(screenHeight),
			color.RGBA{0, 0, 0, 128}, false)

		pausedText := "PAUSED"
		pausedX := float64(screenWidth) * 0.025
		pausedY := float64(screenHeight) * 0.4
		esset.DrawText(screen, pausedText, pausedX, pausedY, assets.FontFaceM, color.RGBA{255, 255, 255, 255})

		hintText := "PRESS ESC TO RESUME"
		hintX := float64(screenWidth) * 0.025
		hintY := float64(screenHeight) * 0.5
		esset.DrawText(screen, hintText, hintX, hintY, assets.FontFaceS, color.RGBA{200, 200, 200, 255})
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
