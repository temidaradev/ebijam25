package src

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type ParticleType int

const (
	ParticleTypeMadness ParticleType = iota
	ParticleTypeGlitch
	ParticleTypeEnergyBeam
	ParticleTypeDimensionRip
	ParticleTypeHallucinationSpark
	ParticleTypeChaosOrb
	ParticleTypeHealingLight
	ParticleTypeStabilityWave
	ParticleTypeHarmonyOrb
	ParticleTypeUnionBeam
	ParticleTypeRealityRestore
)

type Particle struct {
	X, Y           float64
	VelocityX      float64
	VelocityY      float64
	Size           float64
	Life           float64
	MaxLife        float64
	Color          color.RGBA
	ParticleType   ParticleType
	RotationSpeed  float64
	Rotation       float64
	Scale          float64
	Alpha          uint8
	TargetX        float64
	TargetY        float64
	IsAiming       bool
	AimStrength    float64
	TrailLength    int
	TrailPositions []struct{ X, Y float64 }
}

type ParticleSystem struct {
	Particles      []*Particle
	MaxParticles   int
	MadnessLevel   float64
	ScreenShakeX   float64
	ScreenShakeY   float64
	GlitchTimer    float64
	ChaosIntensity float64
}

func NewParticleSystem(maxParticles int) *ParticleSystem {
	return &ParticleSystem{
		Particles:    make([]*Particle, 0),
		MaxParticles: maxParticles,
	}
}

func (ps *ParticleSystem) Update(deltaTime float64, madnessLevel float64) {
	ps.MadnessLevel = madnessLevel
	ps.GlitchTimer += deltaTime
	ps.ChaosIntensity = madnessLevel * (1.0 + 0.5*math.Sin(ps.GlitchTimer*5.0))

	shakeIntensity := madnessLevel * 8.0
	ps.ScreenShakeX = (rand.Float64() - 0.5) * shakeIntensity
	ps.ScreenShakeY = (rand.Float64() - 0.5) * shakeIntensity

	for i := len(ps.Particles) - 1; i >= 0; i-- {
		particle := ps.Particles[i]
		ps.updateParticle(particle, deltaTime)

		if particle.Life <= 0 {
			ps.Particles = append(ps.Particles[:i], ps.Particles[i+1:]...)
		}
	}

	if madnessLevel > 0.2 && len(ps.Particles) < ps.MaxParticles {
		if rand.Float64() < madnessLevel*0.3 {
			ps.SpawnAmbientMadnessParticles()
		}
	}
}

func (ps *ParticleSystem) updateParticle(p *Particle, deltaTime float64) {
	p.Life -= deltaTime
	lifeRatio := p.Life / p.MaxLife

	p.Alpha = uint8(200 * lifeRatio)

	p.Rotation += p.RotationSpeed * deltaTime

	if p.IsAiming && p.TargetX != 0 && p.TargetY != 0 {
		dx := p.TargetX - p.X
		dy := p.TargetY - p.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance > 10 {
			aimForce := p.AimStrength * deltaTime
			p.VelocityX += (dx / distance) * aimForce
			p.VelocityY += (dy / distance) * aimForce

			speed := math.Sqrt(p.VelocityX*p.VelocityX + p.VelocityY*p.VelocityY)
			if speed > 300 {
				p.VelocityX = (p.VelocityX / speed) * 300
				p.VelocityY = (p.VelocityY / speed) * 300
			}
		}
	}

	switch p.ParticleType {
	case ParticleTypeMadness:
		p.VelocityX += (rand.Float64() - 0.5) * 100 * ps.ChaosIntensity * deltaTime
		p.VelocityY += (rand.Float64() - 0.5) * 100 * ps.ChaosIntensity * deltaTime

		if rand.Float64() < 0.2 {
			p.Color.R = uint8(100 + rand.Intn(156))
			p.Color.G = uint8(rand.Intn(100))
			p.Color.B = uint8(150 + rand.Intn(106))
		}
	case ParticleTypeGlitch:
		if rand.Float64() < 0.05*ps.ChaosIntensity {
			p.X += (rand.Float64() - 0.5) * 100
			p.Y += (rand.Float64() - 0.5) * 100
		}

		p.Size = 2 + 8*math.Sin(ps.GlitchTimer*20+p.Rotation)
	case ParticleTypeDimensionRip:
		p.VelocityX *= 0.95
		p.VelocityY *= 0.95
		p.Size += 10 * deltaTime

		if rand.Float64() < 0.3 && len(ps.Particles) < ps.MaxParticles-5 {
			ps.SpawnChildParticle(p)
		}
	case ParticleTypeChaosOrb:
		centerX := p.TargetX
		centerY := p.TargetY
		if centerX != 0 && centerY != 0 {
			angle := math.Atan2(p.Y-centerY, p.X-centerX)
			angle += deltaTime * 2.0
			radius := 50 + 30*math.Sin(ps.GlitchTimer*3+p.Rotation)
			p.X = centerX + math.Cos(angle)*radius
			p.Y = centerY + math.Sin(angle)*radius
		}

	case ParticleTypeEnergyBeam:
		if p.IsAiming {
			dx := p.TargetX - p.X
			dy := p.TargetY - p.Y
			distance := math.Sqrt(dx*dx + dy*dy)

			if distance > 5 {
				aimStrength := 200 * deltaTime
				p.VelocityX += (dx / distance) * aimStrength
				p.VelocityY += (dy / distance) * aimStrength
			}
		}

	case ParticleTypeHealingLight:
		p.VelocityY -= 50 * deltaTime
		p.VelocityX *= 0.98

		p.Size = 3 + 2*math.Sin(ps.GlitchTimer*3+p.Rotation)

	case ParticleTypeStabilityWave:
		p.VelocityX *= 0.95
		p.VelocityY *= 0.95
		p.Size += 15 * deltaTime

		if rand.Float64() < 0.1 {
			ps.stabilizeNearbyParticles(p.X, p.Y, 50)
		}

	case ParticleTypeHarmonyOrb:
		if p.TargetX != 0 && p.TargetY != 0 {
			angle := math.Atan2(p.Y-p.TargetY, p.X-p.TargetX)
			angle += deltaTime * 1.0
			radius := 40 + 20*math.Sin(ps.GlitchTimer*2+p.Rotation)
			p.X = p.TargetX + math.Cos(angle)*radius
			p.Y = p.TargetY + math.Sin(angle)*radius
		}

	case ParticleTypeUnionBeam:
		if p.IsAiming {
			dx := p.TargetX - p.X
			dy := p.TargetY - p.Y
			distance := math.Sqrt(dx*dx + dy*dy)

			if distance > 10 {
				aimStrength := 300 * deltaTime
				p.VelocityX += (dx / distance) * aimStrength
				p.VelocityY += (dy / distance) * aimStrength
			}
		}

	case ParticleTypeRealityRestore:
		p.VelocityX *= 0.9
		p.VelocityY *= 0.9

		if rand.Float64() < 0.2 && len(ps.Particles) < ps.MaxParticles-3 {
			ps.SpawnHealingAura(p.X, p.Y)
		}
	default:
	}

	if p.TrailLength > 0 {
		if len(p.TrailPositions) > p.TrailLength {
			p.TrailPositions = p.TrailPositions[1:]
		}
		p.TrailPositions = append(p.TrailPositions, struct{ X, Y float64 }{p.X, p.Y})
	}

	p.X += p.VelocityX * deltaTime
	p.Y += p.VelocityY * deltaTime

	p.Scale = 0.5 + 1.5*lifeRatio
}

func (ps *ParticleSystem) SpawnParticle(x, y float64, particleType ParticleType) {
	if len(ps.Particles) >= ps.MaxParticles {
		return
	}

	particle := &Particle{
		X:            x,
		Y:            y,
		ParticleType: particleType,
		Rotation:     rand.Float64() * math.Pi * 2,
		Scale:        1.0,
	}

	switch particleType {
	case ParticleTypeMadness:
		particle.VelocityX = (rand.Float64() - 0.5) * 200
		particle.VelocityY = (rand.Float64() - 0.5) * 200
		particle.Size = 3 + rand.Float64()*8
		particle.Life = 2.0 + rand.Float64()*3.0
		particle.MaxLife = particle.Life
		particle.Color = color.RGBA{200, 50, 255, 255}
		particle.RotationSpeed = (rand.Float64() - 0.5) * 10
		particle.TrailLength = 5

	case ParticleTypeGlitch:
		particle.VelocityX = (rand.Float64() - 0.5) * 400
		particle.VelocityY = (rand.Float64() - 0.5) * 400
		particle.Size = 2 + rand.Float64()*6
		particle.Life = 1.0 + rand.Float64()*2.0
		particle.MaxLife = particle.Life
		particle.Color = color.RGBA{255, 255, 0, 255}
		particle.RotationSpeed = (rand.Float64() - 0.5) * 20

	case ParticleTypeDimensionRip:
		particle.VelocityX = (rand.Float64() - 0.5) * 50
		particle.VelocityY = (rand.Float64() - 0.5) * 50
		particle.Size = 20 + rand.Float64()*30
		particle.Life = 5.0 + rand.Float64()*3.0
		particle.MaxLife = particle.Life
		particle.Color = color.RGBA{120, 0, 120, 180}
		particle.RotationSpeed = (rand.Float64() - 0.5) * 5

	case ParticleTypeChaosOrb:
		particle.VelocityX = 0
		particle.VelocityY = 0
		particle.Size = 15 + rand.Float64()*20
		particle.Life = 4.0 + rand.Float64()*4.0
		particle.MaxLife = particle.Life
		particle.Color = color.RGBA{255, 100, 100, 200}
		particle.RotationSpeed = 3.0
		particle.TrailLength = 8

	case ParticleTypeEnergyBeam:
		angle := rand.Float64() * math.Pi * 2
		speed := 100 + rand.Float64()*150
		particle.VelocityX = math.Cos(angle) * speed
		particle.VelocityY = math.Sin(angle) * speed
		particle.Size = 4 + rand.Float64()*8
		particle.Life = 3.0 + rand.Float64()*2.0
		particle.MaxLife = particle.Life
		particle.Color = color.RGBA{0, 255, 255, 255}
		particle.IsAiming = false
		particle.AimStrength = 150
		particle.TrailLength = 10

	case ParticleTypeHallucinationSpark:
		particle.VelocityX = (rand.Float64() - 0.5) * 300
		particle.VelocityY = (rand.Float64() - 0.5) * 300
		particle.Size = 1 + rand.Float64()*4
		particle.Life = 0.5 + rand.Float64()*1.5
		particle.MaxLife = particle.Life
		particle.Color = color.RGBA{255, 150, 150, 200}
		particle.RotationSpeed = (rand.Float64() - 0.5) * 30

	case ParticleTypeHealingLight:
		particle.VelocityX = (rand.Float64() - 0.5) * 50
		particle.VelocityY = -30 - rand.Float64()*50
		particle.Size = 2 + rand.Float64()*4
		particle.Life = 3.0 + rand.Float64()*2.0
		particle.MaxLife = particle.Life
		particle.Color = color.RGBA{150, 255, 200, 255}
		particle.RotationSpeed = (rand.Float64() - 0.5) * 5
		particle.TrailLength = 5

	case ParticleTypeStabilityWave:
		particle.VelocityX = (rand.Float64() - 0.5) * 30
		particle.VelocityY = (rand.Float64() - 0.5) * 30
		particle.Size = 8 + rand.Float64()*12
		particle.Life = 4.0 + rand.Float64()*2.0
		particle.MaxLife = particle.Life
		particle.Color = color.RGBA{100, 200, 255, 180}
		particle.RotationSpeed = (rand.Float64() - 0.5) * 3

	case ParticleTypeHarmonyOrb:
		particle.VelocityX = 0
		particle.VelocityY = 0
		particle.Size = 10 + rand.Float64()*15
		particle.Life = 5.0 + rand.Float64()*3.0
		particle.MaxLife = particle.Life
		particle.Color = color.RGBA{200, 255, 150, 200}
		particle.RotationSpeed = 2.0
		particle.TrailLength = 6

	case ParticleTypeUnionBeam:
		angle := rand.Float64() * math.Pi * 2
		speed := 80 + rand.Float64()*120
		particle.VelocityX = math.Cos(angle) * speed
		particle.VelocityY = math.Sin(angle) * speed
		particle.Size = 3 + rand.Float64()*6
		particle.Life = 4.0 + rand.Float64()*3.0
		particle.MaxLife = particle.Life
		particle.Color = color.RGBA{220, 200, 255, 150}
		particle.IsAiming = true
		particle.AimStrength = 200
		particle.TrailLength = 8

	case ParticleTypeRealityRestore:
		particle.VelocityX = (rand.Float64() - 0.5) * 40
		particle.VelocityY = (rand.Float64() - 0.5) * 40
		particle.Size = 6 + rand.Float64()*10
		particle.Life = 6.0 + rand.Float64()*4.0
		particle.MaxLife = particle.Life
		particle.Color = color.RGBA{200, 230, 255, 120}
		particle.RotationSpeed = (rand.Float64() - 0.5) * 4
		particle.TrailLength = 10
	}

	ps.Particles = append(ps.Particles, particle)
}

func (ps *ParticleSystem) SpawnAimedParticle(x, y, targetX, targetY float64, particleType ParticleType) {
	ps.SpawnParticle(x, y, particleType)

	if len(ps.Particles) > 0 {
		particle := ps.Particles[len(ps.Particles)-1]
		particle.TargetX = targetX
		particle.TargetY = targetY
		particle.IsAiming = true

		dx := targetX - x
		dy := targetY - y
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance > 0 {
			initialSpeed := 80.0
			particle.VelocityX = (dx / distance) * initialSpeed
			particle.VelocityY = (dy / distance) * initialSpeed
		}
	}
}

func (ps *ParticleSystem) SpawnBurst(x, y float64, particleType ParticleType, count int) {
	for i := 0; i < count; i++ {
		offsetX := x + (rand.Float64()-0.5)*20
		offsetY := y + (rand.Float64()-0.5)*20
		ps.SpawnParticle(offsetX, offsetY, particleType)
	}
}

func (ps *ParticleSystem) SpawnAmbientMadnessParticles() {
	for i := 0; i < 3; i++ {
		x := rand.Float64() * 1280
		y := rand.Float64() * 720

		particleTypes := []ParticleType{
			ParticleTypeMadness,
			ParticleTypeGlitch,
			ParticleTypeHallucinationSpark,
		}

		particleType := particleTypes[rand.Intn(len(particleTypes))]
		ps.SpawnParticle(x, y, particleType)
	}
}

func (ps *ParticleSystem) SpawnChildParticle(parent *Particle) {
	child := &Particle{
		X:            parent.X + (rand.Float64()-0.5)*20,
		Y:            parent.Y + (rand.Float64()-0.5)*20,
		VelocityX:    (rand.Float64() - 0.5) * 100,
		VelocityY:    (rand.Float64() - 0.5) * 100,
		Size:         parent.Size * 0.3,
		Life:         parent.Life * 0.5,
		MaxLife:      parent.Life * 0.5,
		Color:        parent.Color,
		ParticleType: ParticleTypeHallucinationSpark,
		Scale:        0.5,
	}

	ps.Particles = append(ps.Particles, child)
}

func (ps *ParticleSystem) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	for _, particle := range ps.Particles {
		ps.drawParticle(screen, particle, cameraX, cameraY)
	}
}

func (ps *ParticleSystem) drawParticle(screen *ebiten.Image, p *Particle, cameraX, cameraY float64) {
	screenX := float32(p.X - cameraX + float64(ps.ScreenShakeX))
	screenY := float32(p.Y - cameraY + float64(ps.ScreenShakeY))

	if ps.MadnessLevel > 0.5 {
		distortion := ps.MadnessLevel * 10
		screenX += float32((rand.Float64() - 0.5) * distortion)
		screenY += float32((rand.Float64() - 0.5) * distortion)
	}

	if len(p.TrailPositions) > 1 {
		for i := 0; i < len(p.TrailPositions)-1; i++ {
			trailAlpha := uint8(float64(p.Alpha) * (float64(i) / float64(len(p.TrailPositions))))
			trailColor := p.Color
			trailColor.A = trailAlpha

			trailX1 := float32(p.TrailPositions[i].X - cameraX)
			trailY1 := float32(p.TrailPositions[i].Y - cameraY)
			trailX2 := float32(p.TrailPositions[i+1].X - cameraX)
			trailY2 := float32(p.TrailPositions[i+1].Y - cameraY)

			vector.StrokeLine(screen, trailX1, trailY1, trailX2, trailY2, 1, trailColor, false)
		}
	}

	particleColor := p.Color
	particleColor.A = p.Alpha

	size := float32(p.Size * p.Scale)

	switch p.ParticleType {
	case ParticleTypeMadness:
		vector.DrawFilledRect(screen, screenX-size/2, screenY-size/2, size, size, particleColor, false)

		glitchColor := color.RGBA{150, 150, 255, p.Alpha / 4}
		for i := 0; i < 4; i++ {
			offset := float32((rand.Float64() - 0.5) * 4)
			vector.StrokeRect(screen, screenX-size/2+offset, screenY-size/2+offset, size, size, 1, glitchColor, false)
		}

	case ParticleTypeGlitch:
		w := size * (0.5 + 0.5*float32(math.Sin(ps.GlitchTimer*20+p.Rotation)))
		h := size * (0.5 + 0.5*float32(math.Cos(ps.GlitchTimer*25+p.Rotation)))
		vector.DrawFilledRect(screen, screenX-w/2, screenY-h/2, w, h, particleColor, false)

	case ParticleTypeDimensionRip:
		vector.DrawFilledCircle(screen, screenX, screenY, size, particleColor, false)

		innerColor := color.RGBA{0, 0, 0, p.Alpha}
		vector.DrawFilledCircle(screen, screenX, screenY, size*0.6, innerColor, false)

	case ParticleTypeChaosOrb:
		vector.DrawFilledCircle(screen, screenX, screenY, size, particleColor, false)

		spikeColor := color.RGBA{255, 200, 100, p.Alpha / 4}
		for i := 0; i < 8; i++ {
			angle := p.Rotation + float64(i)*math.Pi/4
			spikeX := screenX + float32(math.Cos(angle)*float64(size*1.5))
			spikeY := screenY + float32(math.Sin(angle)*float64(size*1.5))
			vector.StrokeLine(screen, screenX, screenY, spikeX, spikeY, 2, spikeColor, false)
		}

	case ParticleTypeEnergyBeam:
		vector.DrawFilledCircle(screen, screenX, screenY, size, particleColor, false)

		coreColor := color.RGBA{255, 255, 150, p.Alpha / 2}
		vector.DrawFilledCircle(screen, screenX, screenY, size*0.3, coreColor, false)

	case ParticleTypeHallucinationSpark:
		vector.DrawFilledCircle(screen, screenX, screenY, size, particleColor, false)

	case ParticleTypeHealingLight:
		vector.DrawFilledCircle(screen, screenX, screenY, size, particleColor, false)

		glowColor := color.RGBA{150, 255, 200, p.Alpha / 6}
		vector.DrawFilledCircle(screen, screenX, screenY, size*2, glowColor, false)

	case ParticleTypeStabilityWave:
		vector.StrokeCircle(screen, screenX, screenY, size, 2, particleColor, false)

		innerColor := particleColor
		innerColor.A = p.Alpha / 2
		vector.StrokeCircle(screen, screenX, screenY, size*0.7, 1, innerColor, false)

	case ParticleTypeHarmonyOrb:
		vector.DrawFilledCircle(screen, screenX, screenY, size, particleColor, false)

		for _, other := range ps.Particles {
			if other != p && other.ParticleType == ParticleTypeHarmonyOrb {
				distance := math.Sqrt(math.Pow(other.X-p.X, 2) + math.Pow(other.Y-p.Y, 2))
				if distance < 100 {
					orbX := other.X - cameraX + float64(ps.ScreenShakeX)
					orbY := other.Y - cameraY + float64(ps.ScreenShakeY)

					lineColor := color.RGBA{200, 255, 150, uint8(p.Alpha / 4)}
					vector.StrokeLine(screen, screenX, screenY, float32(orbX), float32(orbY), 1, lineColor, false)
				}
			}
		}

	case ParticleTypeUnionBeam:
		vector.DrawFilledCircle(screen, screenX, screenY, size, particleColor, false)

		if p.IsAiming && p.TargetX != 0 && p.TargetY != 0 {
			targetScreenX := float32(p.TargetX - cameraX)
			targetScreenY := float32(p.TargetY - cameraY)

			beamColor := color.RGBA{200, 180, 255, p.Alpha / 3}
			vector.StrokeLine(screen, screenX, screenY, targetScreenX, targetScreenY, 2, beamColor, false)
		}

	case ParticleTypeRealityRestore:
		vector.DrawFilledCircle(screen, screenX, screenY, size, particleColor, false)

		crossSize := size * 1.5
		crossColor := color.RGBA{180, 220, 255, p.Alpha / 2}
		vector.StrokeLine(screen, screenX-crossSize, screenY, screenX+crossSize, screenY, 2, crossColor, false)
		vector.StrokeLine(screen, screenX, screenY-crossSize, screenX, screenY+crossSize, 2, crossColor, false)
	}
}

func (ps *ParticleSystem) Clear() {
	ps.Particles = ps.Particles[:0]
}

func (ps *ParticleSystem) GetParticleCount() int {
	return len(ps.Particles)
}

func (ps *ParticleSystem) stabilizeNearbyParticles(x, y, radius float64) {
	for _, particle := range ps.Particles {
		dx := particle.X - x
		dy := particle.Y - y
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance < radius {
			particle.VelocityX *= 0.8
			particle.VelocityY *= 0.8

			if particle.ParticleType == ParticleTypeMadness || particle.ParticleType == ParticleTypeGlitch {
				particle.ParticleType = ParticleTypeStabilityWave
				particle.Color = color.RGBA{100, 200, 255, particle.Color.A}
			}
		}
	}
}

func (ps *ParticleSystem) SpawnHealingAura(x, y float64) {
	for i := 0; i < 3; i++ {
		angle := rand.Float64() * math.Pi * 2
		distance := 10 + rand.Float64()*20

		auraX := x + math.Cos(angle)*distance
		auraY := y + math.Sin(angle)*distance

		ps.SpawnParticle(auraX, auraY, ParticleTypeHealingLight)
	}
}
