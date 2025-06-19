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
	verticalOffset := -80.0
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

	// Handle attacks
	if attackPressed && !p.IsAttacking && p.AttackCooldown <= 0 {
		p.performAttack()
	}

	// Handle rolling/sliding
	if !p.IsRolling && rollPressed && p.OnGround {
		p.IsRolling = true
		p.RollTimer = RollDuration
		if p.FacingRight {
			p.VelocityX = RollSpeed
		} else {
			p.VelocityX = -RollSpeed
		}
	}

	// Continue sliding while key is held or timer is active
	if p.IsRolling {
		// If key is held, keep sliding (reset timer to extend slide)
		if slideHeld && p.OnGround {
			p.RollTimer = RollDuration * 0.5 // Shorter timer for held slides
		}

		p.RollTimer -= deltaTime
		if p.RollTimer <= 0 || !p.OnGround {
			p.IsRolling = false
		}

		// Continue movement during slide but allow direction change
		if (leftPressed || controllerLeft) && !(rightPressed || controllerRight) {
			if p.VelocityX > -RollSpeed {
				p.VelocityX = -RollSpeed
			}
			p.FacingRight = false
		} else if (rightPressed || controllerRight) && !(leftPressed || controllerLeft) {
			if p.VelocityX < RollSpeed {
				p.VelocityX = RollSpeed
			}
			p.FacingRight = true
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
		// Improved deceleration for parkour
		var decelAmount float64
		if p.OnGround {
			// Much faster ground deceleration to prevent sliding
			// Use different friction based on current speed for better feel
			speedMultiplier := math.Max(1.0, math.Abs(p.VelocityX)/100.0) // More friction at higher speeds
			decelAmount = p.Deceleration * 3.5 * speedMultiplier * deltaTime
		} else if p.OnWallLeft || p.OnWallRight {
			// Moderate air friction on walls
			decelAmount = p.Deceleration * 0.8 * deltaTime
		} else {
			// Minimal air friction for better air control
			decelAmount = p.Deceleration * 0.2 * deltaTime
		}

		if p.VelocityX > decelAmount {
			p.VelocityX -= decelAmount
		} else if p.VelocityX < -decelAmount {
			p.VelocityX += decelAmount
		} else {
			p.VelocityX = 0
		}

		// Stop completely when velocity is very low to prevent micro-sliding
		if math.Abs(p.VelocityX) < MinVelocityThreshold {
			p.VelocityX = 0
		}
	}

	if jumpPressed || controllerJump {
		p.jumpBuffer = p.JumpBufferTime
	}

	if p.jumpBuffer > 0 {
		// Wall climbing - if holding against wall and can grab
		if (p.OnWallLeft || p.OnWallRight) && p.CanWallGrab && !p.OnGround {
			// Check if moving toward the wall (holding against it)
			if (p.OnWallLeft && p.IsMovingLeft) ||
				(p.OnWallRight && p.IsMovingRight) {
				// Start wall climbing
				p.IsWallClimbing = true
				p.WallGrabTimer = WALL_GRAB_STAMINA
				p.VelocityY = -WALL_CLIMB_SPEED // Climb up
				p.jumpBuffer = 0
				p.DoubleJumpUsed = false // Reset double jump on wall grab
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
			// Ground jump or coyote time jump
			p.VelocityY = p.JumpPower
			p.OnGround = false
			p.jumpBuffer = 0
			p.coyoteBuffer = 0
			p.DoubleJumpUsed = false // Reset double jump on ground jump
		} else if p.HasDoubleJump && !p.DoubleJumpUsed && !p.OnGround {
			// Double jump in air
			p.VelocityY = p.JumpPower * 0.85 // Slightly weaker double jump
			p.DoubleJumpUsed = true
			p.jumpBuffer = 0
		}
	}
}

func (p *Player) updatePhysics(deltaTime float64) {
	wasOnGround := p.OnGround

	// Variable jump height - reduce upward velocity if jump key is released
	jumpHeld := ebiten.IsKeyPressed(ebiten.KeySpace) ||
		ebiten.IsKeyPressed(ebiten.KeyW) ||
		ebiten.IsKeyPressed(ebiten.KeyArrowUp) ||
		p.Controller.IsJumpPressed()

	if p.VelocityY < -100 && !jumpHeld {
		// Cut jump short if key is released
		p.VelocityY *= 0.5
	}

	// Apply gravity with wall sliding and wall climbing
	if !p.OnGround {
		if p.IsWallClimbing {
			// No gravity while wall climbing, but check if still holding wall
			if !((p.OnWallLeft && p.IsMovingLeft) ||
				(p.OnWallRight && p.IsMovingRight)) {
				// Player let go of wall
				p.IsWallClimbing = false
				p.WallGrabTimer = 0
			}
		} else if (p.OnWallLeft || p.OnWallRight) && p.VelocityY > 0 {
			// Wall sliding - reduce fall speed
			p.VelocityY += Gravity * deltaTime * 0.3
			if p.VelocityY > WALL_SLIDE_SPEED {
				p.VelocityY = WALL_SLIDE_SPEED
			}
		} else {
			// Normal gravity
			p.VelocityY += Gravity * deltaTime
		}
	}

	// Calculate movement deltas
	deltaX := p.VelocityX * deltaTime
	deltaY := p.VelocityY * deltaTime

	if p.CollisionSystem != nil {
		// Get current position
		currentBox := p.GetCollisionBox()

		// Calculate target position
		targetX := currentBox.X + deltaX
		targetY := currentBox.Y + deltaY

		// Check horizontal movement first
		horizontalBox := CollisionBox{
			X:      targetX,
			Y:      currentBox.Y,
			Width:  currentBox.Width,
			Height: currentBox.Height,
		}

		// Check vertical movement
		verticalBox := CollisionBox{
			X:      currentBox.X,
			Y:      targetY,
			Width:  currentBox.Width,
			Height: currentBox.Height,
		}

		// Final target position
		finalX := targetX
		finalY := targetY

		// Check horizontal collision
		if p.CollisionSystem.CheckCollisionAtPoint(horizontalBox) {
			finalX = currentBox.X // Block horizontal movement
			p.VelocityX = 0
		}

		// Check vertical collision
		if p.CollisionSystem.CheckCollisionAtPoint(verticalBox) {
			finalY = currentBox.Y // Block vertical movement
			if p.VelocityY > 0 {  // Landing on ground
				p.VelocityY = 0
				if !p.OnGround {
					// Just landed - apply strong friction based on landing velocity
					landingSpeed := math.Abs(p.VelocityY)
					frictionMultiplier := 0.3 // Base friction (keep 30% of horizontal velocity)

					// Apply extra friction for hard landings
					if landingSpeed > 300 {
						frictionMultiplier = 0.1 // Keep only 10% for hard landings
					} else if landingSpeed > 150 {
						frictionMultiplier = 0.2 // Keep 20% for medium landings
					}

					p.VelocityX *= frictionMultiplier
					p.DoubleJumpUsed = false // Reset double jump on landing
					p.CanWallGrab = true     // Reset wall grab ability on landing
					p.IsWallClimbing = false // Stop wall climbing
					p.WallGrabTimer = 0
				}
				p.OnGround = true
			} else if p.VelocityY < 0 { // Hit ceiling
				p.VelocityY = 0
			}
		} else {
			p.OnGround = false // In air
		}

		// Update position
		p.SetPosition(finalX, finalY)

		// Check if player is on ground (check slightly below current position)
		groundCheckBox := CollisionBox{
			X:      finalX,
			Y:      finalY + 5,
			Width:  currentBox.Width,
			Height: currentBox.Height,
		}

		p.OnGround = p.CollisionSystem.CheckCollisionAtPoint(groundCheckBox)

		// Removed boundary checks - player can move outside world bounds
	} else {
		// Fallback to basic movement if no collision system
		p.X += deltaX
		p.Y += deltaY

		// Removed boundary checks - player can move outside world bounds

		// No artificial ground - player will fall until hitting tiles
		p.OnGround = false
	}

	// Handle coyote time and ground state
	if wasOnGround && !p.OnGround && p.VelocityY >= 0 {
		p.coyoteBuffer = p.CoyoteTime
	}

	// Reset OnGround if we're moving up (jumping)
	if p.VelocityY < -10 {
		p.OnGround = false
	}

	// Remove artificial ground fallback - only use tile collision
}

func (p *Player) updateAnimation() {
	if p.AnimationManager != nil {
		// Priority: Attack animations take precedence
		if p.IsAttacking {
			// Choose attack animation based on combo and ground state
			if !p.OnGround {
				// Air attacks
				if p.ComboCount <= 1 {
					p.AnimationManager.SetAnimation("air-attack1")
				} else {
					p.AnimationManager.SetAnimation("air-attack2")
				}
			} else {
				// Ground attacks
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

		// Check if player is hurt (invulnerable)
		if p.InvulnTimer > 0 {
			p.AnimationManager.SetAnimation("hurt")
			return
		}

		// Rolling animation
		if p.IsRolling {
			p.AnimationManager.SetAnimation("roll")
			return
		}

		// Wall climbing animation
		if p.IsWallClimbing {
			p.AnimationManager.SetAnimation("jump") // Use jump animation for wall climbing
			return
		}

		// Movement animations
		if !p.OnGround {
			if (p.OnWallLeft || p.OnWallRight) && p.VelocityY > 0 {
				// Wall sliding
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

// CheckCollisionAtPosition checks if the player would collide at a specific position
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

// IsOnGroundCheck performs a more precise ground check
func (p *Player) IsOnGroundCheck() bool {
	if p.TileMap == nil {
		return p.Y >= p.GroundLevel-float64(SpriteHeight)*p.Scale
	}

	hitboxWidth := float64(HitboxWidth) * p.Scale
	hitboxHeight := float64(HitboxHeight) * p.Scale
	hitboxOffsetX := float64(HitboxOffsetX) * p.Scale
	hitboxOffsetY := float64(HitboxOffsetY) * p.Scale

	// Check a few pixels below the player
	return p.TileMap.CheckCollision(p.X+hitboxOffsetX, p.Y+hitboxOffsetY+3, hitboxWidth, hitboxHeight)
}

// GetHitboxBounds returns the actual hitbox bounds (used for collision)
func (p *Player) GetHitboxBounds() (x, y, width, height float64) {
	hitboxWidth := float64(HitboxWidth) * p.Scale
	hitboxHeight := float64(HitboxHeight) * p.Scale
	offsetX := float64(HitboxOffsetX) * p.Scale
	offsetY := float64(HitboxOffsetY) * p.Scale

	return p.X + offsetX, p.Y + offsetY, hitboxWidth, hitboxHeight
}

// CanMoveHorizontal checks if horizontal movement is possible
func (p *Player) CanMoveHorizontal(deltaX float64) bool {
	newX := p.X + deltaX
	return !p.CheckCollisionAtPosition(newX, p.Y)
}

// CanMoveVertical checks if vertical movement is possible
func (p *Player) CanMoveVertical(deltaY float64) bool {
	newY := p.Y + deltaY
	return !p.CheckCollisionAtPosition(p.X, newY)
}

// ResetToSafePosition moves the player to a safe position if they're stuck
func (p *Player) ResetToSafePosition() {
	if p.CollisionSystem == nil {
		return
	}

	currentBox := p.GetCollisionBox()
	if p.CollisionSystem.CheckCollisionAtPoint(currentBox) {
		// Try to find a safe position above the current position
		safeBox := CollisionBox{
			X:      currentBox.X,
			Y:      currentBox.Y - 100, // Move up
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
			// Last resort: move to spawn position
			p.X = 100.0
			p.Y = p.GroundLevel - float64(SpriteHeight)*p.Scale
			p.VelocityX = 0
			p.VelocityY = 0
			p.OnGround = true
		}
	}
}

// IsStuck checks if the player is currently stuck in collision
func (p *Player) IsStuck() bool {
	if p.CollisionSystem == nil {
		return false
	}
	return p.CollisionSystem.CheckCollisionAtPoint(p.GetCollisionBox())
}

// checkWallCollision checks if player is touching walls for wall jumping
func (p *Player) checkWallCollision() {
	if p.CollisionSystem == nil {
		p.OnWallLeft = false
		p.OnWallRight = false
		return
	}

	currentBox := p.GetCollisionBox()

	// Check left wall
	leftBox := CollisionBox{
		X:      currentBox.X - 5,
		Y:      currentBox.Y,
		Width:  currentBox.Width,
		Height: currentBox.Height,
	}
	p.OnWallLeft = p.CollisionSystem.CheckCollisionAtPoint(leftBox) && !p.OnGround

	// Check right wall
	rightBox := CollisionBox{
		X:      currentBox.X + 5,
		Y:      currentBox.Y,
		Width:  currentBox.Width,
		Height: currentBox.Height,
	}
	p.OnWallRight = p.CollisionSystem.CheckCollisionAtPoint(rightBox) && !p.OnGround
}

// TakeDamage handles player taking damage
func (p *Player) TakeDamage(damage int) {
	if p.InvulnTimer > 0 || p.IsDead {
		return // Player is invulnerable or already dead
	}

	p.Health -= damage
	if p.Health <= 0 {
		p.Health = 0
		p.IsDead = true
		// Stop player movement on death
		p.VelocityX = 0
		p.VelocityY = 0
		// Don't automatically reset position - let game handle respawn menu
	} else {
		p.InvulnTimer = INVULNERABILITY_TIME
	}
}

// IsInvulnerable returns true if player is currently invulnerable
func (p *Player) IsInvulnerable() bool {
	return p.InvulnTimer > 0
}

// GetHealthPercentage returns health as a percentage
func (p *Player) GetHealthPercentage() float64 {
	return float64(p.Health) / float64(p.MaxHealth)
}

// CheckProjectileCollision checks if player collides with a projectile
func (p *Player) CheckProjectileCollision(projectile *Projectile) bool {
	if p.IsInvulnerable() || !projectile.IsActive {
		return false
	}

	px, py, pw, ph := p.GetBounds()
	return projectile.CheckCollision(px, py, pw, ph)
}

// performAttack initiates a player attack
func (p *Player) performAttack() {
	p.IsAttacking = true
	p.AttackTimer = ATTACK_DURATION
	p.AttackCooldown = ATTACK_COOLDOWN_TIME

	// Handle combo system
	if p.CanCombo && p.ComboTimer > 0 {
		p.ComboCount++
		if p.ComboCount > MAX_COMBO_COUNT {
			p.ComboCount = MAX_COMBO_COUNT
		}
	} else {
		p.ComboCount = 1
	}

	p.ComboTimer = COMBO_WINDOW
	p.CanCombo = true

	// Add slight knockback/movement on attack for better feel
	if p.OnGround {
		// Slight forward movement on ground attacks
		if p.FacingRight {
			p.VelocityX += 50.0
		} else {
			p.VelocityX -= 50.0
		}
	} else {
		// Air attacks have different movement
		if p.ComboCount >= 3 {
			// Third air attack creates downward momentum
			p.VelocityY += 200.0
		}
	}
}

// GetAttackBox returns the attack hitbox based on player position and facing direction
func (p *Player) GetAttackBox() (float64, float64, float64, float64) {
	if !p.IsAttacking {
		return 0, 0, 0, 0
	}

	attackWidth := p.AttackRange
	attackHeight := float64(HitboxHeight) * p.Scale

	var attackX, attackY float64

	if p.FacingRight {
		attackX = p.X + float64(SpriteWidth)*p.Scale
	} else {
		attackX = p.X - attackWidth
	}

	attackY = p.Y + float64(HitboxOffsetY)*p.Scale

	return attackX, attackY, attackWidth, attackHeight
}

// CheckAttackHit checks if the player's attack hits an enemy
func (p *Player) CheckAttackHit(enemyX, enemyY, enemyWidth, enemyHeight float64) bool {
	if !p.IsAttacking {
		return false
	}

	attackX, attackY, attackWidth, attackHeight := p.GetAttackBox()

	// Simple AABB collision detection
	return attackX < enemyX+enemyWidth &&
		attackX+attackWidth > enemyX &&
		attackY < enemyY+enemyHeight &&
		attackY+attackHeight > enemyY
}

// GetAttackDamage returns the damage based on combo count
func (p *Player) GetAttackDamage() int {
	baseDamage := p.AttackDamage

	// Increase damage based on combo
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

// IsPerformingAttack returns true if player is currently attacking
func (p *Player) IsPerformingAttack() bool {
	return p.IsAttacking
}

// GetComboCount returns current combo count
func (p *Player) GetComboCount() int {
	return p.ComboCount
}

func absFloat64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// UpdateCollisionSystem updates the player's collision system with a new tilemap
func (p *Player) UpdateCollisionSystem(tileMap *assets.TileMap) {
	p.CollisionSystem = NewCollisionSystem(tileMap)
}

// IsDead returns true if player is dead
func (p *Player) IsPlayerDead() bool {
	return p.IsDead
}

// Revive restores the player to life with full health
func (p *Player) Revive() {
	p.IsDead = false
	p.Health = p.MaxHealth
	p.InvulnTimer = INVULNERABILITY_TIME
}
