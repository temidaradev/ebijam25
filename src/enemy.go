package src

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type EnemyType int

const (
	EnemyTypeShooter EnemyType = iota
	EnemyTypeJumper
	EnemyTypeSpike
	EnemyTypeGlitched
)

type Enemy struct {
	X, Y           float64
	VelocityX      float64
	VelocityY      float64
	Width, Height  float64
	Health         int
	IsActive       bool
	EnemyType      EnemyType
	PatrolStartX   float64
	PatrolEndX     float64
	PatrolSpeed    float64
	MovingRight    bool
	ShootCooldown  float64
	ShootTimer     float64
	DetectionRange float64
	JumpTimer      float64
	JumpCooldown   float64
	JumpPower      float64
	OnGround       bool
	Projectiles    []*Projectile
	Color          color.RGBA
}

type Projectile struct {
	X, Y          float64
	VelocityX     float64
	VelocityY     float64
	Width, Height float64
	IsActive      bool
	Lifetime      float64
	Color         color.RGBA
}

func NewShooterEnemy(x, y float64) *Enemy {
	return &Enemy{
		X:              x,
		Y:              y,
		Width:          24,
		Height:         24,
		Health:         1,
		IsActive:       true,
		EnemyType:      EnemyTypeShooter,
		ShootCooldown:  3.0,
		ShootTimer:     0,
		DetectionRange: 250,
		PatrolStartX:   x - 80,
		PatrolEndX:     x + 80,
		PatrolSpeed:    30,
		MovingRight:    true,
		Projectiles:    make([]*Projectile, 0),
		Color:          color.RGBA{200, 80, 80, 255},
	}
}

func NewJumperEnemy(x, y float64) *Enemy {
	return &Enemy{
		X:            x,
		Y:            y,
		Width:        28,
		Height:       28,
		Health:       1,
		IsActive:     true,
		EnemyType:    EnemyTypeJumper,
		JumpCooldown: 2.5,
		JumpTimer:    0,
		JumpPower:    -300,
		OnGround:     true,
		PatrolStartX: x - 120,
		PatrolEndX:   x + 120,
		PatrolSpeed:  60,
		MovingRight:  true,
		Color:        color.RGBA{80, 200, 80, 255},
	}
}

func NewSpikeEnemy(x, y float64) *Enemy {
	return &Enemy{
		X:         x,
		Y:         y,
		Width:     20,
		Height:    20,
		Health:    1,
		IsActive:  true,
		EnemyType: EnemyTypeSpike,
		Color:     color.RGBA{150, 150, 150, 255},
	}
}

func NewGlitchedEnemy(x, y float64) *Enemy {
	return &Enemy{
		X:              x,
		Y:              y,
		Width:          20 + rand.Float64()*20,
		Height:         20 + rand.Float64()*20,
		Health:         1,
		IsActive:       true,
		EnemyType:      EnemyTypeGlitched,
		PatrolStartX:   x - 100,
		PatrolEndX:     x + 100,
		PatrolSpeed:    50 + rand.Float64()*100,
		MovingRight:    rand.Float64() > 0.5,
		DetectionRange: 50 + rand.Float64()*150,
		Color: color.RGBA{
			uint8(rand.Intn(256)),
			uint8(rand.Intn(256)),
			uint8(rand.Intn(256)),
			uint8(100 + rand.Intn(156)),
		},
		Projectiles: make([]*Projectile, 0),
	}
}

func NewProjectile(x, y, velocityX, velocityY float64) *Projectile {
	return &Projectile{
		X:         x,
		Y:         y,
		VelocityX: velocityX,
		VelocityY: velocityY,
		Width:     6,
		Height:    6,
		IsActive:  true,
		Lifetime:  4.0,
		Color:     color.RGBA{255, 140, 0, 255},
	}
}

func (e *Enemy) Update(deltaTime float64, playerX, playerY float64, collisionSystem *CollisionSystem) {
	if !e.IsActive {
		return
	}

	switch e.EnemyType {
	case EnemyTypeShooter:
		e.updateShooter(deltaTime, playerX, playerY)
	case EnemyTypeJumper:
		e.updateJumper(deltaTime, playerX, playerY, collisionSystem)
	case EnemyTypeSpike:
	case EnemyTypeGlitched:
		e.updateGlitched(deltaTime, playerX, playerY)
	}

	for i := len(e.Projectiles) - 1; i >= 0; i-- {
		projectile := e.Projectiles[i]
		projectile.Update(deltaTime)
		if !projectile.IsActive {
			e.Projectiles = append(e.Projectiles[:i], e.Projectiles[i+1:]...)
		}
	}
}

func (e *Enemy) updateShooter(deltaTime float64, playerX, playerY float64) {
	if e.ShootTimer > 0 {
		e.ShootTimer -= deltaTime
	}

	if e.MovingRight {
		e.X += e.PatrolSpeed * deltaTime
		if e.X >= e.PatrolEndX {
			e.MovingRight = false
		}
	} else {
		e.X -= e.PatrolSpeed * deltaTime
		if e.X <= e.PatrolStartX {
			e.MovingRight = true
		}
	}
	distanceToPlayer := math.Sqrt(math.Pow(playerX-e.X, 2) + math.Pow(playerY-e.Y, 2))
	if distanceToPlayer < e.DetectionRange && e.ShootTimer <= 0 {
		e.shootAtPlayer(playerX, playerY)
		e.ShootTimer = e.ShootCooldown
	}
}

func (e *Enemy) updateJumper(deltaTime float64, playerX, playerY float64, collisionSystem *CollisionSystem) {
	if e.JumpTimer > 0 {
		e.JumpTimer -= deltaTime
	}

	if !e.OnGround {
		e.VelocityY += 800 * deltaTime
	}

	if e.MovingRight {
		e.VelocityX = e.PatrolSpeed
		if e.X >= e.PatrolEndX {
			e.MovingRight = false
		}
	} else {
		e.VelocityX = -e.PatrolSpeed
		if e.X <= e.PatrolStartX {
			e.MovingRight = true
		}
	}

	if collisionSystem != nil {
		e.applyMovement(deltaTime, collisionSystem)
	} else {
		e.X += e.VelocityX * deltaTime
		e.Y += e.VelocityY * deltaTime
	}
	distanceToPlayer := math.Sqrt(math.Pow(playerX-e.X, 2) + math.Pow(playerY-e.Y, 2))
	if distanceToPlayer < 150 && e.OnGround && e.JumpTimer <= 0 {
		e.VelocityY = e.JumpPower
		e.OnGround = false
		e.JumpTimer = e.JumpCooldown
	}
}

func (e *Enemy) updateGlitched(deltaTime float64, playerX, playerY float64) {
	if rand.Float64() < 0.3 {
		e.MovingRight = !e.MovingRight
	}
	if rand.Float64() < 0.1 {
		e.PatrolSpeed = 20 + rand.Float64()*150
	}
	if rand.Float64() < 0.02 {
		e.X = playerX + (rand.Float64()-0.5)*400
		e.Y = 200 + rand.Float64()*200
	}
	if e.MovingRight {
		e.X += e.PatrolSpeed * deltaTime * (0.5 + rand.Float64())
		if e.X >= e.PatrolEndX || rand.Float64() < 0.05 {
			e.MovingRight = false
		}
	} else {
		e.X -= e.PatrolSpeed * deltaTime * (0.5 + rand.Float64())
		if e.X <= e.PatrolStartX || rand.Float64() < 0.05 {
			e.MovingRight = true
		}
	}
	if rand.Float64() < 0.2 {
		e.Color = color.RGBA{
			uint8(rand.Intn(256)),
			uint8(rand.Intn(256)),
			uint8(rand.Intn(256)),
			uint8(100 + rand.Intn(156)),
		}
	}
	if rand.Float64() < 0.1 {
		e.Width = 10 + rand.Float64()*40
		e.Height = 10 + rand.Float64()*40
	}
	distanceToPlayer := math.Sqrt(math.Pow(playerX-e.X, 2) + math.Pow(playerY-e.Y, 2))
	if distanceToPlayer < e.DetectionRange && rand.Float64() < 0.05 {
		e.shootGlitchedProjectile(playerX, playerY)
	}
}

func (e *Enemy) applyMovement(deltaTime float64, collisionSystem *CollisionSystem) {
	deltaX := e.VelocityX * deltaTime
	deltaY := e.VelocityY * deltaTime

	newX := e.X + deltaX
	if !e.checkCollision(newX, e.Y, collisionSystem) {
		e.X = newX
	} else {
		e.VelocityX = 0
	}

	newY := e.Y + deltaY
	if !e.checkCollision(e.X, newY, collisionSystem) {
		e.Y = newY
		e.OnGround = false
	} else {
		if e.VelocityY > 0 {
			e.VelocityY = 0
			e.OnGround = true
		} else {
			e.VelocityY = 0
		}
	}
}

func (e *Enemy) checkCollision(x, y float64, collisionSystem *CollisionSystem) bool {
	if collisionSystem == nil {
		return false
	}

	box := CollisionBox{
		X: x, Y: y,
		Width: e.Width, Height: e.Height,
	}
	return collisionSystem.CheckCollisionAtPoint(box)
}

func (e *Enemy) shootAtPlayer(targetX, targetY float64) {
	dx := targetX - (e.X + e.Width/2)
	dy := targetY - (e.Y + e.Height/2)
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance > 0 {
		speed := 120.0
		velocityX := (dx / distance) * speed
		velocityY := (dy / distance) * speed

		projectile := NewProjectile(
			e.X+e.Width/2,
			e.Y+e.Height/2,
			velocityX,
			velocityY,
		)

		e.Projectiles = append(e.Projectiles, projectile)
	}
}

func (e *Enemy) shootGlitchedProjectile(playerX, playerY float64) {
	projectileCount := 1 + rand.Intn(4)

	for i := 0; i < projectileCount; i++ {
		angle := math.Atan2(playerY-e.Y, playerX-e.X) + (rand.Float64()-0.5)*math.Pi
		speed := 100 + rand.Float64()*200

		projectile := &Projectile{
			X:         e.X + e.Width/2,
			Y:         e.Y + e.Height/2,
			VelocityX: math.Cos(angle) * speed,
			VelocityY: math.Sin(angle) * speed,
			Width:     4 + rand.Float64()*8,
			Height:    4 + rand.Float64()*8,
			IsActive:  true,
			Lifetime:  2.0 + rand.Float64()*3.0,
			Color: color.RGBA{
				uint8(rand.Intn(256)),
				uint8(rand.Intn(256)),
				uint8(rand.Intn(256)),
				255,
			},
		}

		e.Projectiles = append(e.Projectiles, projectile)
	}
}

func (e *Enemy) Draw(screen *ebiten.Image, camera *Camera) {
	if !e.IsActive {
		return
	}

	screenX, screenY := camera.WorldToScreen(e.X, e.Y)

	switch e.EnemyType {
	case EnemyTypeShooter:
		vector.DrawFilledRect(screen,
			float32(screenX), float32(screenY),
			float32(e.Width), float32(e.Height),
			e.Color, false)
		vector.DrawFilledRect(screen,
			float32(screenX+e.Width), float32(screenY+e.Height/2-2),
			8, 4, color.RGBA{100, 100, 100, 255}, false)
	case EnemyTypeJumper:
		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				dx := float32(i) - 4
				dy := float32(j) - 4
				if dx*dx+dy*dy <= 16 {
					vector.DrawFilledRect(screen,
						float32(screenX)+dx*3, float32(screenY)+dy*3,
						3, 3, e.Color, false)
				}
			}
		}
	case EnemyTypeSpike:
		vector.StrokeLine(screen,
			float32(screenX+e.Width/2), float32(screenY),
			float32(screenX), float32(screenY+e.Height),
			2, e.Color, false)
		vector.StrokeLine(screen,
			float32(screenX+e.Width/2), float32(screenY),
			float32(screenX+e.Width), float32(screenY+e.Height),
			2, e.Color, false)
		vector.StrokeLine(screen,
			float32(screenX), float32(screenY+e.Height),
			float32(screenX+e.Width), float32(screenY+e.Height),
			2, e.Color, false)
	}

	for _, projectile := range e.Projectiles {
		projectile.Draw(screen, camera)
	}
}

func (e *Enemy) GetBounds() (float64, float64, float64, float64) {
	return e.X, e.Y, e.Width, e.Height
}

func (e *Enemy) TakeDamage(damage int) {
	e.Health -= damage
	if e.Health <= 0 {
		e.IsActive = false
	}
}

func (e *Enemy) TakeDamageFromPlayer(damage int) {
	if !e.IsActive {
		return
	}

	e.Health -= damage
	if e.Health <= 0 {
		e.IsActive = false
	}

	if e.EnemyType == EnemyTypeShooter {
		if e.MovingRight {
			e.X -= 10
		} else {
			e.X += 10
		}
	}
}

func (e *Enemy) IsInRange(playerX, playerY float64) bool {
	if !e.IsActive {
		return false
	}

	distance := math.Sqrt(math.Pow(playerX-e.X, 2) + math.Pow(playerY-e.Y, 2))
	return distance <= e.DetectionRange
}

func (e *Enemy) GetDamageDealt() int {
	switch e.EnemyType {
	case EnemyTypeShooter:
		return 1
	case EnemyTypeJumper:
		return 2
	case EnemyTypeSpike:
		return 3
	default:
		return 1
	}
}

func (p *Projectile) Update(deltaTime float64) {
	if !p.IsActive {
		return
	}

	p.X += p.VelocityX * deltaTime
	p.Y += p.VelocityY * deltaTime

	p.Lifetime -= deltaTime
	if p.Lifetime <= 0 {
		p.IsActive = false
	}
}

func (p *Projectile) Draw(screen *ebiten.Image, camera *Camera) {
	if !p.IsActive {
		return
	}

	screenX, screenY := camera.WorldToScreen(p.X, p.Y)

	vector.DrawFilledRect(screen,
		float32(screenX), float32(screenY),
		float32(p.Width), float32(p.Height),
		p.Color, false)
}

func (p *Projectile) GetBounds() (float64, float64, float64, float64) {
	return p.X, p.Y, p.Width, p.Height
}

func (p *Projectile) CheckCollision(x, y, width, height float64) bool {
	if !p.IsActive {
		return false
	}

	return p.X < x+width && p.X+p.Width > x &&
		p.Y < y+height && p.Y+p.Height > y
}
