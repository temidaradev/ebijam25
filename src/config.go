package src

const (
	DefaultScreenWidth    = 1280
	DefaultScreenHeight   = 720
	Gravity               = 1400.0
	MinVelocityThreshold  = 25.0
	SpriteWidth           = 50
	SpriteHeight          = 37
	HitboxWidth           = 18
	HitboxHeight          = 30
	HitboxOffsetX         = 16
	HitboxOffsetY         = 10
	DefaultPlayerSpeed    = 200.0
	DefaultMaxSpeed       = 250.0
	DefaultJumpPower      = 650.0
	DefaultDeceleration   = 3500.0
	DefaultJumpBufferTime = 0.1
	DefaultCoyoteTime     = 0.15
	RollDuration          = 0.4
	RollSpeed             = 400.0
)

type GameConfig struct {
	ShowCollisionBoxes bool
	ShowDebugInfo      bool
	VsyncEnabled       bool
	WindowDecorated    bool
	WindowResizable    bool
}
