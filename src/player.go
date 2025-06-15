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

	// Simple animation system
	AnimationManager *assets.SimpleAnimationManager

	// Better physics properties
	MaxSpeed       float64
	Deceleration   float64
	JumpBufferTime float64
	CoyoteTime     float64
	groundBuffer   float64
	jumpBuffer     float64
	coyoteBuffer   float64

	// World boundaries
	WorldWidth  float64
	WorldHeight float64
	GroundLevel float64
}

const (
	GRAVITY                = 1200.0
	SPRITE_WIDTH           = 50   // Adventurer sprite width
	SPRITE_HEIGHT          = 37   // Adventurer sprite height
	HITBOX_WIDTH           = 30   // Smaller hitbox width for collision
	HITBOX_HEIGHT          = 32   // Smaller hitbox height for collision
	HITBOX_OFFSET_X        = 10   // Offset to center hitbox horizontally
	HITBOX_OFFSET_Y        = 5    // Offset to center hitbox vertically
	MOVE_THRESHOLD         = 5.0  // Threshold for movement animation
	GROUND_TOLERANCE       = 2.0  // Tolerance for ground detection
	MIN_VELOCITY_THRESHOLD = 10.0 // Minimum velocity before stopping completely
)

func NewPlayer(x, y, worldWidth, worldHeight, groundLevel float64) *Player {
	// Initialize simple animation system
	animManager := assets.InitCharacterAnimations()

	// Set animation speed for more natural feel
	animManager.SetAnimationSpeed(1.0) // Normal animation speed

	player := &Player{
		X:            x,
		Y:            y,
		VelocityX:    0,
		VelocityY:    0,
		Speed:        200.0,
		MaxSpeed:     300.0,
		Deceleration: 1200.0, // Increased for less sliding
		JumpPower:    -450.0,
		OnGround:     false, // Start in air and let physics handle ground detection
		FacingRight:  true,
		Scale:        2, // Much smaller scale to make character more proportional

		// Simple animation system
		AnimationManager: animManager,

		// Jump mechanics
		JumpBufferTime: 0.1,
		CoyoteTime:     0.1,
		jumpBuffer:     0,
		coyoteBuffer:   0,
		groundBuffer:   0,

		// World boundaries
		WorldWidth:  worldWidth,
		WorldHeight: worldHeight,
		GroundLevel: groundLevel,
	}

	return player
}

func (p *Player) Update(deltaTime float64) {
	p.updateTimers(deltaTime)
	p.handleInput(deltaTime)
	p.updatePhysics(deltaTime)
	p.updateAnimation()

	// Update the animation system
	if p.AnimationManager != nil {
		p.AnimationManager.Update(deltaTime)
	}
}

func (p *Player) updateTimers(deltaTime float64) {
	// Update jump buffer
	if p.jumpBuffer > 0 {
		p.jumpBuffer -= deltaTime
	}

	// Update coyote time
	if p.coyoteBuffer > 0 {
		p.coyoteBuffer -= deltaTime
	}

	// Update ground buffer
	if p.groundBuffer > 0 {
		p.groundBuffer -= deltaTime
	}
}

func (p *Player) handleInput(deltaTime float64) {
	// Handle horizontal movement with immediate velocity (no acceleration)
	leftPressed := ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft)
	rightPressed := ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight)

	if leftPressed && !rightPressed {
		// Immediate left movement
		p.VelocityX = -p.MaxSpeed
		p.FacingRight = false
	} else if rightPressed && !leftPressed {
		// Immediate right movement
		p.VelocityX = p.MaxSpeed
		p.FacingRight = true
	} else {
		// More aggressive deceleration when no input to reduce sliding
		var decelAmount float64
		if p.OnGround {
			// Ground friction - much stronger deceleration when on ground
			decelAmount = p.Deceleration * 1.8 * deltaTime // 80% stronger on ground
		} else {
			// Air resistance - normal deceleration in air
			decelAmount = p.Deceleration * 0.3 * deltaTime // Reduced in air for better air control
		}

		if p.VelocityX > decelAmount {
			p.VelocityX -= decelAmount
		} else if p.VelocityX < -decelAmount {
			p.VelocityX += decelAmount
		} else {
			p.VelocityX = 0 // Stop completely when velocity is very small
		}

		// Additional check: stop completely if velocity is below minimum threshold
		if absFloat64(p.VelocityX) < MIN_VELOCITY_THRESHOLD {
			p.VelocityX = 0
		}
	}

	// Handle jumping with buffer and coyote time
	jumpPressed := inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyW) ||
		inpututil.IsKeyJustPressed(ebiten.KeyArrowUp)

	if jumpPressed {
		p.jumpBuffer = p.JumpBufferTime
	}

	// Perform jump if conditions are met
	if p.jumpBuffer > 0 && (p.OnGround || p.coyoteBuffer > 0) {
		p.VelocityY = p.JumpPower
		p.OnGround = false
		p.jumpBuffer = 0
		p.coyoteBuffer = 0
	}
}

func (p *Player) updatePhysics(deltaTime float64) {
	// Store previous ground state for coyote time
	wasOnGround := p.OnGround

	// Apply gravity when not on ground
	if !p.OnGround {
		p.VelocityY += GRAVITY * deltaTime
	}

	// Update position
	p.X += p.VelocityX * deltaTime
	p.Y += p.VelocityY * deltaTime

	// Ground collision detection - check if player's hitbox bottom edge hits ground
	hitboxBottom := p.Y + float64(HITBOX_OFFSET_Y)*p.Scale + (float64(HITBOX_HEIGHT) * p.Scale)
	if hitboxBottom >= p.GroundLevel {
		p.Y = p.GroundLevel - float64(HITBOX_OFFSET_Y)*p.Scale - (float64(HITBOX_HEIGHT) * p.Scale) // Position player on top of ground
		p.VelocityY = 0
		p.OnGround = true
		p.groundBuffer = 0.05 // Small buffer for ground detection
	} else {
		p.OnGround = false
	}

	// Start coyote time when leaving ground
	if wasOnGround && !p.OnGround && p.VelocityY >= 0 {
		p.coyoteBuffer = p.CoyoteTime
	}

	// Horizontal bounds checking using hitbox
	hitboxWidth := float64(HITBOX_WIDTH) * p.Scale
	hitboxOffsetX := float64(HITBOX_OFFSET_X) * p.Scale
	if p.X+hitboxOffsetX < 0 {
		p.X = -hitboxOffsetX
		p.VelocityX = 0
	} else if p.X+hitboxOffsetX+hitboxWidth > p.WorldWidth {
		p.X = p.WorldWidth - hitboxWidth - hitboxOffsetX
		p.VelocityX = 0
	}

	// Vertical bounds checking (prevent falling through world)
	if p.Y > p.WorldHeight {
		p.Y = p.GroundLevel - float64(HITBOX_OFFSET_Y)*p.Scale - (float64(HITBOX_HEIGHT) * p.Scale)
		p.VelocityY = 0
		p.OnGround = true
	}
}

func (p *Player) updateAnimation() {
	if p.AnimationManager != nil {
		// More precise animation state detection
		if !p.OnGround {
			// Airborne animations - more responsive to velocity changes
			if p.VelocityY > 50 { // Only fall animation when falling fast enough
				p.AnimationManager.SetAnimation("fall")
			} else {
				p.AnimationManager.SetAnimation("jump")
			}
		} else {
			// Ground animations - distinguish between different movement speeds
			speed := absFloat64(p.VelocityX)
			if speed > p.MaxSpeed*0.7 { // High speed = run
				p.AnimationManager.SetAnimation("run")
			} else if speed > MIN_VELOCITY_THRESHOLD { // Medium speed = walk (using min threshold)
				p.AnimationManager.SetAnimation("walk")
			} else { // Low/no speed = idle (more immediate idle detection)
				p.AnimationManager.SetAnimation("idle")
			}
		}
	}
}

func (p *Player) Draw(screen *ebiten.Image) {
	if p.AnimationManager != nil {
		// Use simple animation system
		op := &ebiten.DrawImageOptions{}

		// Apply scale
		op.GeoM.Scale(p.Scale, p.Scale)

		// Flip horizontally if facing left
		if !p.FacingRight {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(float64(SPRITE_WIDTH)*p.Scale, 0)
		}

		// Apply position
		op.GeoM.Translate(p.X, p.Y)

		p.AnimationManager.DrawWithOptions(screen, op)
	} else {
		// Fallback: draw the first frame of the sprite sheet
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(p.Scale, p.Scale)

		// Flip horizontally if facing left
		if !p.FacingRight {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(float64(SPRITE_WIDTH)*p.Scale, 0)
		}

		op.GeoM.Translate(p.X, p.Y)

		// Draw just the first frame (0, 0, 50, 37) from the sprite sheet
		firstFrame := assets.CharacterSpritesheet.SubImage(image.Rect(0, 0, SPRITE_WIDTH, SPRITE_HEIGHT)).(*ebiten.Image)
		screen.DrawImage(firstFrame, op)
	}
}

// GetBounds returns the player's bounding rectangle for collision detection
func (p *Player) GetBounds() (x, y, width, height float64) {
	hitboxWidth := float64(HITBOX_WIDTH) * p.Scale
	hitboxHeight := float64(HITBOX_HEIGHT) * p.Scale
	offsetX := float64(HITBOX_OFFSET_X) * p.Scale
	offsetY := float64(HITBOX_OFFSET_Y) * p.Scale
	return p.X + offsetX, p.Y + offsetY, hitboxWidth, hitboxHeight
}

// SetPosition sets the player's position
func (p *Player) SetPosition(x, y float64) {
	p.X = x
	p.Y = y
}

// GetPosition returns the player's current position
func (p *Player) GetPosition() (x, y float64) {
	return p.X, p.Y
}

// SetWorldBounds updates the world boundaries
func (p *Player) SetWorldBounds(width, height, groundLevel float64) {
	p.WorldWidth = width
	p.WorldHeight = height
	p.GroundLevel = groundLevel
}

// IsOnGround returns whether the player is currently on the ground
func (p *Player) IsOnGround() bool {
	return p.OnGround
}

// GetVelocity returns the player's current velocity
func (p *Player) GetVelocity() (vx, vy float64) {
	return p.VelocityX, p.VelocityY
}

// SetVelocity sets the player's velocity (useful for external forces)
func (p *Player) SetVelocity(vx, vy float64) {
	p.VelocityX = vx
	p.VelocityY = vy
}

// Helper function for absolute value
func absFloat64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
