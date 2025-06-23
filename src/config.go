package src

const (
	DefaultScreenWidth          = 1280
	DefaultScreenHeight         = 720
	TargetFPS                   = 60
	Gravity                     = 1400.0
	MinVelocityThreshold        = 25.0
	GroundTolerance             = 2.0
	MoveThreshold               = 5.0
	SpriteWidth                 = 50
	SpriteHeight                = 37
	HitboxWidth                 = 18
	HitboxHeight                = 30
	HitboxOffsetX               = 16
	HitboxOffsetY               = 10
	DefaultPlayerSpeed          = 200.0
	DefaultMaxSpeed             = 250.0
	DefaultJumpPower            = 650.0
	DefaultDeceleration         = 3500.0
	DefaultJumpBufferTime       = 0.1
	DefaultCoyoteTime           = 0.15
	RollDuration                = 0.4
	RollSpeed                   = 400.0
	WallJumpDuration            = 0.3
	DashSpeed                   = 600.0
	DashDuration                = 0.2
	DashCooldownTime            = 1.0
	WallGrabTime                = 2.0
	InvulnerabilityTime         = 1.0
	DefaultAttackDamage         = 25
	DefaultAttackRange          = 40.0
	DefaultAttackCooldown       = 0.5
	DefaultComboWindow          = 1.0
	DefaultMaxHealth            = 100
	DefaultCameraFollowSpeed    = 8.0
	DefaultCameraLookAhead      = 100.0
	DefaultCameraDeadZone       = 50.0
	DefaultCameraVerticalOffset = -20.0
	DefaultEnemyHealth          = 50
	DefaultEnemySpeed           = 100.0
	DefaultEnemyAttackDamage    = 20
	DefaultEnemyDetectionRange  = 200.0
	ShooterEnemyProjectileSpeed = 300.0
	ShooterEnemyFireRate        = 2.0
	JumperEnemyJumpForce        = 400.0
	JumperEnemyJumpCooldown     = 3.0
	SpikeEnemyDamage            = 40
	SpikeEnemyKnockback         = 200.0
)

var (
	DebugCollisionColor = [4]float32{1, 0, 0, 0.5}
	DebugPlayerColor    = [4]float32{0, 1, 0, 0.5}
	DebugEnemyColor     = [4]float32{1, 1, 0, 0.5}
)

type GameConfig struct {
	ShowCollisionBoxes bool
	ShowDebugInfo      bool
	VsyncEnabled       bool
	WindowDecorated    bool
	WindowResizable    bool
}

func DefaultGameConfig() GameConfig {
	return GameConfig{
		ShowCollisionBoxes: false,
		ShowDebugInfo:      false,
		VsyncEnabled:       true,
		WindowDecorated:    true,
		WindowResizable:    true,
	}
}
