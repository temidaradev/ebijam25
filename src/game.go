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
	GameStateDead
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

	enemies []*Enemy
}

func init() {
	assets.FontFaceS, _ = esset.GetFont(assets.Font, 16)
	assets.FontFaceM, _ = esset.GetFont(assets.Font, 32)

	assets.InitTileMaps()
}

func NewGame() *Game {
	screenWidth, screenHeight := 1280, 720

	playerStartX := 100.0
	playerStartY := 100.0

	if assets.DesertTileMap != nil {
		assets.DesertTileMap.RecreateCollisionObjects()
	}

	enemies := []*Enemy{
		NewShooterEnemy(400, 350),
		NewJumperEnemy(700, 400),
		NewShooterEnemy(1000, 250),
		NewSpikeEnemy(600, 480),
		NewSpikeEnemy(900, 480),
		NewJumperEnemy(1300, 400),
		NewShooterEnemy(1600, 300),
		NewSpikeEnemy(1200, 480),
		NewJumperEnemy(1800, 450),
		NewShooterEnemy(2000, 350),
	}

	return &Game{
		state:              GameStateMenu,
		menu:               NewMenu(),
		player:             NewPlayer(playerStartX, playerStartY, float64(screenWidth), float64(screenHeight), 0, assets.DesertTileMap),
		lastFrameTime:      0,
		currentEnvironment: "desert",
		controller:         NewControllerInput(),
		showCollisionBoxes: false,
		enemies:            enemies,
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

		if inpututil.IsKeyJustPressed(ebiten.KeyC) {
			g.showCollisionBoxes = !g.showCollisionBoxes
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			g.player.ResetToSafePosition()
		}

		g.parallaxOffset += 0.5

		g.player.Update(deltaTime)

		playerX, playerY, _, _ := g.player.GetBounds()
		for _, enemy := range g.enemies {
			enemy.Update(deltaTime, playerX, playerY, g.player.CollisionSystem)

			for _, projectile := range enemy.Projectiles {
				if g.player.CheckProjectileCollision(projectile) {
					g.player.TakeDamage(1)
					projectile.IsActive = false
				}
			}

			if g.player.IsPerformingAttack() {
				enemyX, enemyY, enemyW, enemyH := enemy.GetBounds()
				if g.player.CheckAttackHit(enemyX, enemyY, enemyW, enemyH) {
					damage := g.player.GetAttackDamage()
					enemy.TakeDamageFromPlayer(damage)
				}
			}

			if !g.player.IsInvulnerable() && enemy.IsActive {
				enemyX, enemyY, enemyW, enemyH := enemy.GetBounds()
				playerX, playerY, playerW, playerH := g.player.GetBounds()

				if playerX < enemyX+enemyW && playerX+playerW > enemyX &&
					playerY < enemyY+enemyH && playerY+playerH > enemyY {
					g.player.TakeDamage(enemy.GetDamageDealt())
				}
			}
		}

		// Check if player died and trigger respawn menu
		if g.player.IsPlayerDead() {
			g.state = GameStateDead
			g.menu.SetRespawnState()
		}

	case GameStatePaused:
		g.menu.Update()

		if g.menu.IsContinueRequested() {
			g.state = GameStatePlaying
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.state = GameStatePlaying
		}

	case GameStateDead:
		g.menu.Update()

		if g.menu.IsRestartRequested() {
			// Restart the entire game - reset everything to initial state
			g.restartGame()
			g.state = GameStatePlaying
		}

		if g.menu.GetState() == MenuStateMain {
			// Player chose to exit to main menu
			g.state = GameStateMenu
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

		if assets.DesertTileMap != nil {
			assets.DesertTileMap.Draw(screen, cameraX, cameraY, float64(screenWidth), float64(screenHeight))
		}

		for _, enemy := range g.enemies {
			enemy.Draw(screen, camera)
		}

		g.drawPlayerWithCamera(screen, camera)

		px, py, pw, ph := g.player.GetBounds()
		screenPX, screenPY := camera.WorldToScreen(px, py)

		vector.StrokeRect(screen, float32(screenPX), float32(screenPY), float32(pw), float32(ph), 1, color.RGBA{0, 255, 0, 255}, false)

		if g.player.IsPerformingAttack() {
			ax, ay, aw, ah := g.player.GetAttackBox()
			screenAX, screenAY := camera.WorldToScreen(ax, ay)
			vector.StrokeRect(screen, float32(screenAX), float32(screenAY), float32(aw), float32(ah), 2, color.RGBA{255, 0, 0, 200}, false)
		}

		g.drawHealthBar(screen)
		g.drawCombatInfo(screen)
		g.drawControlsInfo(screen)

		fps := ebiten.ActualFPS()
		tps := ebiten.ActualTPS()
		fpsTpsText := fmt.Sprintf("FPS: %.0f  TPS: %.0f", fps, tps)
		esset.DrawText(screen, fpsTpsText, 10, 10, assets.FontFaceS, color.RGBA{255, 255, 255, 255})

	case GameStatePaused:
		camera := g.player.GetCamera()
		cameraX, cameraY := camera.GetView()

		layers := assets.GetLayersByEnvironment(g.currentEnvironment)
		assets.DrawBackgroundLayers(screen, layers, cameraX, cameraY, screenWidth, screenHeight)

		if assets.DesertTileMap != nil {
			assets.DesertTileMap.Draw(screen, cameraX, cameraY, float64(screenWidth), float64(screenHeight))
		}

		for _, enemy := range g.enemies {
			enemy.Draw(screen, camera)
		}

		g.drawPlayerWithCamera(screen, camera)

		px, py, pw, ph := g.player.GetBounds()
		screenPX, screenPY := camera.WorldToScreen(px, py)
		vector.StrokeRect(screen, float32(screenPX), float32(screenPY), float32(pw), float32(ph), 1, color.RGBA{0, 255, 0, 255}, false)

		fps := ebiten.ActualFPS()
		tps := ebiten.ActualTPS()
		fpsTpsText := fmt.Sprintf("FPS: %.0f  TPS: %.0f", fps, tps)
		esset.DrawText(screen, fpsTpsText, 10, 10, assets.FontFaceS, color.RGBA{255, 255, 255, 255})

		g.menu.Draw(screen)

	case GameStateDead:
		// Draw the game world in a darkened state
		camera := g.player.GetCamera()
		cameraX, cameraY := camera.GetView()

		layers := assets.GetLayersByEnvironment(g.currentEnvironment)
		assets.DrawBackgroundLayers(screen, layers, cameraX, cameraY, screenWidth, screenHeight)

		if assets.DesertTileMap != nil {
			assets.DesertTileMap.Draw(screen, cameraX, cameraY, float64(screenWidth), float64(screenHeight))
		}

		for _, enemy := range g.enemies {
			enemy.Draw(screen, camera)
		}

		g.drawPlayerWithCamera(screen, camera)

		// Draw dark overlay
		vector.DrawFilledRect(screen, 0, 0, float32(screenWidth), float32(screenHeight),
			color.RGBA{0, 0, 0, 120}, false)

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

func (g *Game) drawHealthBar(screen *ebiten.Image) {
	healthBarX := float32(20)
	healthBarY := float32(50)
	healthBarWidth := float32(200)
	healthBarHeight := float32(20)

	vector.DrawFilledRect(screen, healthBarX, healthBarY, healthBarWidth, healthBarHeight, color.RGBA{50, 50, 50, 200}, false)

	healthPercentage := g.player.GetHealthPercentage()
	fillWidth := healthBarWidth * float32(healthPercentage)
	healthColor := color.RGBA{255, 50, 50, 255}
	if healthPercentage > 0.6 {
		healthColor = color.RGBA{50, 255, 50, 255}
	} else if healthPercentage > 0.3 {
		healthColor = color.RGBA{255, 255, 50, 255}
	}

	vector.DrawFilledRect(screen, healthBarX, healthBarY, fillWidth, healthBarHeight, healthColor, false)

	healthText := fmt.Sprintf("Health: %d/%d", g.player.Health, g.player.MaxHealth)
	esset.DrawText(screen, healthText, float64(healthBarX), float64(healthBarY+healthBarHeight+5), assets.FontFaceS, color.RGBA{255, 255, 255, 255})
}

func (g *Game) drawControlsInfo(screen *ebiten.Image) {
	controlsY := 100
	lineHeight := 20

	controls := []string{
		"PARKOUR CONTROLS:",
		"A/D - Move Left/Right",
		"SPACE - Jump / Wall Jump / Double Jump",
		"X/C - Dash",
		"SHIFT/Z - Roll",
		"",
		"COMBAT CONTROLS:",
		"J/ENTER/LEFT CLICK - Attack",
		"Chain attacks for combos!",
		"",
		"TIPS:",
		"- Wall jump by pressing jump while touching a wall",
		"- Dash to cross large gaps",
		"- Roll to go under low obstacles",
		"- Attack enemies to defeat them",
		"- Avoid enemy projectiles and contact damage!",
	}

	for i, text := range controls {
		y := controlsY + i*lineHeight
		textColor := color.RGBA{255, 255, 255, 200}
		if text == "PARKOUR CONTROLS:" || text == "TIPS:" {
			textColor = color.RGBA{255, 255, 100, 255}
		}
		esset.DrawText(screen, text, 20.0, float64(y), assets.FontFaceS, textColor)
	}
}

func (g *Game) drawCombatInfo(screen *ebiten.Image) {
	y := 70.0

	if g.player.GetComboCount() > 1 {
		comboText := fmt.Sprintf("COMBO x%d", g.player.GetComboCount())
		esset.DrawText(screen, comboText, 10, y, assets.FontFaceM, color.RGBA{255, 255, 0, 255})
		y += 25.0
	}

	if g.player.IsPerformingAttack() {
		attackText := "ATTACKING!"
		esset.DrawText(screen, attackText, 10, y, assets.FontFaceS, color.RGBA{255, 100, 100, 255})
	}
}

// restartGame resets the game to initial state
func (g *Game) restartGame() {
	// Reset player to initial spawn position and full health
	g.player.X = 100.0
	g.player.Y = g.player.GroundLevel - float64(SpriteHeight)*g.player.Scale
	g.player.VelocityX = 0
	g.player.VelocityY = 0
	g.player.OnGround = true
	g.player.Health = g.player.MaxHealth
	g.player.IsDead = false
	g.player.InvulnTimer = 0

	// Reset camera position
	if g.player.Camera != nil {
		g.player.Camera.X = 0
		g.player.Camera.Y = 0
		g.player.Camera.TargetX = 0
		g.player.Camera.TargetY = 0
	}

	// Reset any other game state (enemies, etc.)
	for _, enemy := range g.enemies {
		enemy.Health = DefaultEnemyHealth
		enemy.IsActive = true
		// Clear projectiles
		enemy.Projectiles = enemy.Projectiles[:0]
	}

	// Reset parallax offset
	g.parallaxOffset = 0
}
