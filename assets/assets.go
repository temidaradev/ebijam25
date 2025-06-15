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

// BackgroundLayer represents a single background layer with its properties
type BackgroundLayer struct {
	Image     *ebiten.Image
	Name      string
	ParallaxX float64 // Horizontal parallax multiplier (0.0 = static, 1.0 = moves with camera)
	ParallaxY float64 // Vertical parallax multiplier
	ZDepth    int     // Z-depth for layering (negative = background, positive = foreground)
	OffsetX   float64 // Additional horizontal offset
	OffsetY   float64 // Additional vertical offset
	RepeatX   bool    // Whether to repeat horizontally
	RepeatY   bool    // Whether to repeat vertically
}

// DesertLayers returns all desert background layers properly ordered
func DesertLayers() []BackgroundLayer {
	return []BackgroundLayer{
		{DesertBackground1, "desert_bg1", 0.1, 0.05, -6, 0, 0, true, false},
		{DesertCloud1, "desert_cloud1", 0.15, 0.08, -5, 50, 20, true, false},
		{DesertCloud2, "desert_cloud2", 0.18, 0.08, -5, 200, 30, true, false},
		{DesertBackground2, "desert_bg2", 0.25, 0.1, -4, 0, 0, true, false},
		{DesertCloud3, "desert_cloud3", 0.3, 0.12, -3, 100, 40, true, false},
		{DesertCloud4, "desert_cloud4", 0.32, 0.12, -3, 300, 25, true, false},
		{DesertBackground3, "desert_bg3", 0.4, 0.15, -2, 0, 0, true, false},
		{DesertCloud5, "desert_cloud5", 0.5, 0.2, -1, 150, 35, true, false},
		{DesertCloud6, "desert_cloud6", 0.52, 0.2, -1, 350, 45, true, false},
		{DesertCloud7, "desert_cloud7", 0.7, 0.3, 0, 80, 50, true, false},
		{DesertCloud8, "desert_cloud8", 0.72, 0.3, 0, 280, 40, true, false},
	}
}

// ForestLayers returns all forest background layers properly ordered
func ForestLayers() []BackgroundLayer {
	return []BackgroundLayer{
		{ForestSky, "forest_sky", 0.05, 0.02, -6, 0, 0, true, false},
		{ForestSkyCloud, "forest_sky_cloud", 0.1, 0.05, -5, 0, 0, true, false},
		{ForestMountain, "forest_mountain", 0.2, 0.08, -4, 0, 50, true, false},
		{ForestCloud, "forest_cloud", 0.3, 0.12, -3, 100, 30, true, false},
		{ForestPine1, "forest_pine1", 0.5, 0.2, -2, 0, 100, true, false},
		{ForestPine2, "forest_pine2", 0.8, 0.4, -1, 0, 150, true, false},
	}
}

// MountainsLayers returns all mountain background layers properly ordered
func MountainsLayers() []BackgroundLayer {
	return []BackgroundLayer{
		{MountainsSky, "mountains_sky", 0.05, 0.02, -6, 0, 0, true, false},
		{MountainsCloudsBg, "mountains_clouds_bg", 0.1, 0.05, -5, 0, 20, true, false},
		{MountainsGlacial, "mountains_glacial", 0.15, 0.08, -4, 0, 80, true, false},
		{MountainsCloudsMg3, "mountains_clouds_mg3", 0.25, 0.1, -3, 50, 40, true, false},
		{MountainsCloudsMg2, "mountains_clouds_mg2", 0.4, 0.15, -2, 120, 60, true, false},
		{MountainsCloudsMg1, "mountains_clouds_mg1", 0.6, 0.25, -1, 80, 80, true, false},
		{MountainsCloudLonely, "mountains_cloud_lonely", 0.8, 0.35, 0, 200, 100, false, false},
	}
}

// DrawBackgroundLayers draws a set of background layers with parallax scrolling
func DrawBackgroundLayers(screen *ebiten.Image, layers []BackgroundLayer, cameraX, cameraY float64, screenWidth, screenHeight int) {
	for _, layer := range layers {
		if layer.Image == nil {
			continue
		}

		// Calculate parallax offset
		parallaxOffsetX := cameraX * layer.ParallaxX
		parallaxOffsetY := cameraY * layer.ParallaxY

		// Calculate final position
		finalX := layer.OffsetX - parallaxOffsetX
		finalY := layer.OffsetY - parallaxOffsetY

		opts := &ebiten.DrawImageOptions{}

		if layer.RepeatX {
			// Handle horizontal repetition
			imgWidth := float64(layer.Image.Bounds().Dx())
			startX := finalX

			// Adjust starting position to avoid gaps
			for startX > 0 {
				startX -= imgWidth
			}

			for x := startX; x < float64(screenWidth); x += imgWidth {
				opts.GeoM.Reset()
				opts.GeoM.Translate(x, finalY)
				screen.DrawImage(layer.Image, opts)
			}
		} else {
			// Draw single instance
			opts.GeoM.Reset()
			opts.GeoM.Translate(finalX, finalY)
			screen.DrawImage(layer.Image, opts)
		}
	}
}

// GetLayersByEnvironment returns the appropriate layers for the given environment
func GetLayersByEnvironment(environment string) []BackgroundLayer {
	switch environment {
	case "desert":
		return DesertLayers()
	case "forest":
		return ForestLayers()
	case "mountains":
		return MountainsLayers()
	default:
		return DesertLayers() // Default to desert
	}
}
