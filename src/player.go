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

	Camera          *Camera
	Controller      *ControllerInput
	TileMap         *assets.TileMap
	CollisionSystem *CollisionSystem

	IsRolling bool
	RollTimer float64
}

const (
	GRAVITY                = 1200.0
	SPRITE_WIDTH           = 50
	SPRITE_HEIGHT          = 37
	HITBOX_WIDTH           = 18
	HITBOX_HEIGHT          = 30
	HITBOX_OFFSET_X        = 16
	HITBOX_OFFSET_Y        = 10
	MOVE_THRESHOLD         = 5.0
	GROUND_TOLERANCE       = 2.0
	MIN_VELOCITY_THRESHOLD = 10.0
	ROLL_DURATION          = 0.7 // Increased from 0.4 for longer roll
	ROLL_SPEED             = 500.0
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
		Speed:            200.0,
		MaxSpeed:         300.0,
		Deceleration:     1200.0,
		JumpPower:        -450.0,
		OnGround:         false,
		FacingRight:      true,
		Scale:            1.8,
		AnimationManager: animManager,
		JumpBufferTime:   0.1,
		CoyoteTime:       0.1,
		jumpBuffer:       0,
		coyoteBuffer:     0,
		groundBuffer:     0,
		WorldWidth:       worldWidth,
		WorldHeight:      worldHeight,
		GroundLevel:      groundLevel,
		Camera:           NewCamera(1280, 720, cameraWorldW, cameraWorldH),
		Controller:       NewControllerInput(),
		TileMap:          tileMap,
		CollisionSystem:  NewCollisionSystem(tileMap),
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

	if p.RollTimer > 0 {
		p.RollTimer -= deltaTime
		if p.RollTimer <= 0 {
			p.IsRolling = false
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

	const deadZone = 0.2

	if !p.IsRolling && rollPressed && p.OnGround {
		p.IsRolling = true
		p.RollTimer = ROLL_DURATION
		if p.FacingRight {
			p.VelocityX = ROLL_SPEED
		} else {
			p.VelocityX = -ROLL_SPEED
		}
	}

	if p.IsRolling {
		p.RollTimer -= deltaTime
		if p.RollTimer <= 0 {
			p.IsRolling = false
		}
		return
	}

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

	// Apply gravity
	if !p.OnGround {
		p.VelocityY += GRAVITY * deltaTime
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
			if p.VelocityY > 0 {  // Only stop downward velocity when hitting ground
				p.VelocityY = 0
				p.OnGround = true
			} else if p.VelocityY < 0 { // Hit ceiling
				p.VelocityY = 0
			}
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
		if p.IsRolling {
			p.AnimationManager.SetAnimation("roll")
			return
		}
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

// Implement GameObject interface
func (p *Player) GetCollisionBox() CollisionBox {
	hitboxWidth := float64(HITBOX_WIDTH) * p.Scale
	hitboxHeight := float64(HITBOX_HEIGHT) * p.Scale
	offsetX := float64(HITBOX_OFFSET_X) * p.Scale
	offsetY := float64(HITBOX_OFFSET_Y) * p.Scale

	return CollisionBox{
		X:      p.X + offsetX,
		Y:      p.Y + offsetY,
		Width:  hitboxWidth,
		Height: hitboxHeight,
	}
}

func (p *Player) SetPosition(x, y float64) {
	offsetX := float64(HITBOX_OFFSET_X) * p.Scale
	offsetY := float64(HITBOX_OFFSET_Y) * p.Scale

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

	hitboxWidth := float64(HITBOX_WIDTH) * p.Scale
	hitboxHeight := float64(HITBOX_HEIGHT) * p.Scale
	hitboxOffsetX := float64(HITBOX_OFFSET_X) * p.Scale
	hitboxOffsetY := float64(HITBOX_OFFSET_Y) * p.Scale

	return p.TileMap.CheckCollision(x+hitboxOffsetX, y+hitboxOffsetY, hitboxWidth, hitboxHeight)
}

// IsOnGroundCheck performs a more precise ground check
func (p *Player) IsOnGroundCheck() bool {
	if p.TileMap == nil {
		return p.Y >= p.GroundLevel-float64(SPRITE_HEIGHT)*p.Scale
	}

	hitboxWidth := float64(HITBOX_WIDTH) * p.Scale
	hitboxHeight := float64(HITBOX_HEIGHT) * p.Scale
	hitboxOffsetX := float64(HITBOX_OFFSET_X) * p.Scale
	hitboxOffsetY := float64(HITBOX_OFFSET_Y) * p.Scale

	// Check a few pixels below the player
	return p.TileMap.CheckCollision(p.X+hitboxOffsetX, p.Y+hitboxOffsetY+3, hitboxWidth, hitboxHeight)
}

// GetHitboxBounds returns the actual hitbox bounds (used for collision)
func (p *Player) GetHitboxBounds() (x, y, width, height float64) {
	hitboxWidth := float64(HITBOX_WIDTH) * p.Scale
	hitboxHeight := float64(HITBOX_HEIGHT) * p.Scale
	offsetX := float64(HITBOX_OFFSET_X) * p.Scale
	offsetY := float64(HITBOX_OFFSET_Y) * p.Scale
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
			p.Y = p.GroundLevel - float64(SPRITE_HEIGHT)*p.Scale
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
