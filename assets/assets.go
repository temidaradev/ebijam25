package assets

import (
	"embed"
	"image"
	_ "image/jpeg" // Add JPEG support
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/temidaradev/esset/v2"
)

// Display configuration constants
const (
	WindowWidth800  = 1280
	WindowHeight600 = 720
)

// ScaleMode represents different scaling modes
type ScaleMode int

const (
	ScaleModeStretch ScaleMode = iota // Stretch to fit, may distort aspect ratio
	ScaleModeAspect                   // Maintain aspect ratio, may have letterboxing
	ScaleModePixel                    // Pixel-perfect scaling (integer multiples)
)

// DisplayConfig holds display and scaling configuration
type DisplayConfig struct {
	WindowWidth  int
	WindowHeight int
	GameWidth    int
	GameHeight   int
	ScaleX       float64
	ScaleY       float64
	OffsetX      float64
	OffsetY      float64
	Mode         ScaleMode
	IsFullscreen bool
}

// Global display configuration
var CurrentDisplayConfig *DisplayConfig

// Initialize display configurations
func init() {
	CurrentDisplayConfig = NewDisplayConfig(WindowWidth800, WindowHeight600, ScaleModeAspect, false)
}

// NewDisplayConfig creates a new display configuration
func NewDisplayConfig(windowWidth, windowHeight int, mode ScaleMode, fullscreen bool) *DisplayConfig {
	config := &DisplayConfig{
		WindowWidth:  windowWidth,
		WindowHeight: windowHeight,
		GameWidth:    WindowWidth800,  // Always use 1280x720 as base game resolution
		GameHeight:   WindowHeight600, // Always use 1280x720 as base game resolution
		Mode:         mode,
		IsFullscreen: fullscreen,
	}

	config.calculateScaling()
	return config
}

// calculateScaling computes the scaling factors and offsets
func (dc *DisplayConfig) calculateScaling() {
	windowAspect := float64(dc.WindowWidth) / float64(dc.WindowHeight)
	gameAspect := float64(dc.GameWidth) / float64(dc.GameHeight)

	switch dc.Mode {
	case ScaleModeStretch:
		// Stretch to fit window, ignoring aspect ratio
		dc.ScaleX = float64(dc.WindowWidth) / float64(dc.GameWidth)
		dc.ScaleY = float64(dc.WindowHeight) / float64(dc.GameHeight)
		dc.OffsetX = 0
		dc.OffsetY = 0

	case ScaleModeAspect:
		// Maintain aspect ratio with letterboxing/pillarboxing
		if windowAspect > gameAspect {
			// Window is wider than game - pillarbox (black bars on sides)
			dc.ScaleY = float64(dc.WindowHeight) / float64(dc.GameHeight)
			dc.ScaleX = dc.ScaleY
			scaledWidth := float64(dc.GameWidth) * dc.ScaleX
			dc.OffsetX = (float64(dc.WindowWidth) - scaledWidth) / 2
			dc.OffsetY = 0
		} else {
			// Window is taller than game - letterbox (black bars on top/bottom)
			dc.ScaleX = float64(dc.WindowWidth) / float64(dc.GameWidth)
			dc.ScaleY = dc.ScaleX
			scaledHeight := float64(dc.GameHeight) * dc.ScaleY
			dc.OffsetX = 0
			dc.OffsetY = (float64(dc.WindowHeight) - scaledHeight) / 2
		}

	case ScaleModePixel:
		// Pixel-perfect scaling using integer multiples
		scaleX := float64(dc.WindowWidth) / float64(dc.GameWidth)
		scaleY := float64(dc.WindowHeight) / float64(dc.GameHeight)

		// Use the smaller scale factor and make it an integer
		scale := scaleX
		if scaleY < scaleX {
			scale = scaleY
		}

		// Floor to get integer scaling
		if scale < 1.0 {
			scale = 1.0
		} else {
			scale = float64(int(scale))
		}

		dc.ScaleX = scale
		dc.ScaleY = scale

		scaledWidth := float64(dc.GameWidth) * scale
		scaledHeight := float64(dc.GameHeight) * scale

		dc.OffsetX = (float64(dc.WindowWidth) - scaledWidth) / 2
		dc.OffsetY = (float64(dc.WindowHeight) - scaledHeight) / 2
	}
}

// UpdateDisplayConfig updates the current display configuration
func UpdateDisplayConfig(windowWidth, windowHeight int, mode ScaleMode, fullscreen bool) {
	CurrentDisplayConfig = NewDisplayConfig(windowWidth, windowHeight, mode, fullscreen)
}

// GetScaledDrawOptions returns DrawImageOptions with proper scaling applied
func (dc *DisplayConfig) GetScaledDrawOptions() *ebiten.DrawImageOptions {
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(dc.ScaleX, dc.ScaleY)
	opts.GeoM.Translate(dc.OffsetX, dc.OffsetY)
	return opts
}

// GetScaledPosition converts game coordinates to screen coordinates
func (dc *DisplayConfig) GetScaledPosition(gameX, gameY float64) (screenX, screenY float64) {
	screenX = gameX*dc.ScaleX + dc.OffsetX
	screenY = gameY*dc.ScaleY + dc.OffsetY
	return
}

// GetGamePosition converts screen coordinates to game coordinates
func (dc *DisplayConfig) GetGamePosition(screenX, screenY float64) (gameX, gameY float64) {
	gameX = (screenX - dc.OffsetX) / dc.ScaleX
	gameY = (screenY - dc.OffsetY) / dc.ScaleY
	return
}

// GetScaledSize returns the scaled size of the game area
func (dc *DisplayConfig) GetScaledSize() (width, height float64) {
	width = float64(dc.GameWidth) * dc.ScaleX
	height = float64(dc.GameHeight) * dc.ScaleY
	return
}

// DrawWithScaling draws an image with the current display scaling applied
func (dc *DisplayConfig) DrawWithScaling(screen *ebiten.Image, src *ebiten.Image, gameX, gameY float64) {
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(gameX, gameY)
	opts.GeoM.Scale(dc.ScaleX, dc.ScaleY)
	opts.GeoM.Translate(dc.OffsetX, dc.OffsetY)
	screen.DrawImage(src, opts)
}

// CreateScaledScreen creates a virtual screen for game rendering
func (dc *DisplayConfig) CreateScaledScreen() *ebiten.Image {
	return ebiten.NewImage(dc.GameWidth, dc.GameHeight)
}

// DrawScaledScreen draws the virtual screen to the actual screen with scaling
func (dc *DisplayConfig) DrawScaledScreen(screen *ebiten.Image, virtualScreen *ebiten.Image) {
	opts := dc.GetScaledDrawOptions()
	screen.DrawImage(virtualScreen, opts)
}

//go:embed *
var assets embed.FS

//go:embed font/PublicPixel.ttf
var Font []byte

var FontFaceS text.Face
var FontFaceM text.Face

var (
	// Load the actual adventurer sprite sheet
	CharacterSpritesheet = esset.GetAsset(assets, "images/sprites/adventurer-Sheet.png")

	// Desert background layers (back to front, z-depth from negative to positive)
	DesertBackground1 = esset.GetAsset(assets, "images/backgrounds/desert/background1.png") // Far background (z -6)
	DesertBackground2 = esset.GetAsset(assets, "images/backgrounds/desert/background2.png") // Mid background (z -4)
	DesertBackground3 = esset.GetAsset(assets, "images/backgrounds/desert/background3.png") // Near background (z -2)

	// Desert clouds (various depths for parallax effect)
	DesertCloud1 = esset.GetAsset(assets, "images/backgrounds/desert/cloud1.png") // Far clouds (z -5)
	DesertCloud2 = esset.GetAsset(assets, "images/backgrounds/desert/cloud2.png") // Far clouds (z -5)
	DesertCloud3 = esset.GetAsset(assets, "images/backgrounds/desert/cloud3.png") // Mid clouds (z -3)
	DesertCloud4 = esset.GetAsset(assets, "images/backgrounds/desert/cloud4.png") // Mid clouds (z -3)
	DesertCloud5 = esset.GetAsset(assets, "images/backgrounds/desert/cloud5.png") // Near clouds (z -1)
	DesertCloud6 = esset.GetAsset(assets, "images/backgrounds/desert/cloud6.png") // Near clouds (z -1)
	DesertCloud7 = esset.GetAsset(assets, "images/backgrounds/desert/cloud7.png") // Foreground clouds (z 0)
	DesertCloud8 = esset.GetAsset(assets, "images/backgrounds/desert/cloud8.png") // Foreground clouds (z 0)

	// Forest background layers (back to front)
	ForestSky      = esset.GetAsset(assets, "images/backgrounds/forest/sky.png")       // Far background (z -6)
	ForestSkyCloud = esset.GetAsset(assets, "images/backgrounds/forest/sky_cloud.png") // Sky with clouds (z -5)
	ForestMountain = esset.GetAsset(assets, "images/backgrounds/forest/mountain2.png") // Mountains (z -4)
	ForestCloud    = esset.GetAsset(assets, "images/backgrounds/forest/cloud.png")     // Mid clouds (z -3)
	ForestPine1    = esset.GetAsset(assets, "images/backgrounds/forest/pine1.png")     // Far trees (z -2)
	ForestPine2    = esset.GetAsset(assets, "images/backgrounds/forest/pine2.png")     // Near trees (z -1)

	// Mountains background layers (back to front)
	MountainsSky         = esset.GetAsset(assets, "images/backgrounds/mountains/sky.png")               // Far background (z -6)
	MountainsCloudsBg    = esset.GetAsset(assets, "images/backgrounds/mountains/clouds_bg.png")         // Background clouds (z -5)
	MountainsGlacial     = esset.GetAsset(assets, "images/backgrounds/mountains/glacial_mountains.png") // Mountains (z -4)
	MountainsCloudsMg3   = esset.GetAsset(assets, "images/backgrounds/mountains/clouds_mg_3.png")       // Mid-ground clouds 3 (z -3)
	MountainsCloudsMg2   = esset.GetAsset(assets, "images/backgrounds/mountains/clouds_mg_2.png")       // Mid-ground clouds 2 (z -2)
	MountainsCloudsMg1   = esset.GetAsset(assets, "images/backgrounds/mountains/clouds_mg_1.png")       // Mid-ground clouds 1 (z -1)
	MountainsCloudLonely = esset.GetAsset(assets, "images/backgrounds/mountains/cloud_lonely.png")      // Foreground cloud (z 0)
)

// Initialize simple character animations
func InitCharacterAnimations() *SimpleAnimationManager {
	// Create the simple animation manager
	animManager := NewSimpleAnimationManager(CharacterSpritesheet, 50, 37)

	// Add animations with faster, more responsive timings
	animManager.AddAnimation("idle", 0, 3, 0.12, true)    // idle: frames 0-3, faster idle
	animManager.AddAnimation("run", 8, 13, 0.06, true)    // run: frames 8-13, much faster run
	animManager.AddAnimation("jump", 14, 17, 0.08, false) // jump: frames 14-17, snappy jump
	animManager.AddAnimation("fall", 22, 23, 0.1, true)   // fall: frames 22-23, faster fall
	animManager.AddAnimation("walk", 155, 160, 0.18, true)

	// Set default animation
	animManager.SetAnimation("idle")

	return animManager
}

// SimpleAnimationManager manages animations using simple frame-based approach
type SimpleAnimationManager struct {
	spritesheet    *ebiten.Image
	animations     map[string]*SimpleAnimation
	currentAnim    string
	previousAnim   string
	currentFrame   int
	frameTimer     float64
	frameWidth     int
	frameHeight    int
	animationSpeed float64 // Speed multiplier for all animations
}

// SimpleAnimation represents a single animation sequence
type SimpleAnimation struct {
	name      string
	frames    []image.Rectangle
	durations []float64
	loop      bool
}

// NewSimpleAnimationManager creates a new simple animation manager
func NewSimpleAnimationManager(spritesheet *ebiten.Image, frameWidth, frameHeight int) *SimpleAnimationManager {
	return &SimpleAnimationManager{
		spritesheet:    spritesheet,
		animations:     make(map[string]*SimpleAnimation),
		currentAnim:    "",
		previousAnim:   "",
		currentFrame:   0,
		frameTimer:     0,
		frameWidth:     frameWidth,
		frameHeight:    frameHeight,
		animationSpeed: 1.0, // Default speed
	}
}

// AddAnimation adds a new animation sequence
func (sam *SimpleAnimationManager) AddAnimation(name string, startFrame, endFrame int, frameDuration float64, loop bool) {
	frames := make([]image.Rectangle, 0, endFrame-startFrame+1)
	durations := make([]float64, 0, endFrame-startFrame+1)

	for i := startFrame; i <= endFrame; i++ {
		// Calculate frame position (assuming horizontal sprite strip)
		x := (i % (sam.spritesheet.Bounds().Dx() / sam.frameWidth)) * sam.frameWidth
		y := (i / (sam.spritesheet.Bounds().Dx() / sam.frameWidth)) * sam.frameHeight

		frames = append(frames, image.Rect(x, y, x+sam.frameWidth, y+sam.frameHeight))
		durations = append(durations, frameDuration)
	}

	sam.animations[name] = &SimpleAnimation{
		name:      name,
		frames:    frames,
		durations: durations,
		loop:      loop,
	}
}

// SetAnimation sets the current animation with immediate transition
func (sam *SimpleAnimationManager) SetAnimation(name string) {
	if name != sam.currentAnim {
		sam.previousAnim = sam.currentAnim
		sam.currentAnim = name
		sam.currentFrame = 0
		sam.frameTimer = 0 // Immediate transition, no blending
	}
}

// SetAnimationSpeed sets the speed multiplier for animations
func (sam *SimpleAnimationManager) SetAnimationSpeed(speed float64) {
	sam.animationSpeed = speed
}

// Update updates the current animation with improved timing
func (sam *SimpleAnimationManager) Update(dt float64) {
	if sam.currentAnim == "" {
		return
	}

	anim, exists := sam.animations[sam.currentAnim]
	if !exists || len(anim.frames) == 0 {
		return
	}

	// Apply animation speed multiplier
	sam.frameTimer += dt * sam.animationSpeed

	if sam.currentFrame < len(anim.durations) && sam.frameTimer >= anim.durations[sam.currentFrame] {
		sam.frameTimer = 0
		sam.currentFrame++

		if sam.currentFrame >= len(anim.frames) {
			if anim.loop {
				sam.currentFrame = 0
			} else {
				sam.currentFrame = len(anim.frames) - 1
			}
		}
	}
}

// DrawWithOptions draws the current animation frame with custom options
func (sam *SimpleAnimationManager) DrawWithOptions(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	if sam.currentAnim == "" {
		return
	}

	anim, exists := sam.animations[sam.currentAnim]
	if !exists || len(anim.frames) == 0 {
		return
	}

	frameRect := anim.frames[sam.currentFrame]
	sprite := sam.spritesheet.SubImage(frameRect).(*ebiten.Image)
	screen.DrawImage(sprite, op)
}

// Original animation system (kept for backward compatibility)
type SpriteAnimation struct {
	Spritesheet  *ebiten.Image
	Frames       []image.Rectangle
	FrameWidth   int
	FrameHeight  int
	FrameTime    float64
	CurrentTime  float64
	CurrentFrame int
	Loop         bool
	Playing      bool
}

func NewSpriteAnimation(spritesheet *ebiten.Image, frameWidth, frameHeight int, frameTime float64, loop bool) *SpriteAnimation {
	return &SpriteAnimation{
		Spritesheet:  spritesheet,
		FrameWidth:   frameWidth,
		FrameHeight:  frameHeight,
		FrameTime:    frameTime,
		Loop:         loop,
		CurrentFrame: 0,
		CurrentTime:  0,
		Playing:      true,
	}
}

// AddFrame adds a single frame at the specified position
func (a *SpriteAnimation) AddFrame(row, col int) {
	x := col * a.FrameWidth
	y := row * a.FrameHeight
	rect := image.Rect(x, y, x+a.FrameWidth, y+a.FrameHeight)
	a.Frames = append(a.Frames, rect)
}

func (a *SpriteAnimation) Update(dt float64) {
	if !a.Playing || len(a.Frames) <= 1 {
		return
	}

	a.CurrentTime += dt
	if a.CurrentTime >= a.FrameTime {
		a.CurrentTime -= a.FrameTime
		a.CurrentFrame++

		if a.CurrentFrame >= len(a.Frames) {
			if a.Loop {
				a.CurrentFrame = 0
			} else {
				a.CurrentFrame = len(a.Frames) - 1
				a.Playing = false
			}
		}
	}
}

// DrawWithOptions draws the animation with custom draw options
func (a *SpriteAnimation) DrawWithOptions(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	if len(a.Frames) == 0 {
		return
	}

	frame := a.Frames[a.CurrentFrame]
	screen.DrawImage(a.Spritesheet.SubImage(frame).(*ebiten.Image), op)
}

func (a *SpriteAnimation) IsFinished() bool {
	return !a.Loop && a.CurrentFrame == len(a.Frames)-1
}

func (a *SpriteAnimation) Reset() {
	a.CurrentFrame = 0
	a.CurrentTime = 0
	a.Playing = true
}

// AnimationState represents different animation states
type AnimationState string

const (
	AnimationIdle   AnimationState = "idle"
	AnimationWalk   AnimationState = "walk"
	AnimationRun    AnimationState = "run"
	AnimationJump   AnimationState = "jump"
	AnimationAttack AnimationState = "attack"
	AnimationDamage AnimationState = "damage"
	AnimationDeath  AnimationState = "death"
)

// AnimationManager manages multiple animations for a sprite
type AnimationManager struct {
	animations    map[AnimationState]*SpriteAnimation
	currentState  AnimationState
	previousState AnimationState
}

// NewAnimationManager creates a new animation manager
func NewAnimationManager() *AnimationManager {
	return &AnimationManager{
		animations:    make(map[AnimationState]*SpriteAnimation),
		currentState:  AnimationIdle,
		previousState: AnimationIdle,
	}
}

// AddAnimation adds an animation for a specific state
func (am *AnimationManager) AddAnimation(state AnimationState, animation *SpriteAnimation) {
	am.animations[state] = animation
}

// SetState changes the current animation state
func (am *AnimationManager) SetState(state AnimationState) {
	if am.currentState != state {
		if currentAnim, exists := am.animations[am.currentState]; exists {
			currentAnim.Reset()
		}
		am.previousState = am.currentState
		am.currentState = state
		if newAnim, exists := am.animations[state]; exists {
			newAnim.Reset()
			newAnim.Play()
		}
	}
}

// Update updates the current animation
func (am *AnimationManager) Update(dt float64) {
	if anim, exists := am.animations[am.currentState]; exists {
		anim.Update(dt)
	}
}

// DrawWithOptions draws the current animation with custom options
func (am *AnimationManager) DrawWithOptions(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	if anim, exists := am.animations[am.currentState]; exists {
		anim.DrawWithOptions(screen, op)
	}
}

// Play starts or resumes the animation
func (a *SpriteAnimation) Play() {
	a.Playing = true
}

// SpriteSheet represents a spritesheet with multiple sprites
type SpriteSheet struct {
	Image        *ebiten.Image
	SpriteWidth  int
	SpriteHeight int
	Columns      int
	Rows         int
}

// NewSpriteSheet creates a new spritesheet
func NewSpriteSheet(image *ebiten.Image, spriteWidth, spriteHeight int) *SpriteSheet {
	columns := image.Bounds().Dx() / spriteWidth
	rows := image.Bounds().Dy() / spriteHeight

	return &SpriteSheet{
		Image:        image,
		SpriteWidth:  spriteWidth,
		SpriteHeight: spriteHeight,
		Columns:      columns,
		Rows:         rows,
	}
}

// GetCurrentAnimation returns the name of the current animation
func (sam *SimpleAnimationManager) GetCurrentAnimation() string {
	return sam.currentAnim
}

// GetCurrentFrame returns the current frame index
func (sam *SimpleAnimationManager) GetCurrentFrame() int {
	return sam.currentFrame
}

// IsAnimationFinished returns true if the current non-looping animation has finished
func (sam *SimpleAnimationManager) IsAnimationFinished() bool {
	if sam.currentAnim == "" {
		return true
	}

	anim, exists := sam.animations[sam.currentAnim]
	if !exists {
		return true
	}

	return !anim.loop && sam.currentFrame >= len(anim.frames)-1
}

type BackgroundLayer struct {
	Image     *ebiten.Image
	Name      string
	ParallaxX float64
	ParallaxY float64
	ZDepth    int
	OffsetX   float64
	OffsetY   float64
	RepeatX   bool
	RepeatY   bool
	ScaleX    float64
	ScaleY    float64
}

func DesertLayers() []BackgroundLayer {
	return []BackgroundLayer{
		{DesertBackground1, "desert_bg1", 0.1, 0.05, -6, 0, 0, true, false, 1.5, 1.5},
		{DesertBackground2, "desert_bg2", 0.25, 0.1, -4, 0, 0, true, false, 1.5, 1.5},
		{DesertBackground3, "desert_bg3", 0.4, 0.15, -2, 0, 0, true, false, 1.5, 1.5},
	}
}

func ForestLayers() []BackgroundLayer {
	return []BackgroundLayer{
		{ForestSky, "forest_sky", 0.05, 0.02, -6, 0, 0, true, false, 2, 2},
		{ForestSkyCloud, "forest_sky_cloud", 0.1, 0.05, -5, 0, 10, true, false, 3.2, 2},
		{ForestMountain, "forest_mountain", 0.2, 0.08, -4, 0, 60, true, false, 2, 2},
		{ForestCloud, "forest_cloud", 0.3, 0.12, -3, 160, 48, true, false, 2, 2},
		{ForestPine1, "forest_pine1", 0.5, 0.2, -2, 0, 125, true, false, 2, 2},
		{ForestPine2, "forest_pine2", 0.8, 0.4, -1, 0, 160, true, false, 2, 2},
	}
}

func MountainsLayers() []BackgroundLayer {
	return []BackgroundLayer{
		{MountainsSky, "mountains_sky", 0.05, 0.02, -6, 0, 0, true, false, 2, 2},
		{MountainsCloudsBg, "mountains_clouds_bg", 0.1, 0.05, -5, 0, 10, true, false, 2, 2},
		{MountainsGlacial, "mountains_glacial", 0.2, 0.08, -4, 0, 80, true, false, 2, 2},
		{MountainsCloudsMg3, "mountains_clouds_mg3", 0.25, 0.1, -3, 50, 120, true, false, 2, 2},
		{MountainsCloudsMg2, "mountains_clouds_mg2", 0.4, 0.15, -2, 208, 140, true, false, 2, 2},
		{MountainsCloudsMg1, "mountains_clouds_mg1", 0.6, 0.25, -1, 128, 145, true, false, 2, 2},
		{MountainsCloudLonely, "mountains_cloud_lonely", 0.8, 0.35, 0, 320, 112, false, false, 2, 2},
	}
}

func GetLayersByEnvironment(environment string) []BackgroundLayer {
	switch environment {
	case "desert":
		return DesertLayers()
	case "forest":
		return ForestLayers()
	case "mountains":
		return MountainsLayers()
	default:
		return DesertLayers()
	}
}

func DrawBackgroundLayers(screen *ebiten.Image, layers []BackgroundLayer, cameraX, cameraY float64, screenWidth, screenHeight int) {
	DrawBackgroundLayersScaled(screen, layers, cameraX, cameraY, screenWidth, screenHeight, nil)
}

// DrawBackgroundLayersScaled draws background layers with display scaling applied
func DrawBackgroundLayersScaled(screen *ebiten.Image, layers []BackgroundLayer, cameraX, cameraY float64, screenWidth, screenHeight int, displayConfig *DisplayConfig) {
	for _, layer := range layers {
		if layer.Image == nil {
			continue
		}

		parallaxOffsetX := cameraX * layer.ParallaxX
		parallaxOffsetY := cameraY * layer.ParallaxY

		finalX := layer.OffsetX - parallaxOffsetX
		finalY := layer.OffsetY - parallaxOffsetY

		opts := &ebiten.DrawImageOptions{}

		if layer.RepeatX {
			imgWidth := float64(layer.Image.Bounds().Dx())
			startX := finalX

			// Adjust for display scaling if provided
			effectiveScreenWidth := float64(screenWidth)
			if displayConfig != nil {
				effectiveScreenWidth = float64(displayConfig.GameWidth)
			}

			for startX > 0 {
				startX -= imgWidth
			}

			for x := startX; x < effectiveScreenWidth; x += imgWidth {
				opts.GeoM.Reset()
				opts.GeoM.Translate(x, finalY)

				// Apply per-layer scaling first
				opts.GeoM.Scale(layer.ScaleX, layer.ScaleY)

				// Apply display scaling if provided
				if displayConfig != nil {
					opts.GeoM.Scale(displayConfig.ScaleX, displayConfig.ScaleY)
					opts.GeoM.Translate(displayConfig.OffsetX, displayConfig.OffsetY)
				}

				screen.DrawImage(layer.Image, opts)
			}
		} else {
			opts.GeoM.Reset()
			opts.GeoM.Translate(finalX, finalY)

			// Apply per-layer scaling first
			opts.GeoM.Scale(layer.ScaleX, layer.ScaleY)

			// Apply display scaling if provided
			if displayConfig != nil {
				opts.GeoM.Scale(displayConfig.ScaleX, displayConfig.ScaleY)
				opts.GeoM.Translate(displayConfig.OffsetX, displayConfig.OffsetY)
			}

			screen.DrawImage(layer.Image, opts)
		}
	}
}

// ScaleBackgroundLayer scales a background layer for different resolutions
func ScaleBackgroundLayer(layer *BackgroundLayer, scaleX, scaleY float64) BackgroundLayer {
	scaledLayer := *layer // Copy the layer

	// Scale the offsets
	scaledLayer.OffsetX *= scaleX
	scaledLayer.OffsetY *= scaleY

	// Adjust parallax factors slightly for different scales
	// This helps maintain the parallax effect across different resolutions
	parallaxScale := (scaleX + scaleY) / 2.0
	if parallaxScale != 1.0 {
		scaledLayer.ParallaxX *= parallaxScale
		scaledLayer.ParallaxY *= parallaxScale
	}

	return scaledLayer
}

// GetScaledLayersByEnvironment returns background layers scaled for current display config
func GetScaledLayersByEnvironment(environment string, displayConfig *DisplayConfig) []BackgroundLayer {
	baseLayers := GetLayersByEnvironment(environment)

	if displayConfig == nil {
		return baseLayers
	}

	scaledLayers := make([]BackgroundLayer, len(baseLayers))
	for i, layer := range baseLayers {
		scaledLayers[i] = ScaleBackgroundLayer(&layer, displayConfig.ScaleX, displayConfig.ScaleY)
	}

	return scaledLayers
}
