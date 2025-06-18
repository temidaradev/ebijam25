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
	showCollisionBoxes bool
}

func init() {
	assets.FontFaceS, _ = esset.GetFont(assets.Font, 16)
	assets.FontFaceM, _ = esset.GetFont(assets.Font, 32)

	// Initialize tilemaps
	assets.InitTileMaps()
}

func NewGame() *Game {
	screenWidth, screenHeight := 1280, 720

	playerStartX := 100.0
	playerStartY := 100.0 // Start higher up so player can fall onto tiles

	// Recreate collision objects to ensure they use the updated IsTileSolid logic
	if assets.DesertTileMap != nil {
		assets.DesertTileMap.RecreateCollisionObjects()
	}

	return &Game{
		state:              GameStateMenu,
		menu:               NewMenu(),
		player:             NewPlayer(playerStartX, playerStartY, float64(screenWidth), float64(screenHeight), 0, assets.DesertTileMap), // groundLevel set to 0 since we're not using it
		lastFrameTime:      0,
		currentEnvironment: "desert",
		controller:         NewControllerInput(),
		showCollisionBoxes: false,
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

		if inpututil.IsKeyJustPressed(ebiten.KeyC) {
			g.showCollisionBoxes = !g.showCollisionBoxes
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			g.player.ResetToSafePosition()
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
		cameraX, cameraY := camera.GetView()

		layers := assets.GetLayersByEnvironment(g.currentEnvironment)
		assets.DrawBackgroundLayers(screen, layers, cameraX, cameraY, screenWidth, screenHeight)

		// Draw the desert tilemap if in desert environment
		if g.currentEnvironment == "desert" {
			if assets.DesertTileMap != nil {
				assets.DesertTileMap.Draw(screen, cameraX, cameraY, float64(screenWidth), float64(screenHeight))
			}
		}

		g.drawPlayerWithCamera(screen, camera)

		px, py, pw, ph := g.player.GetBounds()
		screenPX, screenPY := camera.WorldToScreen(px, py)
		vector.StrokeRect(screen, float32(screenPX), float32(screenPY), float32(pw), float32(ph), 1, color.RGBA{0, 255, 0, 255}, false)

		// Draw FPS and TPS only
		fps := ebiten.ActualFPS()
		tps := ebiten.ActualTPS()
		fpsTpsText := fmt.Sprintf("FPS: %.0f  TPS: %.0f", fps, tps)
		esset.DrawText(screen, fpsTpsText, 10, 10, assets.FontFaceS, color.RGBA{255, 255, 255, 255})

		// Draw debug information
		playerDebugText := fmt.Sprintf("Player: (%.0f, %.0f) OnGround: %v VelY: %.0f", g.player.X, g.player.Y, g.player.OnGround, g.player.VelocityY)
		esset.DrawText(screen, playerDebugText, 10, 30, assets.FontFaceS, color.RGBA{255, 255, 255, 255})

		// Show controls
		esset.DrawText(screen, "C: Toggle collision boxes | R: Reset player position", 10, 50, assets.FontFaceS, color.RGBA{200, 200, 200, 255})

		// Show tile collision info at player position
		if assets.DesertTileMap != nil {
			hitboxX, hitboxY, _, _ := g.player.GetHitboxBounds()
			tileID, isSolid := assets.DesertTileMap.GetTileCollisionInfo(hitboxX, hitboxY)
			tileDebugText := fmt.Sprintf("Tile at player: ID=%d Solid=%v", tileID, isSolid)
			esset.DrawText(screen, tileDebugText, 10, 50, assets.FontFaceS, color.RGBA{255, 255, 255, 255})
		}

		// Controls help
		controlsText := "Controls: C=Toggle Collision Boxes, R=Reset Position"
		esset.DrawText(screen, controlsText, 10, 70, assets.FontFaceS, color.RGBA{200, 200, 200, 255})

		// Draw collision boxes if enabled
		if g.showCollisionBoxes {
			g.DrawCollisionBoxes(screen, camera)
		}

		// Draw stuck indicator if player is stuck
		if g.player.IsStuck() {
			esset.DrawText(screen, "PLAYER STUCK - Press R to reset", 10, 70, assets.FontFaceS, color.RGBA{255, 0, 0, 255})
		}
		if g.showCollisionBoxes {
			g.DrawCollisionBoxes(screen, camera)
		}

	case GameStatePaused:
		camera := g.player.GetCamera()
		cameraX, cameraY := camera.GetView()

		layers := assets.GetLayersByEnvironment(g.currentEnvironment)
		assets.DrawBackgroundLayers(screen, layers, cameraX, cameraY, screenWidth, screenHeight)

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

// DrawCollisionBoxes draws collision boxes for debugging
func (g *Game) DrawCollisionBoxes(screen *ebiten.Image, camera *Camera) {
	if g.currentEnvironment == "desert" && assets.DesertTileMap != nil {
		// Draw tile-based collision boxes
		tileWidth := float64(assets.DesertTileMap.TileWidth)
		tileHeight := float64(assets.DesertTileMap.TileHeight)

		// Draw a simple grid to show collision areas
		for y := 0; y < 10; y++ {
			for x := 0; x < 20; x++ {
				worldX := float64(x) * tileWidth
				worldY := float64(y) * tileHeight

				// Check if this tile has collision
				if assets.DesertTileMap.CheckCollision(worldX, worldY, tileWidth-1, tileHeight-1) {
					screenX, screenY := camera.WorldToScreen(worldX, worldY)

					// Draw collision box outline
					vector.StrokeRect(screen,
						float32(screenX), float32(screenY),
						float32(tileWidth), float32(tileHeight),
						1, color.RGBA{255, 0, 255, 100}, false)
				}
			}
		}
	}
}
