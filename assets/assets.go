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
	BgZ1         = esset.GetAsset(assets, "images/start/bg_1 (z -1).png")
	BgZ2         = esset.GetAsset(assets, "images/start/bg_2 (z -2).png")
	BgZ3         = esset.GetAsset(assets, "images/start/bg_3 (z -3).png")
	MountainsZ4  = esset.GetAsset(assets, "images/start/mtn (z -4).png")
	GradientZ6   = esset.GetAsset(assets, "images/start/Gradient (z -6).png")
	MiddleGround = esset.GetAsset(assets, "images/start/middleground (z 0).png")
	Foreground   = esset.GetAsset(assets, "images/start/middleplus (z 1).png")

	// Load the actual adventurer sprite sheet
	CharacterSpritesheet = esset.GetAsset(assets, "images/sprites/adventurer-Sheet.png")
)

// Initialize simple character animations
func InitCharacterAnimations() *SimpleAnimationManager {
	// Create the simple animation manager
	animManager := NewSimpleAnimationManager(CharacterSpritesheet, 50, 37)

	// Add animations with faster, more responsive timings
	animManager.AddAnimation("idle", 0, 3, 0.12, true)     // idle: frames 0-3, faster idle
	animManager.AddAnimation("run", 8, 13, 0.06, true)     // run: frames 8-13, much faster run
	animManager.AddAnimation("jump", 14, 17, 0.08, false)  // jump: frames 14-17, snappy jump
	animManager.AddAnimation("fall", 22, 23, 0.1, true)    // fall: frames 22-23, faster fall
	animManager.AddAnimation("walk", 155, 160, 0.18, true) // walk: frames 155-160, slower walk for natural pace

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
