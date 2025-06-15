package src

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/temidaradev/ebijam25/assets"
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
	state              GameState
	menu               *Menu
	parallaxOffset     float64
	player             *Player
	lastFrameTime      float64
	currentEnvironment string  // Current background environment
	cameraX            float64 // Camera X position for parallax
	cameraY            float64 // Camera Y position for parallax
}

func init() {
	assets.FontFaceS, _ = esset.GetFont(assets.Font, 16)
	assets.FontFaceM, _ = esset.GetFont(assets.Font, 32)
}

func NewGame() *Game {
	screenWidth, screenHeight := 1280, 720
	// Set ground level to match where the ground images are positioned
	// Middle ground has yOffset=20, so ground should be about 20-30 pixels up from bottom
	groundLevel := float64(screenHeight) - 30

	// Position player at ground level
	playerStartX := 100.0            // 100 pixels from left edge
	playerStartY := groundLevel - 25 // Start just above ground level

	return &Game{
		state:              GameStateMenu,
		menu:               NewMenu(),
		player:             NewPlayer(playerStartX, playerStartY, float64(screenWidth), float64(screenHeight), groundLevel),
		lastFrameTime:      0,
		currentEnvironment: "desert", // Start with desert environment
		cameraX:            0,
		cameraY:            0,
	}
}

func (g *Game) Update() error {
	// Calculate delta time properly
	deltaTime := 1.0 / 60.0 // Fixed 60 FPS delta time
	if ebiten.ActualTPS() > 0 {
		deltaTime = 1.0 / ebiten.ActualTPS()
	}
	// Clamp deltaTime to avoid huge jumps (e.g., after alt-tab)
	if deltaTime > 1.0/20.0 {
		deltaTime = 1.0 / 20.0 // Max 1/20th of a second per frame
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
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.state = GameStatePaused
		}

		// Environment switching for testing
		if inpututil.IsKeyJustPressed(ebiten.Key1) {
			g.currentEnvironment = "desert"
		}
		if inpututil.IsKeyJustPressed(ebiten.Key2) {
			g.currentEnvironment = "forest"
		}
		if inpututil.IsKeyJustPressed(ebiten.Key3) {
			g.currentEnvironment = "mountains"
		}

		g.parallaxOffset += 0.5

		// Update player
		g.player.Update(deltaTime)

		// Update camera to follow player (simple follow camera)
		g.cameraX = g.player.X - 640 // Center camera on player (assuming 1280 screen width)
		g.cameraY = g.player.Y - 360 // Center camera on player (assuming 720 screen height)

	case GameStatePaused:
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
		// Draw background layers for menu (static camera for menu)
		layers := assets.GetLayersByEnvironment(g.currentEnvironment)
		assets.DrawBackgroundLayers(screen, layers, g.parallaxOffset*0.1, 0, screenWidth, screenHeight)
		g.menu.Draw(screen)

	case GameStatePlaying:
		// Draw background layers with camera-based parallax
		layers := assets.GetLayersByEnvironment(g.currentEnvironment)
		assets.DrawBackgroundLayers(screen, layers, g.cameraX, g.cameraY, screenWidth, screenHeight)

		// Draw ground line for debugging
		groundY := float32(g.player.GroundLevel)
		vector.StrokeLine(screen, 0, groundY, float32(screenWidth), groundY, 2, color.RGBA{255, 0, 0, 255}, false)

		// Draw player
		g.player.Draw(screen)

		// Debug: Draw player bounding box
		px, py, pw, ph := g.player.GetBounds()
		vector.StrokeRect(screen, float32(px), float32(py), float32(pw), float32(ph), 1, color.RGBA{0, 255, 0, 255}, false)

		// Draw environment switching instructions
		instructionsText := "Press 1=Desert, 2=Forest, 3=Mountains"
		esset.DrawText(screen, instructionsText, 10, 10, assets.FontFaceS, color.RGBA{255, 255, 255, 255})

	case GameStatePaused:
		// Draw background layers (same as playing but static)
		layers := assets.GetLayersByEnvironment(g.currentEnvironment)
		assets.DrawBackgroundLayers(screen, layers, g.cameraX, g.cameraY, screenWidth, screenHeight)

		// Draw ground line for debugging
		groundY := float32(g.player.GroundLevel)
		vector.StrokeLine(screen, 0, groundY, float32(screenWidth), groundY, 2, color.RGBA{255, 0, 0, 255}, false)

		// Draw player (still visible while paused)
		g.player.Draw(screen)

		// Debug: Draw player bounding box
		px, py, pw, ph := g.player.GetBounds()
		vector.StrokeRect(screen, float32(px), float32(py), float32(pw), float32(ph), 1, color.RGBA{0, 255, 0, 255}, false)

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
