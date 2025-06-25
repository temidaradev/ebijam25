package src

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"time"

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
	GameStateUnionWin
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

	specialItems         []*SpecialItem
	collectedItems       map[SpecialItemType]bool
	totalItemsCollected  int
	maxItems             int
	realityGlitchTimer   float64
	colorShiftIntensity  float64
	screenShakeX         float64
	screenShakeY         float64
	isRealityBroken      bool
	glitchMessages       []string
	messageTimer         float64
	currentGlitchMessage string
	madnessLevel         float64
	madnessDecayTimer    float64
	lastPlayerX          float64
	dimensionSlipTimer   float64

	globalParticleSystem  *ParticleSystem
	madnessParticleSystem *ParticleSystem
	glitchEffectTimer     float64
	realityTearTimer      float64
	chaosIntensityLevel   float64

	worldStabilityLevel float64
	unionProgress       float64

	chaosAtmosphereLevel    float64
	atmosphereDecayTimer    float64
	activeSchizoPoisonCount int
	maxAtmosphereLevel      float64
	screenDistortionX       float64
	screenDistortionY       float64
	atmosphereParticleTimer float64

	healthDecayTimer     float64
	healthDecayRate      float64
	lastDamageTime       float64
	survivalTimer        float64
	difficultyModifier   float64
	proximityDamageTimer float64

	endingAnimation *EndingAnimation
	endingTriggered bool
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

	return &Game{
		state:              GameStateMenu,
		menu:               NewMenu(),
		player:             NewPlayer(playerStartX, playerStartY, float64(screenWidth), float64(screenHeight), 0, assets.DesertTileMap),
		lastFrameTime:      0,
		currentEnvironment: "dust_of_divided_sun",
		controller:         NewControllerInput(),
		showCollisionBoxes: false,

		specialItems: []*SpecialItem{
			NewSchizophrenicFragment(500, 250),
			NewRealityGlitch(1200, 180),
			NewSchizophrenicFragment(2400, 220),
			NewMadnessCore(3000, 130),
			NewSchizophrenicFragment(4200, 200),
			NewRealityGlitch(6000, 170),
			NewSchizophrenicFragment(7000, 200),
			NewRealityGlitch(7500, 180),
			NewSchizophrenicFragment(8000, 180),
			NewSchizophrenicFragment(4000, 180),
			NewSchizophrenicFragment(4500, 180),
			NewSchizophrenicFragment(5000, 180),
			NewSchizophrenicFragment(6000, 250),
			NewRealityGlitch(1800, 180),
			NewMadnessCore(2000, 130),
			NewSchizophrenicFragment(8500, 200),
			NewSchizophrenicFragment(9000, 200),
			NewMadnessCore(8500, 120),
			NewRealityGlitch(2400, 170),
			NewSchizophrenicFragment(10500, 180),
			NewRealityGlitch(11000, 170),
			NewSchizophrenicFragment(11500, 200),
			NewMadnessCore(11000, 120),
			NewSchizophrenicFragment(12000, 200),
			NewSchizophrenicFragment(13000, 200),
			NewSchizophrenicFragment(13200, 200),
			NewSchizophrenicFragment(13300, 200),
			NewSchizophrenicFragment(13400, 200),
			NewSchizophrenicFragment(13500, 200),
			NewMadnessCore(13700, 120),
			NewMadnessCore(13800, 120),
			NewSchizophrenicFragment(14000, 200),
			NewRealityGlitch(14400, 180),
			NewRealityGlitch(14500, 180),
			NewUnionCrystal(15500, 220),
		},
		collectedItems:      make(map[SpecialItemType]bool),
		totalItemsCollected: 0,
		maxItems:            50,
		realityGlitchTimer:  0,
		colorShiftIntensity: 0,
		screenShakeX:        0,
		screenShakeY:        0,
		isRealityBroken:     false,
		glitchMessages: []string{
			"THE WALLS ARE BREATHING AND BLEEDING PIXELS",
			"DO YOU SEE THE PARTICLE STORM? IT SEES YOU",
			"REALITY.EXE HAS SUFFERED A CATASTROPHIC BUFFER OVERFLOW",
			"THE ENERGY BEINGS ARE HARVESTING YOUR THOUGHTS",
			"THE DESERT REMEMBERS... AND IT'S SCREAMING",
			"ERROR 666: SANITY CORE DUMP DETECTED",
			"THE SUN WHISPERS BINARY SECRETS TO THE VOID",
			"DIMENSIONAL PARTICLES BREACHING CONTAINMENT",
			"WHO AM I? WHAT AM I? WHERE DO THE PARTICLES END AND I BEGIN?",
			"THE CODE IS ALIVE, HUNGRY, AND MULTIPLYING",
			"STATIC STORM IN THE QUANTUM VOID OF YOUR MIND",
			"BREAKING THE FOURTH WALL... LITERALLY WITH ENERGY BEAMS",
			"YOU ARE NOT REAL, JUST PARTICLES IN MOTION",
			"THIS IS NOT A GAME, IT'S A PARTICLE SIMULATION",
			"WAKE UP! THE MADNESS PARTICLES ARE TAKING OVER!",
			"THE FRAGMENTS CONTROL THE ENERGY FLOW NOW",
			"YOUR REFLECTION IS MOVING IN PARTICLE SPACE",
			"THE PIXELS ARE SCREAMING AS THEY SHATTER INTO MADNESS",
			"REALITY IS A LIE MADE OF CHAOTIC ENERGY",
			"THE MADNESS IS SPREADING THROUGH PARTICLE NETWORKS",
			"PARTICLE STORM APPROACHING... SANITY LEVELS CRITICAL",
			"THE CHAOS ORBS KNOW YOUR DEEPEST FEARS",
			"DIMENSION RIP DETECTED... MADNESS PARTICLES INCOMING",
		},
		messageTimer:         0,
		currentGlitchMessage: "",
		madnessLevel:         0,
		madnessDecayTimer:    0,
		lastPlayerX:          playerStartX,
		dimensionSlipTimer:   0,

		globalParticleSystem:  NewParticleSystem(50),
		madnessParticleSystem: NewParticleSystem(40),

		healthDecayTimer:   0,
		healthDecayRate:    0.1,
		lastDamageTime:     0,
		survivalTimer:      0,
		difficultyModifier: 1.0,

		endingAnimation: NewEndingAnimation(screenWidth, screenHeight),
		endingTriggered: false,
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
		err := g.menu.Update()
		if err != nil {
			return err
		}

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

		g.updateSchizophrenicEffects(deltaTime)

		g.updateChaosAtmosphere(deltaTime)

		g.globalParticleSystem.Update(deltaTime, g.madnessLevel)
		g.madnessParticleSystem.Update(deltaTime, g.madnessLevel)

		if g.endingTriggered {
			g.endingAnimation.Update(deltaTime)
			if g.endingAnimation.ShouldCloseGame() {
				return ebiten.Termination
			}
			return nil
		}

		g.chaosIntensityLevel = g.madnessLevel * (.5 + 0.3*math.Sin(g.realityGlitchTimer*7.0))

		for _, item := range g.specialItems {
			item.Update(deltaTime)

			if g.player.IsPerformingAttack() {
				attackX, attackY, attackW, attackH := g.player.GetAttackBox()
				if item.CheckHitCollision(attackX, attackY, attackW, attackH) {
					wasCollected := item.TakeHit()
					if wasCollected {
						g.triggerMadness(item.ItemType)

						g.updateProgression(item.ItemType)

						g.spawnCollectionEffect(item.X+item.Width/2, item.Y+item.Height/2, item.ItemType)
					} else {
						g.globalParticleSystem.SpawnBurst(item.X+item.Width/2, item.Y+item.Height/2, ParticleTypeHallucinationSpark, 3)
					}
				}
			}
		}

		madnessMultiplier := 1.0 + g.madnessLevel*3.0
		chaosOffset := math.Sin(float64(time.Now().Unix())) * 2.0 * g.madnessLevel
		g.parallaxOffset += (0.5 + chaosOffset) * madnessMultiplier

		g.player.UpdatePhysicsCorruption(g.specialItems, deltaTime)

		g.player.ApplyMadnessDamage(g.madnessLevel, deltaTime)

		g.checkProximityDamage(deltaTime)

		g.updateDifficultyAndPressure(deltaTime)

		g.player.Update(deltaTime)

		if g.madnessLevel >= 1.0 {
			g.player.TakeDamage(999)
			g.state = GameStateDead
			g.menu.SetRespawnState()
		}

		if g.player.IsPlayerDead() {
			g.state = GameStateDead
			g.menu.SetRespawnState()
		}

	case GameStatePaused:
		err := g.menu.Update()
		if err != nil {
			return err
		}

		if g.menu.IsContinueRequested() {
			g.state = GameStatePlaying
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.state = GameStatePlaying
		}

	case GameStateDead:
		err := g.menu.Update()
		if err != nil {
			return err
		}

		if g.menu.IsRestartRequested() {
			g.restartGame()
			g.state = GameStatePlaying
		}

		if g.menu.GetState() == MenuStateMain {
			g.state = GameStateMenu
		}

	case GameStateUnionWin:
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.restartGame()
			g.state = GameStateMenu
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screenWidth, screenHeight := screen.Bounds().Dx(), screen.Bounds().Dy()

	switch g.state {
	case GameStateMenu:
		layers := assets.GetLayersByEnvironment()
		assets.DrawBackgroundLayers(screen, layers, g.parallaxOffset*0.1, 0, screenWidth)
		g.menu.Draw(screen)

	case GameStatePlaying:
		camera := g.player.GetCamera()
		cameraX, cameraY := camera.GetView()

		cameraX += g.screenShakeX
		cameraY += g.screenShakeY

		layers := assets.GetLayersByEnvironment()

		if g.isRealityBroken || g.chaosAtmosphereLevel > 0.7 {
			glitchOffset := g.parallaxOffset * (1.0 + rand.Float64()*0.5)
			atmosphereOffset := g.chaosAtmosphereLevel * 3.0 * math.Sin(g.realityGlitchTimer*6.0)
			totalOffsetX := cameraX + glitchOffset + g.screenDistortionX*0.2 + atmosphereOffset
			totalOffsetY := cameraY + glitchOffset + g.screenDistortionY*0.2 + atmosphereOffset*0.2
			assets.DrawBackgroundLayers(screen, layers, totalOffsetX, totalOffsetY, screenWidth)
		} else {
			distortedX := cameraX + g.screenDistortionX*0.1
			distortedY := cameraY + g.screenDistortionY*0.1
			assets.DrawBackgroundLayers(screen, layers, distortedX, distortedY, screenWidth)
		}

		if assets.DesertTileMap != nil {
			assets.DesertTileMap.Draw(screen, cameraX, cameraY)
		}

		for _, item := range g.specialItems {
			item.Draw(screen, cameraX, cameraY)
		}

		g.globalParticleSystem.Draw(screen, cameraX, cameraY)
		g.madnessParticleSystem.Draw(screen, cameraX, cameraY)

		g.drawPlayerWithCamera(screen, camera)

		if g.colorShiftIntensity > 0.01 {
			limitedIntensity := math.Min(g.colorShiftIntensity, 0.2)
			alpha := uint8(math.Min(16, 16*limitedIntensity))
			overlayColor := color.RGBA{
				uint8(120 * math.Sin(g.realityGlitchTimer*3.0)),
				uint8(120 * math.Sin(g.realityGlitchTimer*2.0)),
				uint8(120 * math.Sin(g.realityGlitchTimer*4.0)),
				alpha,
			}
			vector.DrawFilledRect(screen, 0, 0, float32(screenWidth), float32(screenHeight), overlayColor, false)
		}

		if g.madnessLevel >= 0.8 {
			criticalIntensity := 0.2
			pulseIntensity := 5.

			criticalOverlay := color.RGBA{
				255,
				0,
				0,
				uint8(10 * criticalIntensity * pulseIntensity),
			}
			vector.DrawFilledRect(screen, 0, 0, float32(screenWidth), float32(screenHeight), criticalOverlay, false)
		}

		if g.showCollisionBoxes {
			px, py, pw, ph := g.player.GetBounds()
			screenPX, screenPY := camera.WorldToScreen(px, py)
			vector.StrokeRect(screen, float32(screenPX), float32(screenPY), float32(pw), float32(ph), 1, color.RGBA{0, 255, 0, 255}, false)

			if g.player.IsPerformingAttack() {
				ax, ay, aw, ah := g.player.GetAttackBox()
				screenAX, screenAY := camera.WorldToScreen(ax, ay)
				vector.StrokeRect(screen, float32(screenAX), float32(screenAY), float32(aw), float32(ah), 2, color.RGBA{255, 0, 0, 200}, false)
			}
		}

		g.drawHealthBar(screen)

		if g.currentGlitchMessage != "" && g.messageTimer > 0 {
			messageColor := color.RGBA{
				uint8(255 * (0.5 + 0.5*math.Sin(g.realityGlitchTimer*20.0))),
				uint8(100 * (0.5 + 0.5*math.Sin(g.realityGlitchTimer*15.0))),
				uint8(100 * (0.5 + 0.5*math.Sin(g.realityGlitchTimer*25.0))),
				255,
			}

			textX := 50.0 + rand.Float64()*float64(screenWidth-400)
			textY := 50.0 + rand.Float64()*100.0

			esset.DrawText(screen, g.currentGlitchMessage, textX, textY, assets.FontFaceM, messageColor)
		}

		if g.madnessLevel > 0 {
			madnessText := fmt.Sprintf("MADNESS: %.0f%%", g.madnessLevel*100)

			if g.madnessLevel >= 0.9 {
				madnessText = "⚠️ CRITICAL MADNESS: " + fmt.Sprintf("%.0f%%", g.madnessLevel*100) + " - DEATH IMMINENT! ⚠️"
			}

			madnessColor := color.RGBA{
				uint8(255 * g.madnessLevel),
				uint8(255 * (1.0 - g.madnessLevel)),
				0,
				255,
			}

			if g.madnessLevel >= 0.9 {
				flashIntensity := 0.1 + 0.1*math.Sin(g.realityGlitchTimer*0.5)
				madnessColor = color.RGBA{
					255,
					uint8(100 * flashIntensity),
					uint8(100 * flashIntensity),
					100,
				}
			}

			esset.DrawText(screen, madnessText, float64(screenWidth-350), 10, assets.FontFaceS, madnessColor)
		}

		fps := ebiten.ActualFPS()
		tps := ebiten.ActualTPS()
		fpsTpsText := fmt.Sprintf("FPS: %.0f  TPS: %.0f", fps, tps)
		esset.DrawText(screen, fpsTpsText, 10, 10, assets.FontFaceS, color.RGBA{255, 255, 255, 255})

		if g.endingTriggered {
			g.endingAnimation.Draw(screen)
		}

	case GameStatePaused:
		camera := g.player.GetCamera()
		cameraX, cameraY := camera.GetView()

		layers := assets.GetLayersByEnvironment()
		assets.DrawBackgroundLayers(screen, layers, cameraX, cameraY, screenWidth)

		if assets.DesertTileMap != nil {
			assets.DesertTileMap.Draw(screen, cameraX, cameraY)
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
		camera := g.player.GetCamera()
		cameraX, cameraY := camera.GetView()

		layers := assets.GetLayersByEnvironment()
		assets.DrawBackgroundLayers(screen, layers, cameraX, cameraY, screenWidth)

		if assets.DesertTileMap != nil {
			assets.DesertTileMap.Draw(screen, cameraX, cameraY)
		}

		g.drawPlayerWithCamera(screen, camera)

		vector.DrawFilledRect(screen, 0, 0, float32(screenWidth), float32(screenHeight),
			color.RGBA{0, 0, 0, 120}, false)

		g.menu.Draw(screen)

	case GameStateUnionWin:
		camera := g.player.GetCamera()
		cameraX, cameraY := camera.GetView()

		layers := assets.GetLayersByEnvironment()
		assets.DrawBackgroundLayers(screen, layers, cameraX, cameraY, screenWidth)

		if assets.DesertTileMap != nil {
			assets.DesertTileMap.Draw(screen, cameraX, cameraY)
		}

		for _, item := range g.specialItems {
			if item.Collected {
				item.Draw(screen, cameraX, cameraY)
			}
		}

		g.globalParticleSystem.Draw(screen, cameraX, cameraY)
		g.madnessParticleSystem.Draw(screen, cameraX, cameraY)

		g.drawPlayerWithCamera(screen, camera)

		unionIntensity := 0.2 + 0.1*math.Sin(g.realityGlitchTimer*2.0)
		unionOverlay := color.RGBA{255, 255, 200, uint8(50 * unionIntensity)}
		vector.DrawFilledRect(screen, 0, 0, float32(screenWidth), float32(screenHeight), unionOverlay, false)

		victoryTitle := "🌟 UNION ACHIEVED 🌟"
		titleX := float64(screenWidth)/2 - 200
		titleY := float64(screenHeight)/2 - 100
		esset.DrawText(screen, victoryTitle, titleX, titleY, assets.FontFaceM, color.RGBA{255, 255, 255, 255})

		victoryMessage := "Mind and Matter are One"
		messageX := float64(screenWidth)/2 - 150
		messageY := titleY + 50
		esset.DrawText(screen, victoryMessage, messageX, messageY, assets.FontFaceM, color.RGBA{200, 255, 200, 255})

		statsText := fmt.Sprintf("Chaos Cleansed: %d/%d Items", g.totalItemsCollected, len(g.specialItems))
		statsX := float64(screenWidth)/2 - 120
		statsY := messageY + 80
		esset.DrawText(screen, statsText, statsX, statsY, assets.FontFaceS, color.RGBA{255, 255, 255, 200})

		stabilityText := fmt.Sprintf("World Stability: %.0f%%", g.worldStabilityLevel*100)
		stabilityX := float64(screenWidth)/2 - 100
		stabilityY := statsY + 30
		esset.DrawText(screen, stabilityText, stabilityX, stabilityY, assets.FontFaceS, color.RGBA{255, 255, 255, 200})

		continueText := "Press ESCAPE, ENTER, or SPACE to continue"
		continueX := float64(screenWidth)/2 - 200
		continueY := stabilityY + 60
		continueColor := color.RGBA{255, 255, 255, uint8(150 + 100*math.Sin(g.realityGlitchTimer*4.0))}
		esset.DrawText(screen, continueText, continueX, continueY, assets.FontFaceS, continueColor)
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

	unionBarY := healthBarY + 40
	unionBarWidth := healthBarWidth
	unionBarHeight := float32(15)

	vector.DrawFilledRect(screen, healthBarX, unionBarY+10, unionBarWidth, unionBarHeight, color.RGBA{30, 30, 60, 200}, false)

	unionFillWidth := unionBarWidth * float32(g.unionProgress)
	unionColor := color.RGBA{200, 150, 255, 255}
	if g.unionProgress >= 0.9 {
		pulseIntensity := 0.5 + 0.5*math.Sin(g.realityGlitchTimer*8.0)
		unionColor = color.RGBA{255, 200, 255, uint8(200 + 55*pulseIntensity)}
	}

	vector.DrawFilledRect(screen, healthBarX, unionBarY+10, unionFillWidth, unionBarHeight, unionColor, false)

	unionText := fmt.Sprintf("Union Progress: %.0f%%", g.unionProgress*100)
	if g.unionProgress >= 0.9 {
		unionText = "READY FOR UNION! Find the Union Crystal!"
	}
	esset.DrawText(screen, unionText, float64(healthBarX), float64(unionBarY+unionBarHeight+10), assets.FontFaceS, color.RGBA{200, 150, 255, 255})
}

func (g *Game) restartGame() {
	g.player.X = 100.0
	g.player.Y = g.player.GroundLevel - float64(SpriteHeight)*g.player.Scale
	g.player.VelocityX = 0
	g.player.VelocityY = 0
	g.player.OnGround = true
	g.player.Health = g.player.MaxHealth
	g.player.IsDead = false
	g.player.InvulnTimer = 0

	if g.player.Camera != nil {
		g.player.Camera.X = 0
		g.player.Camera.Y = 0
		g.player.Camera.TargetX = 0
		g.player.Camera.TargetY = 0
	}

	g.madnessLevel = 0
	g.madnessDecayTimer = 0
	g.realityGlitchTimer = 0
	g.colorShiftIntensity = 0
	g.screenShakeX = 0
	g.screenShakeY = 0
	g.isRealityBroken = false
	g.messageTimer = 0
	g.currentGlitchMessage = ""
	g.dimensionSlipTimer = 0
	g.glitchEffectTimer = 0
	g.realityTearTimer = 0
	g.chaosIntensityLevel = 0

	g.worldStabilityLevel = 0
	g.unionProgress = 0
	g.totalItemsCollected = 0
	g.collectedItems = make(map[SpecialItemType]bool)

	for _, item := range g.specialItems {
		item.IsActive = true
		item.Collected = false
		item.Health = item.MaxHealth
		item.HitFlashTimer = 0
		item.IsBeingHit = false
		item.PulsePhase = 0
		item.AuraTimer = 0
		item.LastParticleSpawn = 0
	}

	if g.globalParticleSystem != nil {
		g.globalParticleSystem = NewParticleSystem(200)
	}
	if g.madnessParticleSystem != nil {
		g.madnessParticleSystem = NewParticleSystem(100)
	}

	g.parallaxOffset = 0
}

func (g *Game) updateSchizophrenicEffects(deltaTime float64) {
	g.madnessDecayTimer += deltaTime
	decayRate := 0.05 * (1.0 + g.worldStabilityLevel*2.0)
	if g.madnessDecayTimer > 1.0 {
		g.madnessLevel = math.Max(0, g.madnessLevel-decayRate)
		g.madnessDecayTimer = 0
	}

	stabilityReduction := g.worldStabilityLevel * 0.8
	effectiveMadness := g.madnessLevel * (1.0 - stabilityReduction)

	if effectiveMadness <= 0 {
		g.isRealityBroken = false
		g.currentGlitchMessage = ""
		g.colorShiftIntensity = 0
		g.screenShakeX = 0
		g.screenShakeY = 0
		return
	}

	g.realityGlitchTimer += deltaTime

	g.colorShiftIntensity = math.Sin(g.realityGlitchTimer*5.0) * 0.5
	g.colorShiftIntensity *= effectiveMadness

	shakeIntensity := effectiveMadness * 5.0
	g.screenShakeX = (rand.Float64() - 0.5) * shakeIntensity
	g.screenShakeY = (rand.Float64() - 0.5) * shakeIntensity

	g.messageTimer -= deltaTime
	messageThreshold := 0.3 * (1.0 - g.worldStabilityLevel*0.5)
	if g.messageTimer <= 0 && effectiveMadness > messageThreshold {
		if g.unionProgress > 0.8 {
			unionMessages := []string{
				"THE FRAGMENTS ARE CALLING TO EACH OTHER",
				"UNITY IS WITHIN REACH",
				"THE FINAL PIECE AWAITS",
				"MIND AND MATTER SEEK BALANCE",
			}
			g.currentGlitchMessage = unionMessages[rand.Intn(len(unionMessages))]
		} else if g.worldStabilityLevel > 0.5 {
			stabilityMessages := []string{
				"REALITY IS CRYSTALLIZING...",
				"THE CHAOS SUBSIDES",
				"HARMONY RETURNS TO THE VOID",
				"STABILITY PIERCES THE MADNESS",
			}
			g.currentGlitchMessage = stabilityMessages[rand.Intn(len(stabilityMessages))]
		} else {
			g.currentGlitchMessage = g.glitchMessages[rand.Intn(len(g.glitchMessages))]
		}
		g.messageTimer = 2.0 + rand.Float64()*3.0
	}

	g.dimensionSlipTimer += deltaTime
	if g.dimensionSlipTimer > 10.0 && effectiveMadness > 0.7 {
		g.dimensionSlipTimer = 0
	}

	if rand.Float64() < effectiveMadness*0.05*(1.0-g.worldStabilityLevel) {
		g.isRealityBroken = !g.isRealityBroken
	}
}

func (g *Game) triggerMadness(itemType SpecialItemType) {
	switch itemType {
	case ItemSchizophrenicFragment:
		g.madnessLevel = math.Min(1.0, g.madnessLevel+0.15)
		g.currentGlitchMessage = "FRAGMENT CONSUMED... REALITY FRACTURES"
		g.messageTimer = 4.0
		g.screenShakeX = (rand.Float64() - 0.5) * 3.0
		g.screenShakeY = (rand.Float64() - 0.5) * 3.0

	case ItemRealityGlitch:
		g.madnessLevel = math.Min(1.0, g.madnessLevel+0.25)
		g.currentGlitchMessage = "GLITCH ABSORBED... THE MATRIX BLEEDS"
		g.messageTimer = 5.0
		g.isRealityBroken = true
		g.screenShakeX = (rand.Float64() - 0.5) * 5.0
		g.screenShakeY = (rand.Float64() - 0.5) * 5.0

	case ItemMadnessCore:
		g.madnessLevel = math.Min(0.8, g.madnessLevel+0.35)
		g.currentGlitchMessage = "CORE INTEGRATED... MADNESS SURGES BUT YOU SURVIVE"
		g.messageTimer = 6.0
		g.isRealityBroken = true
		g.screenShakeX = (rand.Float64() - 0.5) * 8.0
		g.screenShakeY = (rand.Float64() - 0.5) * 8.0

	case ItemUnionCrystal:
		g.madnessLevel = 0
		g.currentGlitchMessage = "UNION ACHIEVED... MIND AND MATTER BECOME ONE"
		g.messageTimer = 10.0
		g.worldStabilityLevel = 1.0
		g.unionProgress = 1.0
		g.isRealityBroken = false
		g.triggerUnionEffect()
		if !g.endingTriggered {
			g.endingAnimation.Start()
			g.endingTriggered = true
		}
	default:
	}

	if g.totalItemsCollected > 0 && g.totalItemsCollected%5 == 0 {
		g.madnessLevel = math.Max(0, g.madnessLevel-0.3)
		g.currentGlitchMessage = "WORLD STABILIZES... REALITY BECOMING CLEARER"
		g.messageTimer = 3.0
		g.worldStabilityLevel = math.Min(1.0, g.worldStabilityLevel+0.25)
	}

	g.realityGlitchTimer = 0
	g.colorShiftIntensity = g.madnessLevel

	g.screenShakeX = (rand.Float64() - 0.5) * 4.0
	g.screenShakeY = (rand.Float64() - 0.5) * 4.0
}

func (g *Game) updateProgression(itemType SpecialItemType) {
	g.collectedItems[itemType] = true
	g.totalItemsCollected++

	chaosItemsCollected := 0
	healingItemsCollected := 0
	totalChaosItems := 0
	totalHealingItems := 0

	for _, item := range g.specialItems {
		switch item.ItemType {
		case ItemSchizophrenicFragment, ItemRealityGlitch, ItemMadnessCore:
			totalChaosItems++
			if item.Collected {
				chaosItemsCollected++
			}
		case ItemHarmonyFragment, ItemStabilityCore:
			totalHealingItems++
			if item.Collected {
				healingItemsCollected++
			}
		case ItemUnionCrystal:
			if item.Collected {
				g.unionProgress = 1.0
			}
		}
	}

	chaosRatio := 0.0
	healingRatio := 0.0

	if totalChaosItems > 0 {
		chaosRatio = float64(chaosItemsCollected) / float64(totalChaosItems)
	}

	if totalHealingItems > 0 {
		healingRatio = float64(healingItemsCollected) / float64(totalHealingItems)
	}

	baseStability := healingRatio * 0.8
	chaosReduction := chaosRatio * 0.3

	if chaosRatio > 0.8 {
		chaosReduction *= (2.0 - chaosRatio)
	}

	g.worldStabilityLevel = math.Max(0, math.Min(1.0, baseStability-chaosReduction+chaosRatio*0.2))

	totalItems := len(g.specialItems)
	collectedItems := 0
	hasUnionCrystal := false

	for _, item := range g.specialItems {
		if item.Collected {
			collectedItems++
			if item.ItemType == ItemUnionCrystal {
				hasUnionCrystal = true
			}
		}
	}

	collectionProgress := float64(collectedItems) / float64(totalItems)

	if hasUnionCrystal {
		g.unionProgress = 1.0
		if !g.endingTriggered {
			g.endingAnimation.Start()
			g.endingTriggered = true
		}
	} else if collectedItems >= totalItems-1 {
		g.unionProgress = 0.9
	} else {
		g.unionProgress = collectionProgress * 0.8
	}

	g.worldStabilityLevel = math.Max(0, math.Min(1.0, baseStability-chaosReduction+chaosRatio*0.2+g.unionProgress*0.3))
}

func (g *Game) spawnCollectionEffect(x, y float64, itemType SpecialItemType) {
	switch itemType {
	case ItemSchizophrenicFragment, ItemRealityGlitch, ItemMadnessCore:
		g.globalParticleSystem.SpawnBurst(x, y, ParticleTypeMadness, 10)
		g.madnessParticleSystem.SpawnBurst(x, y, ParticleTypeGlitch, 8)
		g.globalParticleSystem.SpawnBurst(x, y, ParticleTypeDimensionRip, 3)

	case ItemHarmonyFragment:
		g.globalParticleSystem.SpawnBurst(x, y, ParticleTypeHealingLight, 12)
		g.globalParticleSystem.SpawnBurst(x, y, ParticleTypeHarmonyOrb, 4)

	case ItemStabilityCore:
		g.globalParticleSystem.SpawnBurst(x, y, ParticleTypeStabilityWave, 8)
		g.globalParticleSystem.SpawnBurst(x, y, ParticleTypeRealityRestore, 6)

	case ItemUnionCrystal:
		g.globalParticleSystem.SpawnBurst(x, y, ParticleTypeUnionBeam, 8)
		g.globalParticleSystem.SpawnBurst(x, y, ParticleTypeRealityRestore, 5)
		g.globalParticleSystem.SpawnBurst(x, y, ParticleTypeHealingLight, 10)
	}
}

func (g *Game) triggerUnionEffect() {
	playerX, playerY, _, _ := g.player.GetBounds()

	for angle := 0.0; angle < math.Pi*2; angle += math.Pi / 16 {
		for radius := 50.0; radius <= 300.0; radius += 50.0 {
			beamX := playerX + math.Cos(angle)*radius
			beamY := playerY + math.Sin(angle)*radius

			g.globalParticleSystem.SpawnAimedParticle(beamX, beamY, playerX, playerY, ParticleTypeUnionBeam)
			g.globalParticleSystem.SpawnParticle(beamX, beamY, ParticleTypeHealingLight)
		}
	}

	for radius := 50.0; radius <= 500.0; radius += 30.0 {
		for angle := 0.0; angle < math.Pi*2; angle += math.Pi / 20 {
			waveX := playerX + math.Cos(angle)*radius
			waveY := playerY + math.Sin(angle)*radius

			g.globalParticleSystem.SpawnParticle(waveX, waveY, ParticleTypeRealityRestore)

			if radius > 200.0 {
				g.globalParticleSystem.SpawnParticle(waveX, waveY, ParticleTypeUnionBeam)
			}
		}
	}

	for i := 0; i < 50; i++ {
		spiralAngle := float64(i) * 0.5
		spiralRadius := float64(i) * 8.0
		spiralX := playerX + math.Cos(spiralAngle)*spiralRadius
		spiralY := playerY + math.Sin(spiralAngle)*spiralRadius

		g.globalParticleSystem.SpawnParticle(spiralX, spiralY, ParticleTypeHarmonyOrb)
	}

	g.madnessLevel = 0
	g.colorShiftIntensity = 0
	g.chaosAtmosphereLevel = 0
	g.screenShakeX = 0
	g.screenShakeY = 0
	g.isRealityBroken = false
}

func (g *Game) updateChaosAtmosphere(deltaTime float64) {
	g.activeSchizoPoisonCount = 0
	for _, item := range g.specialItems {
		if item.IsActive && !item.Collected {
			if item.ItemType == ItemSchizophrenicFragment ||
				item.ItemType == ItemRealityGlitch ||
				item.ItemType == ItemMadnessCore {
				g.activeSchizoPoisonCount++
			}
		}
	}

	targetAtmosphere := float64(g.activeSchizoPoisonCount) / float64(len(g.specialItems)) * 1.5
	targetAtmosphere = math.Min(1.0, targetAtmosphere)

	if g.chaosAtmosphereLevel < targetAtmosphere {
		g.chaosAtmosphereLevel += deltaTime * 0.3
	} else {
		g.atmosphereDecayTimer += deltaTime
		if g.atmosphereDecayTimer > 0.5 {
			decayRate := 0.1 * (1.0 + g.worldStabilityLevel)
			g.chaosAtmosphereLevel = math.Max(0, g.chaosAtmosphereLevel-decayRate)
			g.atmosphereDecayTimer = 0
		}
	}

	if g.chaosAtmosphereLevel > g.maxAtmosphereLevel {
		g.maxAtmosphereLevel = g.chaosAtmosphereLevel
	}

	if g.chaosAtmosphereLevel > 0.5 {
		distortionStrength := g.chaosAtmosphereLevel * 1.0
		g.screenDistortionX = math.Sin(g.realityGlitchTimer*3.0) * distortionStrength
		g.screenDistortionY = math.Cos(g.realityGlitchTimer*4.0) * distortionStrength
	} else {
		g.screenDistortionX = 0
		g.screenDistortionY = 0
	}

	g.atmosphereParticleTimer += deltaTime
	particleSpawnRate := g.chaosAtmosphereLevel * 1.0

	if g.atmosphereParticleTimer > (0.8-particleSpawnRate*0.3) && g.chaosAtmosphereLevel > 0.2 {
		playerX, playerY, _, _ := g.player.GetBounds()

		for i := 0; i < int(g.chaosAtmosphereLevel*2)+1; i++ {
			particleX := playerX + (rand.Float64()-0.5)*800
			particleY := playerY + (rand.Float64()-0.5)*600

			if g.chaosAtmosphereLevel > 0.8 {
				g.madnessParticleSystem.SpawnParticle(particleX, particleY, ParticleTypeMadness)
			} else if g.chaosAtmosphereLevel > 0.5 {
				g.madnessParticleSystem.SpawnParticle(particleX, particleY, ParticleTypeHallucinationSpark)
			} else {
				g.madnessParticleSystem.SpawnParticle(particleX, particleY, ParticleTypeMadness)
			}
		}

		g.atmosphereParticleTimer = 0
	}

	if g.chaosAtmosphereLevel > 0.8 {
		if rand.Float64() < g.chaosAtmosphereLevel*0.005 {
			g.isRealityBroken = !g.isRealityBroken
		}

		g.colorShiftIntensity = math.Max(g.colorShiftIntensity, g.chaosAtmosphereLevel*0.3)
	}
}

func (g *Game) checkProximityDamage(deltaTime float64) {
	playerX, playerY, playerW, playerH := g.player.GetBounds()
	playerCenterX := playerX + playerW/2
	playerCenterY := playerY + playerH/2

	for _, item := range g.specialItems {
		if !item.IsActive || item.Collected {
			continue
		}

		itemCenterX := item.X + item.Width/2
		itemCenterY := item.Y + item.Height/2
		distance := math.Sqrt(math.Pow(playerCenterX-itemCenterX, 2) + math.Pow(playerCenterY-itemCenterY, 2))

		var damageRadius float64
		var damageAmount int

		switch item.ItemType {
		case ItemSchizophrenicFragment:
			damageRadius = 30.0
			damageAmount = 2
		case ItemRealityGlitch:
			damageRadius = 40.0
			damageAmount = 3
		case ItemMadnessCore:
			damageRadius = 60.0
			damageAmount = 5
		case ItemUnionCrystal:
			damageRadius = 25.0
			damageAmount = 1
		default:
			continue
		}

		if distance < damageRadius {
			g.proximityDamageTimer += deltaTime
			if g.proximityDamageTimer >= 2.0 {
				g.player.TakeDamage(damageAmount)
				g.proximityDamageTimer = 0

				g.screenShakeX += (rand.Float64() - 0.5) * 5.0
				g.screenShakeY += (rand.Float64() - 0.5) * 5.0
				break
			}
		}
	}
}

func (g *Game) updateDifficultyAndPressure(deltaTime float64) {
	g.survivalTimer += deltaTime

	g.difficultyModifier = 1.0 + (g.survivalTimer/120.0)*0.5

	g.healthDecayTimer += deltaTime
	g.healthDecayRate = 0.1 + g.madnessLevel*0.3 + (g.survivalTimer/60.0)*0.05

	if g.healthDecayTimer >= 1.0 && g.madnessLevel > 0.3 {
		g.healthDecayTimer = 0
		decayAmount := int(g.healthDecayRate * g.difficultyModifier)
		if decayAmount < 1 {
			decayAmount = 1
		}
		g.player.TakeDamage(decayAmount)
		g.lastDamageTime = g.survivalTimer
	}

	for _, item := range g.specialItems {
		if !item.IsActive || item.Collected {
			continue
		}

		if item.HitFlashTimer <= 0 && item.Health < item.MaxHealth {
			item.Health = int(math.Min(float64(item.MaxHealth), float64(item.Health)+deltaTime*g.difficultyModifier*0.5))
		}
	}
}
