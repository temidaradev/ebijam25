package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/temidaradev/ebijam25/src"
)

func main() {
	g := src.NewGame()

	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("SCHIZOPHRENIC DESERT")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowDecorated(true)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(60)
	ebiten.SetCursorMode(ebiten.CursorModeVisible)
	ebiten.SetCursorShape(ebiten.CursorShapeDefault)

	gameWrapper := &GameWrapper{
		game:           g,
		windowedWidth:  1280,
		windowedHeight: 720,
	}

	if err := ebiten.RunGame(gameWrapper); err != nil {
		panic(err)
	}
}

type GameWrapper struct {
	game           *src.Game
	windowedWidth  int
	windowedHeight int
	isFullscreen   bool
}

func (gw *GameWrapper) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		gw.toggleFullscreen()
	}

	if gw.game.IsFullscreenToggleRequested() {
		gw.toggleFullscreen()
	}

	return gw.game.Update()
}

func (gw *GameWrapper) toggleFullscreen() {
	if gw.isFullscreen {
		fmt.Printf("Exiting fullscreen, restoring to %dx%d\n", gw.windowedWidth, gw.windowedHeight)
		gw.isFullscreen = false
		ebiten.SetFullscreen(false)
		ebiten.SetWindowSize(gw.windowedWidth, gw.windowedHeight)
	} else {
		currentW, currentH := ebiten.WindowSize()
		fmt.Printf("Current window size before fullscreen: %dx%d\n", currentW, currentH)

		if currentW >= 640 && currentH >= 480 {
			gw.windowedWidth = currentW
			gw.windowedHeight = currentH
		}

		fmt.Printf("Entering fullscreen, saved size: %dx%d\n", gw.windowedWidth, gw.windowedHeight)
		gw.isFullscreen = true
		ebiten.SetFullscreen(true)
	}
}

func (gw *GameWrapper) Draw(screen *ebiten.Image) {
	gw.game.Draw(screen)
}

func (gw *GameWrapper) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 1280, 720
}
