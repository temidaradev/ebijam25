package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/temidaradev/ebijam25/src"
)

func main() {
	g := src.NewGame()
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("Ebijam 25 - Temidara")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowDecorated(true)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(60)
	ebiten.SetCursorMode(ebiten.CursorModeVisible)
	ebiten.SetCursorShape(ebiten.CursorShapeDefault)

	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}
