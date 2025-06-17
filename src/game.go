package src

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/temidaradev/ebijam25/assets"
	"github.com/temidaradev/esset/v2"
)

type GameState int

const (
	GameStateMenu GameState = iota
	GameStatePlaying
	GameStatePaused
)

type Game struct {
	state              GameState
	menu               *Menu
	parallaxOffset     float64
	player             *Player
	lastFrameTime      float64
	currentEnvironment string
	controller         *ControllerInput
}

func init() {
	assets.FontFaceS, _ = esset.GetFont(assets.Font, 16)
	assets.FontFaceM, _ = esset.GetFont(assets.Font, 32)
}

func NewGame() *Game {
	screenWidth, screenHeight := 1280, 720
	groundLevel := float64(screenHeight) - 30

	playerStartX := 100.0
	playerStartY := groundLevel - 25

	return &Game{
		state:              GameStateMenu,
		menu:               NewMenu(),
		player:             NewPlayer(playerStartX, playerStartY, float64(screenWidth), float64(screenHeight), groundLevel),
		lastFrameTime:      0,
		currentEnvironment: "forest",
		controller:         NewControllerInput(),
	}
}

func (g *Game) Update() error {
	g.controller.Update()

	deltaTime := 1.0 / 60.0
	if ebiten.ActualTPS() > 0 {
		deltaTime = 1.0 / ebiten.ActualTPS()
	}
	if deltaTime > 1.0/20.0 {
		deltaTime = 1.0 / 20.0
	}

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
		pausePressed := inpututil.IsKeyJustPressed(ebiten.KeyEscape) || g.controller.IsPauseJustPressed()

		if pausePressed {
			g.state = GameStatePaused
			g.menu.SetPauseState()
		}

		if inpututil.IsKeyJustPressed(ebiten.Key1) {
			g.currentEnvironment = "desert"
		}
		if inpututil.IsKeyJustPressed(ebiten.Key2) {
			g.currentEnvironment = "forest"
		}
		if inpututil.IsKeyJustPressed(ebiten.Key3) {
			g.currentEnvironment = "mountains"
		}
		if inpututil.IsKeyJustPressed(ebiten.Key4) {
			g.currentEnvironment = "cave"
		}

		g.parallaxOffset += 0.5

		g.player.Update(deltaTime)

	case GameStatePaused:
		g.menu.Update()

		if g.menu.IsContinueRequested() {
			g.state = GameStatePlaying
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.state = GameStatePlaying
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screenWidth, screenHeight := screen.Bounds().Dx(), screen.Bounds().Dy()

	switch g.state {
	case GameStateMenu:
		layers := assets.GetLayersByEnvironment(g.currentEnvironment)
		assets.DrawBackgroundLayers(screen, layers, g.parallaxOffset*0.1, 0, screenWidth, screenHeight)
		g.menu.Draw(screen)

	case GameStatePlaying:
		camera := g.player.GetCamera()
		cameraX, cameraY, _, _ := camera.GetView()

		layers := assets.GetLayersByEnvironment(g.currentEnvironment)
		assets.DrawBackgroundLayers(screen, layers, cameraX, cameraY, screenWidth, screenHeight)

		groundY := float32(g.player.GroundLevel)
		screenGroundX1, screenGroundY1 := camera.WorldToScreen(0, float64(groundY))
		screenGroundX2, screenGroundY2 := camera.WorldToScreen(float64(screenWidth), float64(groundY))
		vector.StrokeLine(screen, float32(screenGroundX1), float32(screenGroundY1), float32(screenGroundX2), float32(screenGroundY2), 2, color.RGBA{255, 0, 0, 255}, false)

		g.drawPlayerWithCamera(screen, camera)

		px, py, pw, ph := g.player.GetBounds()
		screenPX, screenPY := camera.WorldToScreen(px, py)
		vector.StrokeRect(screen, float32(screenPX), float32(screenPY), float32(pw), float32(ph), 1, color.RGBA{0, 255, 0, 255}, false)

		// Draw FPS and TPS only
		fps := ebiten.ActualFPS()
		tps := ebiten.ActualTPS()
		fpsTpsText := fmt.Sprintf("FPS: %.0f  TPS: %.0f", fps, tps)
		esset.DrawText(screen, fpsTpsText, 10, 10, assets.FontFaceS, color.RGBA{255, 255, 255, 255})

	case GameStatePaused:
		camera := g.player.GetCamera()
		cameraX, cameraY, _, _ := camera.GetView()

		layers := assets.GetLayersByEnvironment(g.currentEnvironment)
		assets.DrawBackgroundLayers(screen, layers, cameraX, cameraY, screenWidth, screenHeight)

		groundY := float32(g.player.GroundLevel)
		screenGroundX1, screenGroundY1 := camera.WorldToScreen(0, float64(groundY))
		screenGroundX2, screenGroundY2 := camera.WorldToScreen(float64(screenWidth), float64(groundY))
		vector.StrokeLine(screen, float32(screenGroundX1), float32(screenGroundY1), float32(screenGroundX2), float32(screenGroundY2), 2, color.RGBA{255, 0, 0, 255}, false)

		g.drawPlayerWithCamera(screen, camera)

		px, py, pw, ph := g.player.GetBounds()
		screenPX, screenPY := camera.WorldToScreen(px, py)
		vector.StrokeRect(screen, float32(screenPX), float32(screenPY), float32(pw), float32(ph), 1, color.RGBA{0, 255, 0, 255}, false)

		// Draw FPS and TPS only
		fps := ebiten.ActualFPS()
		tps := ebiten.ActualTPS()
		fpsTpsText := fmt.Sprintf("FPS: %.0f  TPS: %.0f", fps, tps)
		esset.DrawText(screen, fpsTpsText, 10, 10, assets.FontFaceS, color.RGBA{255, 255, 255, 255})

		// Draw the pause menu overlay
		g.menu.Draw(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 1280, 720
}

func (g *Game) GetState() GameState {
	return g.state
}

func (g *Game) IsFullscreenToggleRequested() bool {
	return g.menu.IsFullscreenToggleRequested()
}

func (g *Game) drawPlayerWithCamera(screen *ebiten.Image, camera *Camera) {
	if g.player.AnimationManager != nil {
		op := &ebiten.DrawImageOptions{}

		op.GeoM.Scale(g.player.Scale, g.player.Scale)

		if !g.player.FacingRight {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(50*g.player.Scale, 0)
		}

		op.GeoM.Translate(g.player.X, g.player.Y)

		cameraTransform := camera.GetTransform()
		op.GeoM.Concat(*cameraTransform)

		g.player.AnimationManager.DrawWithOptions(screen, op)
	}
}
