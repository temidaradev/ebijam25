package src

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/temidaradev/ebijam25/assets"
)

type Player struct {
	X           float64
	Y           float64
	VelocityX   float64
	VelocityY   float64
	Speed       float64
	JumpPower   float64
	OnGround    bool
	FacingRight bool
	Scale       float64

	AnimationManager *assets.SimpleAnimationManager

	MaxSpeed       float64
	Deceleration   float64
	JumpBufferTime float64
	CoyoteTime     float64
	groundBuffer   float64
	jumpBuffer     float64
	coyoteBuffer   float64

	WorldWidth  float64
	WorldHeight float64
	GroundLevel float64

	Camera     *Camera
	Controller *ControllerInput
}

const (
	GRAVITY                = 1200.0
	SPRITE_WIDTH           = 50
	SPRITE_HEIGHT          = 37
	HITBOX_WIDTH           = 20
	HITBOX_HEIGHT          = 32
	HITBOX_OFFSET_X        = 14
	HITBOX_OFFSET_Y        = 5
	MOVE_THRESHOLD         = 5.0
	GROUND_TOLERANCE       = 2.0
	MIN_VELOCITY_THRESHOLD = 10.0
)

func NewPlayer(x, y, worldWidth, worldHeight, groundLevel float64) *Player {
	animManager := assets.InitCharacterAnimations()

	animManager.SetAnimationSpeed(1.0)

	player := &Player{
		X:                x,
		Y:                y,
		VelocityX:        0,
		VelocityY:        0,
		Speed:            200.0,
		MaxSpeed:         300.0,
		Deceleration:     1200.0,
		JumpPower:        -450.0,
		OnGround:         false,
		FacingRight:      true,
		Scale:            2,
		AnimationManager: animManager,
		JumpBufferTime:   0.1,
		CoyoteTime:       0.1,
		jumpBuffer:       0,
		coyoteBuffer:     0,
		groundBuffer:     0,
		WorldWidth:       worldWidth,
		WorldHeight:      worldHeight,
		GroundLevel:      groundLevel,
		Camera:           NewCamera(1280, 720, 0, worldHeight),
		Controller:       NewControllerInput(),
	}

	return player
}

func (p *Player) Update(deltaTime float64) {
	p.updateTimers(deltaTime)

	p.Controller.Update()

	p.handleInput(deltaTime)
	p.updatePhysics(deltaTime)
	p.updateAnimation()

	if p.AnimationManager != nil {
		p.AnimationManager.Update(deltaTime)
	}

	if p.Camera != nil {
		p.Camera.Follow(p.X+(float64(SPRITE_WIDTH)*p.Scale/2), p.Y+(float64(SPRITE_HEIGHT)*p.Scale/2), p.VelocityX, p.VelocityY)
		p.Camera.Update(deltaTime)
	}
}

func (p *Player) updateTimers(deltaTime float64) {
	if p.jumpBuffer > 0 {
		p.jumpBuffer -= deltaTime
	}

	if p.coyoteBuffer > 0 {
		p.coyoteBuffer -= deltaTime
	}

	if p.groundBuffer > 0 {
		p.groundBuffer -= deltaTime
	}
}

func (p *Player) handleInput(deltaTime float64) {
	leftPressed := ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft)
	rightPressed := ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight)
	jumpPressed := inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyW) ||
		inpututil.IsKeyJustPressed(ebiten.KeyArrowUp)

	controllerLeft := p.Controller.IsLeftPressed()
	controllerRight := p.Controller.IsRightPressed()
	controllerJump := p.Controller.IsJumpJustPressed()
	horizontalAxis := p.Controller.GetHorizontalAxis()

	const deadZone = 0.2

	if (leftPressed || controllerLeft) && !(rightPressed || controllerRight) {
		if controllerLeft && !leftPressed && absFloat64(horizontalAxis) > deadZone {
			intensity := absFloat64(horizontalAxis)
			if intensity > 1.0 {
				intensity = 1.0
			}
			p.VelocityX = -p.MaxSpeed * intensity
		} else {
			p.VelocityX = -p.MaxSpeed
		}
		p.FacingRight = false
	} else if (rightPressed || controllerRight) && !(leftPressed || controllerLeft) {
		if controllerRight && !rightPressed && absFloat64(horizontalAxis) > deadZone {
			intensity := absFloat64(horizontalAxis)
			if intensity > 1.0 {
				intensity = 1.0
			}
			p.VelocityX = p.MaxSpeed * intensity
		} else {
			p.VelocityX = p.MaxSpeed
		}
		p.FacingRight = true
	} else {
		var decelAmount float64
		if p.OnGround {
			decelAmount = p.Deceleration * 1.8 * deltaTime
		} else {
			decelAmount = p.Deceleration * 0.3 * deltaTime
		}

		if p.VelocityX > decelAmount {
			p.VelocityX -= decelAmount
		} else if p.VelocityX < -decelAmount {
			p.VelocityX += decelAmount
		} else {
			p.VelocityX = 0
		}

		if absFloat64(p.VelocityX) < MIN_VELOCITY_THRESHOLD {
			p.VelocityX = 0
		}
	}

	if jumpPressed || controllerJump {
		p.jumpBuffer = p.JumpBufferTime
	}

	if p.jumpBuffer > 0 && (p.OnGround || p.coyoteBuffer > 0) {
		p.VelocityY = p.JumpPower
		p.OnGround = false
		p.jumpBuffer = 0
		p.coyoteBuffer = 0
	}
}

func (p *Player) updatePhysics(deltaTime float64) {
	wasOnGround := p.OnGround

	if !p.OnGround {
		p.VelocityY += GRAVITY * deltaTime
	}

	p.X += p.VelocityX * deltaTime
	p.Y += p.VelocityY * deltaTime

	hitboxBottom := p.Y + float64(HITBOX_OFFSET_Y)*p.Scale + (float64(HITBOX_HEIGHT) * p.Scale)
	if hitboxBottom >= p.GroundLevel {
		p.Y = p.GroundLevel - float64(HITBOX_OFFSET_Y)*p.Scale - (float64(HITBOX_HEIGHT) * p.Scale)
		p.VelocityY = 0
		p.OnGround = true
		p.groundBuffer = 0.05
	} else {
		p.OnGround = false
	}

	if wasOnGround && !p.OnGround && p.VelocityY >= 0 {
		p.coyoteBuffer = p.CoyoteTime
	}

	if p.Y > p.WorldHeight {
		p.Y = p.GroundLevel - float64(HITBOX_OFFSET_Y)*p.Scale - (float64(HITBOX_HEIGHT) * p.Scale)
		p.VelocityY = 0
		p.OnGround = true
	}
}

func (p *Player) updateAnimation() {
	if p.AnimationManager != nil {
		if !p.OnGround {
			if p.VelocityY > 50 {
				p.AnimationManager.SetAnimation("fall")
			} else {
				p.AnimationManager.SetAnimation("jump")
			}
		} else {
			speed := absFloat64(p.VelocityX)
			if speed > p.MaxSpeed*0.7 {
				p.AnimationManager.SetAnimation("run")
			} else if speed > MIN_VELOCITY_THRESHOLD {
				p.AnimationManager.SetAnimation("walk")
			} else {
				p.AnimationManager.SetAnimation("idle")
			}
		}
	}
}

func (p *Player) Draw(screen *ebiten.Image) {
	if p.AnimationManager != nil {
		op := &ebiten.DrawImageOptions{}

		op.GeoM.Scale(p.Scale, p.Scale)

		if !p.FacingRight {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(float64(SPRITE_WIDTH)*p.Scale, 0)
		}

		op.GeoM.Translate(p.X, p.Y)

		p.AnimationManager.DrawWithOptions(screen, op)
	} else {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(p.Scale, p.Scale)

		if !p.FacingRight {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(float64(SPRITE_WIDTH)*p.Scale, 0)
		}

		op.GeoM.Translate(p.X, p.Y)

		firstFrame := assets.CharacterSpritesheet.SubImage(image.Rect(0, 0, SPRITE_WIDTH, SPRITE_HEIGHT)).(*ebiten.Image)
		screen.DrawImage(firstFrame, op)
	}
}

func (p *Player) GetBounds() (x, y, width, height float64) {
	hitboxWidth := float64(HITBOX_WIDTH) * p.Scale
	hitboxHeight := float64(HITBOX_HEIGHT) * p.Scale
	offsetX := float64(HITBOX_OFFSET_X) * p.Scale
	offsetY := float64(HITBOX_OFFSET_Y) * p.Scale
	return p.X + offsetX, p.Y + offsetY, hitboxWidth, hitboxHeight
}

func (p *Player) SetPosition(x, y float64) {
	p.X = x
	p.Y = y
}

func (p *Player) GetPosition() (x, y float64) {
	return p.X, p.Y
}

func (p *Player) SetWorldBounds(width, height, groundLevel float64) {
	p.WorldWidth = width
	p.WorldHeight = height
	p.GroundLevel = groundLevel
}

func (p *Player) IsOnGround() bool {
	return p.OnGround
}

func (p *Player) GetVelocity() (vx, vy float64) {
	return p.VelocityX, p.VelocityY
}

func (p *Player) SetVelocity(vx, vy float64) {
	p.VelocityX = vx
	p.VelocityY = vy
}

func (p *Player) GetCamera() *Camera {
	return p.Camera
}

func (p *Player) SetCameraSettings(followSpeed, lookAhead, deadZone, verticalOffset float64) {
	if p.Camera != nil {
		p.Camera.FollowSpeed = followSpeed
		p.Camera.LookAhead = lookAhead
		p.Camera.DeadZone = deadZone
		p.Camera.VerticalOffset = verticalOffset
	}
}

func absFloat64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
