package assets

import (
	"embed"
	"image"
	_ "image/jpeg"
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
	CharacterSpritesheet = esset.GetAsset(assets, "images/sprites/adventurer-Sheet.png")

	DesertBackground1 = esset.GetAsset(assets, "images/backgrounds/desert/background1.png")
	DesertBackground2 = esset.GetAsset(assets, "images/backgrounds/desert/background2.png")
	DesertBackground3 = esset.GetAsset(assets, "images/backgrounds/desert/background3.png")

	ForestSky      = esset.GetAsset(assets, "images/backgrounds/forest/sky.png")
	ForestSkyCloud = esset.GetAsset(assets, "images/backgrounds/forest/sky_cloud.png")
	ForestMountain = esset.GetAsset(assets, "images/backgrounds/forest/mountain2.png")
	ForestCloud    = esset.GetAsset(assets, "images/backgrounds/forest/cloud.png")
	ForestPine1    = esset.GetAsset(assets, "images/backgrounds/forest/pine1.png")
	ForestPine2    = esset.GetAsset(assets, "images/backgrounds/forest/pine2.png")

	MountainsSky         = esset.GetAsset(assets, "images/backgrounds/mountains/sky.png")
	MountainsCloudsBg    = esset.GetAsset(assets, "images/backgrounds/mountains/clouds_bg.png")
	MountainsGlacial     = esset.GetAsset(assets, "images/backgrounds/mountains/glacial_mountains.png")
	MountainsCloudsMg3   = esset.GetAsset(assets, "images/backgrounds/mountains/clouds_mg_3.png")
	MountainsCloudsMg2   = esset.GetAsset(assets, "images/backgrounds/mountains/clouds_mg_2.png")
	MountainsCloudsMg1   = esset.GetAsset(assets, "images/backgrounds/mountains/clouds_mg_1.png")
	MountainsCloudLonely = esset.GetAsset(assets, "images/backgrounds/mountains/cloud_lonely.png")

	Cave1 = esset.GetAsset(assets, "images/backgrounds/cave/1.png")
	Cave2 = esset.GetAsset(assets, "images/backgrounds/cave/2.png")
	Cave3 = esset.GetAsset(assets, "images/backgrounds/cave/3.png")
	Cave4 = esset.GetAsset(assets, "images/backgrounds/cave/4.png")
	Cave5 = esset.GetAsset(assets, "images/backgrounds/cave/5.png")
	Cave6 = esset.GetAsset(assets, "images/backgrounds/cave/6.png")
	Cave7 = esset.GetAsset(assets, "images/backgrounds/cave/7.png")
)

func InitCharacterAnimations() *SimpleAnimationManager {
	animManager := NewSimpleAnimationManager(CharacterSpritesheet, 50, 37)
	animManager.AddAnimation("idle", 0, 3, 0.12, true)
	animManager.AddAnimation("run", 8, 13, 0.06, true)
	animManager.AddAnimation("jump", 14, 17, 0.08, false)
	animManager.AddAnimation("fall", 22, 23, 0.1, true)
	animManager.AddAnimation("walk", 155, 160, 0.18, true)
	animManager.AddAnimation("roll", 24, 27, 0.06, false)
	animManager.SetAnimation("idle")
	return animManager
}

type SimpleAnimationManager struct {
	spritesheet    *ebiten.Image
	animations     map[string]*SimpleAnimation
	currentAnim    string
	previousAnim   string
	currentFrame   int
	frameTimer     float64
	frameWidth     int
	frameHeight    int
	animationSpeed float64
}

type SimpleAnimation struct {
	name      string
	frames    []image.Rectangle
	durations []float64
	loop      bool
}

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
		animationSpeed: 1.0,
	}
}

func (sam *SimpleAnimationManager) AddAnimation(name string, startFrame, endFrame int, frameDuration float64, loop bool) {
	frames := make([]image.Rectangle, 0, endFrame-startFrame+1)
	durations := make([]float64, 0, endFrame-startFrame+1)

	for i := startFrame; i <= endFrame; i++ {
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

func (sam *SimpleAnimationManager) SetAnimation(name string) {
	if name != sam.currentAnim {
		sam.previousAnim = sam.currentAnim
		sam.currentAnim = name
		sam.currentFrame = 0
		sam.frameTimer = 0
	}
}

func (sam *SimpleAnimationManager) SetAnimationSpeed(speed float64) {
	sam.animationSpeed = speed
}

func (sam *SimpleAnimationManager) Update(dt float64) {
	if sam.currentAnim == "" {
		return
	}
	anim, exists := sam.animations[sam.currentAnim]
	if !exists || len(anim.frames) == 0 {
		return
	}
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

type AnimationManager struct {
	animations    map[AnimationState]*SpriteAnimation
	currentState  AnimationState
	previousState AnimationState
}

func NewAnimationManager() *AnimationManager {
	return &AnimationManager{
		animations:    make(map[AnimationState]*SpriteAnimation),
		currentState:  AnimationIdle,
		previousState: AnimationIdle,
	}
}

func (am *AnimationManager) AddAnimation(state AnimationState, animation *SpriteAnimation) {
	am.animations[state] = animation
}

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

func (am *AnimationManager) Update(dt float64) {
	if anim, exists := am.animations[am.currentState]; exists {
		anim.Update(dt)
	}
}

func (am *AnimationManager) DrawWithOptions(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	if anim, exists := am.animations[am.currentState]; exists {
		anim.DrawWithOptions(screen, op)
	}
}

func (a *SpriteAnimation) Play() {
	a.Playing = true
}

type SpriteSheet struct {
	Image        *ebiten.Image
	SpriteWidth  int
	SpriteHeight int
	Columns      int
	Rows         int
}

func (sam *SimpleAnimationManager) GetCurrentAnimation() string {
	return sam.currentAnim
}

func (sam *SimpleAnimationManager) GetCurrentFrame() int {
	return sam.currentFrame
}

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
		{MountainsSky, "mountains_sky", 0.05, 0.02, -6, 0, 0, true, false, 2.5, 2.5},
		{MountainsCloudsBg, "mountains_clouds_bg", 0.1, 0.05, -5, 0, 10, true, false, 2.5, 2.5},
		{MountainsGlacial, "mountains_glacial", 0.2, 0.08, -4, 0, 80, true, false, 2.5, 2.5},
		{MountainsCloudsMg3, "mountains_clouds_mg3", 0.25, 0.1, -3, 50, 120, true, false, 2.5, 2.5},
		{MountainsCloudsMg2, "mountains_clouds_mg2", 0.4, 0.15, -2, 208, 140, true, false, 2.5, 2.5},
		{MountainsCloudsMg1, "mountains_clouds_mg1", 0.6, 0.25, -1, 128, 145, true, false, 2.5, 2.5},
		{MountainsCloudLonely, "mountains_cloud_lonely", 0.8, 0.35, 0, 320, 112, false, false, 2.5, 2.5},
	}
}

func CaveLayers() []BackgroundLayer {
	return []BackgroundLayer{
		{Cave7, "cave_7", 0.1, 0.05, -6, 0, -100, true, false, 2, 2},
		{Cave6, "cave_6", 0.25, 0.1, -5, 0, -80, true, false, 2, 2},
		{Cave5, "cave_5", 0.4, 0.15, -4, 0, -60, true, false, 2, 2},
		{Cave4, "cave_4", 0.55, 0.2, -3, 0, -40, true, false, 2, 2},
		{Cave3, "cave_3", 0.7, 0.25, -2, 0, -20, true, false, 2, 2},
		{Cave2, "cave_2", 0.85, 0.3, -1, 0, -10, true, false, 2, 2},
		{Cave1, "cave_1", 1.0, 0.35, 0, 0, 0, true, false, 2, 2},
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
	case "cave":
		return CaveLayers()
	default:
		return DesertLayers()
	}
}

func DrawBackgroundLayers(screen *ebiten.Image, layers []BackgroundLayer, cameraX, cameraY float64, screenWidth, screenHeight int) {
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

			effectiveScreenWidth := float64(screenWidth)

			for startX > 0 {
				startX -= imgWidth
			}

			for x := startX; x < effectiveScreenWidth; x += imgWidth {
				opts.GeoM.Reset()
				opts.GeoM.Translate(x, finalY)
				opts.GeoM.Scale(layer.ScaleX, layer.ScaleY)
				screen.DrawImage(layer.Image, opts)
			}
		} else {
			opts.GeoM.Reset()
			opts.GeoM.Translate(finalX, finalY)
			opts.GeoM.Scale(layer.ScaleX, layer.ScaleY)
			screen.DrawImage(layer.Image, opts)
		}
	}
}
