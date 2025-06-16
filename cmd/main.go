package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/temidaradev/ebijam25/assets"
	"github.com/temidaradev/ebijam25/src"
)

func main() {
	g := src.NewGame()

	// Initial window setup for 1280x720
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("Ebijam 25 - Temidaradev")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowDecorated(true)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(60)
	ebiten.SetCursorMode(ebiten.CursorModeVisible)
	ebiten.SetCursorShape(ebiten.CursorShapeDefault)

	// Initialize display configuration for 1280x720 with aspect ratio scaling
	assets.UpdateDisplayConfig(1280, 720, assets.ScaleModeAspect, false)

	if err := ebiten.RunGame(&GameWrapper{game: g}); err != nil {
		panic(err)
	}
}

// GameWrapper wraps the game to handle display scaling and fullscreen
type GameWrapper struct {
	game *src.Game
}

func (gw *GameWrapper) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		if ebiten.IsFullscreen() {
			ebiten.SetFullscreen(false)
			ebiten.SetWindowSize(1280, 720)
			assets.UpdateDisplayConfig(1280, 720, assets.ScaleModeAspect, false)
		} else {
			// Switch to fullscreen
			ebiten.SetFullscreen(true)
			// Get monitor size for fullscreen scaling, but keep 1280x720 as base resolution
			w, h := ebiten.Monitor().Size()
			assets.UpdateDisplayConfig(w, h, assets.ScaleModeAspect, true)
		}
	}

	// Handle scaling mode changes (F1-F3) only when not in menu
	gameState := gw.game.GetState()
	if gameState != src.GameStateMenu {
		if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
			// Aspect ratio scaling (default)
			w, h := ebiten.WindowSize()
			if ebiten.IsFullscreen() {
				w, h = ebiten.Monitor().Size()
			}
			assets.UpdateDisplayConfig(w, h, assets.ScaleModeAspect, ebiten.IsFullscreen())
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
			// Stretch scaling
			w, h := ebiten.WindowSize()
			if ebiten.IsFullscreen() {
				w, h = ebiten.Monitor().Size()
			}
			assets.UpdateDisplayConfig(w, h, assets.ScaleModeStretch, ebiten.IsFullscreen())
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
			w, h := ebiten.WindowSize()
			if ebiten.IsFullscreen() {
				w, h = ebiten.Monitor().Size()
			}
			assets.UpdateDisplayConfig(w, h, assets.ScaleModePixel, ebiten.IsFullscreen())
		}
	}

	return gw.game.Update()
}

func (gw *GameWrapper) Draw(screen *ebiten.Image) {
	virtualScreen := ebiten.NewImage(1280, 720)

	gw.game.Draw(virtualScreen)

	if ebiten.IsFullscreen() {
		screen.Fill(color.RGBA{0, 0, 0, 255})
		assets.CurrentDisplayConfig.DrawScaledScreen(screen, virtualScreen)
	} else {
		screen.DrawImage(virtualScreen, nil)
	}
}

func (gw *GameWrapper) Layout(outsideWidth, outsideHeight int) (int, int) {
	if !ebiten.IsFullscreen() {
		currentConfig := assets.CurrentDisplayConfig
		if currentConfig.WindowWidth != outsideWidth || currentConfig.WindowHeight != outsideHeight {
			assets.UpdateDisplayConfig(outsideWidth, outsideHeight, currentConfig.Mode, false)
		}
		return 1280, 720
	} else {
		currentConfig := assets.CurrentDisplayConfig
		if currentConfig.WindowWidth != outsideWidth || currentConfig.WindowHeight != outsideHeight {
			assets.UpdateDisplayConfig(outsideWidth, outsideHeight, currentConfig.Mode, true)
		}
		return outsideWidth, outsideHeight
	}
}
