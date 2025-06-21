package src

import (
	"image"
	"math"

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

	SpeedMultiplier  float64
	SlowdownTimer    float64
	SlowdownDuration float64

	WorldWidth  float64
	WorldHeight float64
	GroundLevel float64

	Camera          *Camera
	Controller      *ControllerInput
	TileMap         *assets.TileMap
	CollisionSystem *CollisionSystem

	IsRolling bool
	RollTimer float64

	// Parkour mechanics
	CanWallJump   bool
	WallJumpTimer float64
	OnWallLeft    bool
	OnWallRight   bool

	// Dashing
	CanDash      bool
	IsDashing    bool
	DashTimer    float64
	DashCooldown float64
	DashSpeed    float64
	DashDuration float64

	// Double jump
	HasDoubleJump  bool
	DoubleJumpUsed bool

	// Health system
	Health      int
	MaxHealth   int
	InvulnTimer float64
	IsDead      bool

	// Combat system
	IsAttacking    bool
	AttackTimer    float64
	AttackDamage   int
	AttackRange    float64
	AttackCooldown float64
	ComboCount     int
	ComboTimer     float64
	CanCombo       bool

	// Wall climbing system
	IsWallClimbing bool
	WallGrabTimer  float64
	CanWallGrab    bool

	// Input state (for physics calculations)
	IsMovingLeft  bool
	IsMovingRight bool

	// Slipping state
	IsSlipping bool
	SlipTimer  float64
}

const (
	WALL_JUMP_POWER      = -550.0
	WALL_JUMP_HORIZONTAL = 320.0
	WALL_SLIDE_SPEED     = 120.0
	WALL_JUMP_TIME       = 0.15

	DASH_SPEED    = 450.0
	DASH_DURATION = 0.2
	DASH_COOLDOWN = 0.8

	INVULNERABILITY_TIME = 1.0

	ATTACK_DURATION      = 0.3
	ATTACK_COOLDOWN_TIME = 0.4
	ATTACK_RANGE         = 60.0
	ATTACK_DAMAGE        = 1
	COMBO_WINDOW         = 1.2
	MAX_COMBO_COUNT      = 3

	WALL_CLIMB_SPEED  = 200.0
	WALL_GRAB_STAMINA = 3.0
)

func NewPlayer(x, y, worldWidth, worldHeight, groundLevel float64, tileMap *assets.TileMap) *Player {
	animManager := assets.InitCharacterAnimations()

	animManager.SetAnimationSpeed(1.0)

	var cameraWorldW, cameraWorldH float64
	if tileMap != nil {
		cameraWorldW = float64(tileMap.PixelWidth)
		cameraWorldH = float64(tileMap.PixelHeight)
	} else {
		cameraWorldW = worldWidth
		cameraWorldH = worldHeight
	}
	// Set vertical offset to move camera view a bit higher
	verticalOffset := -90.0
	player := &Player{
		X:                x,
		Y:                y,
		VelocityX:        0,
		VelocityY:        0,
		Speed:            DefaultPlayerSpeed,
		MaxSpeed:         DefaultMaxSpeed,
		Deceleration:     DefaultDeceleration,
		JumpPower:        -DefaultJumpPower,
		OnGround:         false,
		FacingRight:      true,
		Scale:            1.8,
		AnimationManager: animManager,
		JumpBufferTime:   DefaultJumpBufferTime,
		CoyoteTime:       DefaultCoyoteTime,
		jumpBuffer:       0,
		coyoteBuffer:     0,
		groundBuffer:     0,
		WorldWidth:       worldWidth,
		WorldHeight:      worldHeight,
		GroundLevel:      groundLevel,
		Camera:           NewCamera(DefaultScreenWidth, DefaultScreenHeight, cameraWorldW, cameraWorldH),
		Controller:       NewControllerInput(),
		TileMap:          tileMap,
		CollisionSystem:  NewCollisionSystem(tileMap),

		// Slowdown system
		SpeedMultiplier:  1.0,
		SlowdownTimer:    0,
		SlowdownDuration: 0,

		// Parkour mechanics
		CanWallJump:   true,
		WallJumpTimer: 0,
		OnWallLeft:    false,
		OnWallRight:   false,

		// Dashing
		CanDash:      true,
		IsDashing:    false,
		DashTimer:    0,
		DashCooldown: 0,
		DashSpeed:    DASH_SPEED,
		DashDuration: DASH_DURATION,

		// Double jump
		HasDoubleJump:  true,
		DoubleJumpUsed: false,

		// Health
		Health:      5, // More health for parkour challenges
		MaxHealth:   5,
		InvulnTimer: 0,
		IsDead:      false,

		// Combat
		IsAttacking:    false,
		AttackTimer:    0,
		AttackDamage:   ATTACK_DAMAGE,
		AttackRange:    ATTACK_RANGE,
		AttackCooldown: 0,
		ComboCount:     0,
		ComboTimer:     0,
		CanCombo:       false,

		// Wall climbing
		IsWallClimbing: false,
		WallGrabTimer:  0,
		CanWallGrab:    true,

		// Input state
		IsMovingLeft:  false,
		IsMovingRight: false,
	}

	player.Camera.VerticalOffset = verticalOffset

	return player
}

func (p *Player) Update(deltaTime float64) {
	p.updateTimers(deltaTime)
	p.updateSlowdown(deltaTime)

	p.Controller.Update()

	p.handleInput(deltaTime)
	p.updatePhysics(deltaTime)
	p.updateAnimation()

	if p.AnimationManager != nil {
		p.AnimationManager.Update(deltaTime)
	}

	if p.Camera != nil {
		p.Camera.Follow(p.X+(float64(SpriteWidth)*p.Scale/2), p.Y+(float64(SpriteHeight)*p.Scale/2), p.VelocityX, p.VelocityY)
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

	if p.RollTimer > 0 {
		p.RollTimer -= deltaTime
		if p.RollTimer <= 0 {
			p.IsRolling = false
		}
	}

	if p.InvulnTimer > 0 {
		p.InvulnTimer -= deltaTime
	}

	// Combat timers
	if p.AttackTimer > 0 {
		p.AttackTimer -= deltaTime
		if p.AttackTimer <= 0 {
			p.IsAttacking = false
		}
	}

	if p.AttackCooldown > 0 {
		p.AttackCooldown -= deltaTime
	}

	if p.ComboTimer > 0 {
		p.ComboTimer -= deltaTime
		if p.ComboTimer <= 0 {
			p.ComboCount = 0
			p.CanCombo = false
		}
	}

	// Wall climbing timers
	if p.WallGrabTimer > 0 {
		p.WallGrabTimer -= deltaTime
		if p.WallGrabTimer <= 0 {
			p.IsWallClimbing = false
			p.CanWallGrab = false // Player gets tired and falls
		}
	}

	// Slip timer
	if p.SlipTimer > 0 {
		p.SlipTimer -= deltaTime
		if p.SlipTimer <= 0 {
			p.IsSlipping = false
		}
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

	rollPressed := inpututil.IsKeyJustPressed(ebiten.KeyShift) || inpututil.IsKeyJustPressed(ebiten.KeyZ) || p.Controller.IsRollJustPressed()
	slideHeld := ebiten.IsKeyPressed(ebiten.KeyShift) || ebiten.IsKeyPressed(ebiten.KeyZ)
	attackPressed := inpututil.IsKeyJustPressed(ebiten.KeyJ) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || p.Controller.IsAttackJustPressed()

	const deadZone = 0.2

	// Store input state for physics calculations
	p.IsMovingLeft = (leftPressed || controllerLeft)
	p.IsMovingRight = (rightPressed || controllerRight)

	// Handle attacks (annoying restriction: can't attack while sliding AND brief delay after landing)
	landingDelay := p.OnGround && p.groundBuffer > 0 // Just landed
	if attackPressed && !p.IsAttacking && p.AttackCooldown <= 0 && !p.IsRolling && !landingDelay {
		p.performAttack()
	}

	// Handle rolling/sliding
	if !p.IsRolling && rollPressed && p.OnGround {
		p.IsRolling = true
		p.RollTimer = RollDuration
		// Inherit current velocity for momentum-based sliding
		slideSpeed := RollSpeed
		if math.Abs(p.VelocityX) > slideSpeed {
			// Maintain momentum if moving faster than roll speed
			slideSpeed = math.Abs(p.VelocityX)
		}

		if p.FacingRight {
			p.VelocityX = slideSpeed
		} else {
			p.VelocityX = -slideSpeed
		}
	}

	// Continue sliding while key is held or timer is active
	if p.IsRolling {
		// Enhanced slide control for parkour
		if slideHeld && p.OnGround {
			p.RollTimer = RollDuration * 0.6 // Extended slide duration
		}

		// Apply slide friction gradually
		slideFriction := 0.95 // Less friction for smoother slides
		if p.OnGround {
			p.VelocityX *= slideFriction
		}

		// Allow directional control during slide (but maintain momentum)
		if (leftPressed || controllerLeft) && !(rightPressed || controllerRight) {
			if p.VelocityX > 0 {
				// Turning around during slide - reduce speed more
				p.VelocityX *= 0.8
			}
			if p.VelocityX > -RollSpeed*0.5 {
				p.VelocityX = math.Max(p.VelocityX-RollSpeed*0.3, -RollSpeed)
			}
			p.FacingRight = false
		} else if (rightPressed || controllerRight) && !(leftPressed || controllerLeft) {
			if p.VelocityX < 0 {
				// Turning around during slide - reduce speed more
				p.VelocityX *= 0.8
			}
			if p.VelocityX < RollSpeed*0.5 {
				p.VelocityX = math.Min(p.VelocityX+RollSpeed*0.3, RollSpeed)
			}
			p.FacingRight = true
		}

		// End slide conditions
		p.RollTimer -= deltaTime
		if p.RollTimer <= 0 || !p.OnGround || math.Abs(p.VelocityX) < 50 {
			p.IsRolling = false
		}

		return
	}

	// Update wall detection for parkour
	p.checkWallCollision()

	if (leftPressed || controllerLeft) && !(rightPressed || controllerRight) {
		if controllerLeft && !leftPressed && absFloat64(horizontalAxis) > deadZone {
			intensity := absFloat64(horizontalAxis)
			if intensity > 1.0 {
				intensity = 1.0
			}
			p.VelocityX = -p.MaxSpeed * intensity * p.SpeedMultiplier
		} else {
			p.VelocityX = -p.MaxSpeed * p.SpeedMultiplier
		}
		p.FacingRight = false
	} else if (rightPressed || controllerRight) && !(leftPressed || controllerLeft) {
		if controllerRight && !rightPressed && absFloat64(horizontalAxis) > deadZone {
			intensity := absFloat64(horizontalAxis)
			if intensity > 1.0 {
				intensity = 1.0
			}
			p.VelocityX = p.MaxSpeed * intensity * p.SpeedMultiplier
		} else {
			p.VelocityX = p.MaxSpeed * p.SpeedMultiplier
		}
		p.FacingRight = true
	} else {
		var decelAmount float64
		if p.OnGround {
			baseDecel := p.Deceleration * 2.8
			speedFactor := math.Min(2.0, math.Abs(p.VelocityX)/150.0)
			decelAmount = baseDecel * speedFactor * deltaTime

			if math.Abs(p.VelocityX) > p.MaxSpeed*1.2 {
				if math.Mod(p.X+p.Y, 100) < 10 {
					decelAmount *= 0.1
					if !p.IsSlipping {
						p.IsSlipping = true
						p.SlipTimer = 0.4
					}
				}
			}

			if math.Abs(p.VelocityX) > p.MaxSpeed*1.5 {
				decelAmount *= 0.6
			}
		} else if p.OnWallLeft || p.OnWallRight {
			decelAmount = p.Deceleration * 0.7 * deltaTime
		} else {
			decelAmount = p.Deceleration * 0.15 * deltaTime
		}

		if p.VelocityX > decelAmount {
			p.VelocityX -= decelAmount
		} else if p.VelocityX < -decelAmount {
			p.VelocityX += decelAmount
		} else {
			p.VelocityX = 0
		}

		velocityThreshold := 15.0
		if math.Abs(p.VelocityX) < velocityThreshold {
			p.VelocityX = 0
		}
	}

	if jumpPressed || controllerJump {
		p.jumpBuffer = p.JumpBufferTime
	}

	if p.jumpBuffer > 0 {
		if (p.OnWallLeft || p.OnWallRight) && p.CanWallGrab && !p.OnGround {
			if (p.OnWallLeft && p.IsMovingLeft) ||
				(p.OnWallRight && p.IsMovingRight) {
				p.IsWallClimbing = true
				p.WallGrabTimer = WALL_GRAB_STAMINA
				p.VelocityY = -WALL_CLIMB_SPEED
				p.jumpBuffer = 0
				p.DoubleJumpUsed = false
			} else {
				// Wall jump (jumping away from wall)
				if p.OnWallLeft {
					p.VelocityX = WALL_JUMP_HORIZONTAL
					p.FacingRight = true
				} else if p.OnWallRight {
					p.VelocityX = -WALL_JUMP_HORIZONTAL
					p.FacingRight = false
				}
				p.VelocityY = WALL_JUMP_POWER
				p.WallJumpTimer = WALL_JUMP_TIME
				p.jumpBuffer = 0
				p.DoubleJumpUsed = false
			}
		} else if p.OnGround || p.coyoteBuffer > 0 {
			p.VelocityY = p.JumpPower
			p.OnGround = false
			p.jumpBuffer = 0
			p.coyoteBuffer = 0
			p.DoubleJumpUsed = false

			if math.Mod(p.X*p.Y, 100) < 10 {
				p.VelocityY *= 0.7
			}
		} else if p.HasDoubleJump && !p.DoubleJumpUsed && !p.OnGround {
			p.VelocityY = p.JumpPower * 0.85
			p.DoubleJumpUsed = true
			p.jumpBuffer = 0
		}
	}
}

func (p *Player) updatePhysics(deltaTime float64) {
	wasOnGround := p.OnGround

	jumpHeld := ebiten.IsKeyPressed(ebiten.KeySpace) ||
		ebiten.IsKeyPressed(ebiten.KeyW) ||
		ebiten.IsKeyPressed(ebiten.KeyArrowUp) ||
		p.Controller.IsJumpPressed()

	if p.VelocityY < -100 && !jumpHeld {
		p.VelocityY *= 0.5
	}

	if !p.OnGround {
		if p.IsWallClimbing {
			if !((p.OnWallLeft && p.IsMovingLeft) ||
				(p.OnWallRight && p.IsMovingRight)) {
				// Player let go of wall
				p.IsWallClimbing = false
				p.WallGrabTimer = 0
			}
		} else if (p.OnWallLeft || p.OnWallRight) && p.VelocityY > 0 {
			p.VelocityY += Gravity * deltaTime * 0.3
			if p.VelocityY > WALL_SLIDE_SPEED {
				p.VelocityY = WALL_SLIDE_SPEED
			}
		} else {
			p.VelocityY += Gravity * deltaTime
		}
	}

	deltaX := p.VelocityX * deltaTime
	deltaY := p.VelocityY * deltaTime

	if p.CollisionSystem != nil {
		currentBox := p.GetCollisionBox()

		targetX := currentBox.X + deltaX
		targetY := currentBox.Y + deltaY

		horizontalBox := CollisionBox{
			X:      targetX,
			Y:      currentBox.Y,
			Width:  currentBox.Width,
			Height: currentBox.Height,
		}

		verticalBox := CollisionBox{
			X:      currentBox.X,
			Y:      targetY,
			Width:  currentBox.Width,
			Height: currentBox.Height,
		}

		finalX := targetX
		finalY := targetY

		if p.CollisionSystem.CheckCollisionAtPoint(horizontalBox) {
			finalX = currentBox.X
			p.VelocityX = 0
		}

		if p.CollisionSystem.CheckCollisionAtPoint(verticalBox) {
			finalY = currentBox.Y
			if p.VelocityY > 0 {
				p.VelocityY = 0
				if !p.OnGround {
					landingSpeed := math.Abs(p.VelocityY)

					var frictionMultiplier float64
					if landingSpeed > 400 {
						frictionMultiplier = 0.6
					} else if landingSpeed > 250 {
						frictionMultiplier = 0.7
					} else if landingSpeed > 150 {
						frictionMultiplier = 0.8
					} else {
						frictionMultiplier = 0.9
					}

					horizontalSpeed := math.Abs(p.VelocityX)
					if horizontalSpeed > p.MaxSpeed*1.2 {
						frictionMultiplier = math.Max(frictionMultiplier, 0.8)
					}

					p.VelocityX *= frictionMultiplier
					p.DoubleJumpUsed = false
					p.CanWallGrab = true
					p.IsWallClimbing = false
					p.WallGrabTimer = 0

					p.groundBuffer = 0.15
				}
				p.OnGround = true
			} else if p.VelocityY < 0 {
				p.VelocityY = 0
				p.VelocityX *= 0.8
			}
		} else {
			p.OnGround = false
		}

		p.SetPosition(finalX, finalY)

		groundCheckBox := CollisionBox{
			X:      finalX,
			Y:      finalY + 5,
			Width:  currentBox.Width,
			Height: currentBox.Height,
		}

		p.OnGround = p.CollisionSystem.CheckCollisionAtPoint(groundCheckBox)
	} else {
		p.X += deltaX
		p.Y += deltaY
		p.OnGround = false
	}

	if wasOnGround && !p.OnGround && p.VelocityY >= 0 {
		p.coyoteBuffer = p.CoyoteTime
	}

	if p.VelocityY < -10 {
		p.OnGround = false
	}
}

func (p *Player) updateAnimation() {
	if p.AnimationManager != nil {
		if p.IsAttacking {
			if !p.OnGround {
				if p.ComboCount <= 1 {
					p.AnimationManager.SetAnimation("air-attack1")
				} else {
					p.AnimationManager.SetAnimation("air-attack2")
				}
			} else {
				switch p.ComboCount {
				case 1:
					p.AnimationManager.SetAnimation("attack1")
				case 2:
					p.AnimationManager.SetAnimation("attack2")
				case 3:
					p.AnimationManager.SetAnimation("attack3")
				default:
					p.AnimationManager.SetAnimation("attack1")
				}
			}
			return
		}

		if p.InvulnTimer > 0 {
			p.AnimationManager.SetAnimation("hurt")
			return
		}

		if p.IsRolling {
			p.AnimationManager.SetAnimation("roll")
			return
		}

		if p.IsSlipping && p.OnGround {
			p.AnimationManager.SetAnimation("slip")
			return
		}

		if p.IsWallClimbing {
			p.AnimationManager.SetAnimation("jump")
			return
		}

		if !p.OnGround {
			if (p.OnWallLeft || p.OnWallRight) && p.VelocityY > 0 {
				p.AnimationManager.SetAnimation("fall")
			} else if p.VelocityY > 50 {
				p.AnimationManager.SetAnimation("fall")
			} else {
				p.AnimationManager.SetAnimation("jump")
			}
		} else {
			speed := math.Abs(p.VelocityX)
			if speed > p.MaxSpeed*0.7 {
				p.AnimationManager.SetAnimation("run")
			} else if speed > MinVelocityThreshold {
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
			op.GeoM.Translate(float64(SpriteWidth)*p.Scale, 0)
		}

		op.GeoM.Translate(p.X, p.Y)

		p.AnimationManager.DrawWithOptions(screen, op)
	} else {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(p.Scale, p.Scale)

		if !p.FacingRight {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(float64(SpriteWidth)*p.Scale, 0)
		}

		op.GeoM.Translate(p.X, p.Y)

		firstFrame := assets.CharacterSpritesheet.SubImage(image.Rect(0, 0, SpriteWidth, SpriteHeight)).(*ebiten.Image)
		screen.DrawImage(firstFrame, op)
	}
}

func (p *Player) GetBounds() (x, y, width, height float64) {
	hitboxWidth := float64(HitboxWidth) * p.Scale
	hitboxHeight := float64(HitboxHeight) * p.Scale
	offsetX := float64(HitboxOffsetX) * p.Scale
	offsetY := float64(HitboxOffsetY) * p.Scale
	return p.X + offsetX, p.Y + offsetY, hitboxWidth, hitboxHeight
}

// Implement GameObject interface
func (p *Player) GetCollisionBox() CollisionBox {
	hitboxWidth := float64(HitboxWidth) * p.Scale
	hitboxHeight := float64(HitboxHeight) * p.Scale
	offsetX := float64(HitboxOffsetX) * p.Scale
	offsetY := float64(HitboxOffsetY) * p.Scale

	return CollisionBox{
		X:      p.X + offsetX,
		Y:      p.Y + offsetY,
		Width:  hitboxWidth,
		Height: hitboxHeight,
	}
}

func (p *Player) SetPosition(x, y float64) {
	offsetX := float64(HitboxOffsetX) * p.Scale
	offsetY := float64(HitboxOffsetY) * p.Scale

	p.X = x - offsetX
	p.Y = y - offsetY
}

func (p *Player) OnCollision(info CollisionInfo) {
	if info.HitWall {
		p.VelocityX = 0
	}
	if info.HitGround {
		p.OnGround = true
		p.VelocityY = 0
		p.groundBuffer = 0.05
	}
	if info.HitCeiling {
		p.VelocityY = 0
	}

	// Additional safety check: if we have any collision and are falling, check if we should be on ground
	if info.HasCollision && p.VelocityY > 0 {
		currentBox := p.GetCollisionBox()
		if p.CollisionSystem.CheckCollisionAtPoint(currentBox) {
			p.OnGround = true
			p.VelocityY = 0
		}
	}
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

func (p *Player) CheckCollisionAtPosition(x, y float64) bool {
	if p.TileMap == nil {
		return false
	}

	hitboxWidth := float64(HitboxWidth) * p.Scale
	hitboxHeight := float64(HitboxHeight) * p.Scale
	hitboxOffsetX := float64(HitboxOffsetX) * p.Scale
	hitboxOffsetY := float64(HitboxOffsetY) * p.Scale

	return p.TileMap.CheckCollision(x+hitboxOffsetX, y+hitboxOffsetY, hitboxWidth, hitboxHeight)
}

func (p *Player) IsOnGroundCheck() bool {
	if p.TileMap == nil {
		return p.Y >= p.GroundLevel-float64(SpriteHeight)*p.Scale
	}

	hitboxWidth := float64(HitboxWidth) * p.Scale
	hitboxHeight := float64(HitboxHeight) * p.Scale
	hitboxOffsetX := float64(HitboxOffsetX) * p.Scale
	hitboxOffsetY := float64(HitboxOffsetY) * p.Scale

	return p.TileMap.CheckCollision(p.X+hitboxOffsetX, p.Y+hitboxOffsetY+3, hitboxWidth, hitboxHeight)
}

func (p *Player) GetHitboxBounds() (x, y, width, height float64) {
	hitboxWidth := float64(HitboxWidth) * p.Scale
	hitboxHeight := float64(HitboxHeight) * p.Scale
	offsetX := float64(HitboxOffsetX) * p.Scale
	offsetY := float64(HitboxOffsetY) * p.Scale

	return p.X + offsetX, p.Y + offsetY, hitboxWidth, hitboxHeight
}

func (p *Player) CanMoveHorizontal(deltaX float64) bool {
	newX := p.X + deltaX
	return !p.CheckCollisionAtPosition(newX, p.Y)
}

func (p *Player) CanMoveVertical(deltaY float64) bool {
	newY := p.Y + deltaY
	return !p.CheckCollisionAtPosition(p.X, newY)
}

func (p *Player) ResetToSafePosition() {
	if p.CollisionSystem == nil {
		return
	}

	currentBox := p.GetCollisionBox()
	if p.CollisionSystem.CheckCollisionAtPoint(currentBox) {
		safeBox := CollisionBox{
			X:      currentBox.X,
			Y:      currentBox.Y - 100,
			Width:  currentBox.Width,
			Height: currentBox.Height,
		}

		safeX, safeY, found := p.CollisionSystem.GetSafePosition(safeBox)
		if found {
			p.SetPosition(safeX, safeY)
			p.VelocityX = 0
			p.VelocityY = 0
			p.OnGround = false
		} else {
			p.X = 100.0
			p.Y = p.GroundLevel - float64(SpriteHeight)*p.Scale
			p.VelocityX = 0
			p.VelocityY = 0
			p.OnGround = true
		}
	}
}

func (p *Player) IsStuck() bool {
	if p.CollisionSystem == nil {
		return false
	}
	return p.CollisionSystem.CheckCollisionAtPoint(p.GetCollisionBox())
}

func (p *Player) checkWallCollision() {
	if p.CollisionSystem == nil {
		p.OnWallLeft = false
		p.OnWallRight = false
		return
	}

	currentBox := p.GetCollisionBox()

	leftBox := CollisionBox{
		X:      currentBox.X - 5,
		Y:      currentBox.Y,
		Width:  currentBox.Width,
		Height: currentBox.Height,
	}
	p.OnWallLeft = p.CollisionSystem.CheckCollisionAtPoint(leftBox) && !p.OnGround

	rightBox := CollisionBox{
		X:      currentBox.X + 5,
		Y:      currentBox.Y,
		Width:  currentBox.Width,
		Height: currentBox.Height,
	}
	p.OnWallRight = p.CollisionSystem.CheckCollisionAtPoint(rightBox) && !p.OnGround
}

func (p *Player) TakeDamage(damage int) {
	if p.InvulnTimer > 0 || p.IsDead {
		return
	}

	p.Health -= damage
	if p.Health <= 0 {
		p.Health = 0
		p.IsDead = true
		p.VelocityX = 0
		p.VelocityY = 0
	} else {
		p.InvulnTimer = INVULNERABILITY_TIME
	}
}

func (p *Player) IsInvulnerable() bool {
	return p.InvulnTimer > 0
}

func (p *Player) GetHealthPercentage() float64 {
	return float64(p.Health) / float64(p.MaxHealth)
}

func (p *Player) CheckProjectileCollision(projectile *Projectile) bool {
	if p.IsInvulnerable() || !projectile.IsActive {
		return false
	}

	px, py, pw, ph := p.GetBounds()
	return projectile.CheckCollision(px, py, pw, ph)
}

func (p *Player) performAttack() {
	p.IsAttacking = true
	p.AttackTimer = ATTACK_DURATION
	p.AttackCooldown = ATTACK_COOLDOWN_TIME

	if p.ComboTimer > 0 && p.CanCombo {
		p.ComboCount++
		if p.ComboCount > MAX_COMBO_COUNT {
			p.ComboCount = MAX_COMBO_COUNT
		}
	} else {
		p.ComboCount = 1
	}

	p.ComboTimer = COMBO_WINDOW
	p.CanCombo = true

	if p.OnGround {
		if p.FacingRight {
			p.VelocityX += 100.0
		} else {
			p.VelocityX -= 100.0
		}
	} else {
		if p.ComboCount >= 3 {
			p.VelocityY += 200.0
		}
	}
}

func (p *Player) GetAttackBox() (float64, float64, float64, float64) {
	if !p.IsAttacking {
		return 0, 0, 0, 0
	}

	attackWidth := p.AttackRange
	attackHeight := float64(HitboxHeight) * p.Scale * 0.8

	var attackX, attackY float64

	playerHitboxX, playerHitboxY, _, _ := p.GetBounds()

	if p.FacingRight {
		attackX = playerHitboxX + float64(HitboxWidth)*p.Scale
	} else {
		attackX = playerHitboxX - attackWidth
	}

	attackY = playerHitboxY + (float64(HitboxHeight)*p.Scale-attackHeight)/2

	return attackX, attackY, attackWidth, attackHeight
}

func (p *Player) CheckAttackHit(enemyX, enemyY, enemyWidth, enemyHeight float64) bool {
	if !p.IsAttacking {
		return false
	}

	attackX, attackY, attackWidth, attackHeight := p.GetAttackBox()

	return attackX < enemyX+enemyWidth &&
		attackX+attackWidth > enemyX &&
		attackY < enemyY+enemyHeight &&
		attackY+attackHeight > enemyY
}

func (p *Player) GetAttackDamage() int {
	baseDamage := p.AttackDamage

	switch p.ComboCount {
	case 1:
		return baseDamage
	case 2:
		return baseDamage + 1
	case 3:
		return baseDamage + 2
	default:
		return baseDamage
	}
}

func (p *Player) IsPerformingAttack() bool {
	return p.IsAttacking
}

func (p *Player) GetComboCount() int {
	return p.ComboCount
}

func (p *Player) GetAttackDebugInfo() (bool, bool, float64, float64, float64, float64) {
	attackX, attackY, attackWidth, attackHeight := p.GetAttackBox()
	return p.IsAttacking, p.AttackCooldown <= 0, attackX, attackY, attackWidth, attackHeight
}

func absFloat64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func (p *Player) UpdateCollisionSystem(tileMap *assets.TileMap) {
	p.CollisionSystem = NewCollisionSystem(tileMap)
}

func (p *Player) IsPlayerDead() bool {
	return p.IsDead
}

func (p *Player) Revive() {
	p.IsDead = false
	p.Health = p.MaxHealth
	p.InvulnTimer = INVULNERABILITY_TIME
}

func (p *Player) ApplySlowdown(multiplier, duration float64) {
	p.SpeedMultiplier = multiplier
	p.SlowdownTimer = duration
}

func (p *Player) updateSlowdown(deltaTime float64) {
	if p.SlowdownTimer > 0 {
		p.SlowdownTimer -= deltaTime
		if p.SlowdownTimer <= 0 {
			p.SpeedMultiplier = 1.0 // Reset to normal speed
		}
	}
}
