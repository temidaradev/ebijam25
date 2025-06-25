package src

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type SpecialItemType int

const (
	ItemSchizophrenicFragment SpecialItemType = iota
	ItemRealityGlitch
	ItemMadnessCore
	ItemHarmonyFragment
	ItemStabilityCore
	ItemUnionCrystal
)

type SpecialItem struct {
	X, Y              float64
	Width, Height     float64
	ItemType          SpecialItemType
	IsActive          bool
	Collected         bool
	PulsePhase        float64
	Color             color.RGBA
	GlowColor         color.RGBA
	Name              string
	Description       string
	ParticleSystem    *ParticleSystem
	AuraTimer         float64
	IntensityLevel    float64
	MadnessRadius     float64
	LastParticleSpawn float64
	Health            int
	MaxHealth         int
	HitFlashTimer     float64
	IsBeingHit        bool

	VelocityX     float64
	VelocityY     float64
	OriginalX     float64
	OriginalY     float64
	MovementTimer float64
	MovementType  int
	TeleportTimer float64
	CanTeleport   bool
}

func NewSchizophrenicFragment(x, y float64) *SpecialItem {
	return &SpecialItem{
		X:                 x,
		Y:                 y,
		Width:             24,
		Height:            24,
		ItemType:          ItemSchizophrenicFragment,
		IsActive:          true,
		Collected:         false,
		PulsePhase:        0,
		Color:             color.RGBA{200, 50, 255, 255},
		GlowColor:         color.RGBA{255, 100, 255, 100},
		Name:              "FRAGMENT OF MADNESS",
		Description:       "A shard that breaks reality...",
		ParticleSystem:    NewParticleSystem(10),
		IntensityLevel:    0.4,
		MadnessRadius:     60,
		LastParticleSpawn: 0,
		Health:            8,
		MaxHealth:         8,
		HitFlashTimer:     0,
		IsBeingHit:        false,

		VelocityX:     (rand.Float64() - 0.5) * 20.0,
		VelocityY:     (rand.Float64() - 0.5) * 20.0,
		OriginalX:     x,
		OriginalY:     y,
		MovementTimer: 0,
		MovementType:  rand.Intn(3),
		TeleportTimer: rand.Float64() * 10.0,
		CanTeleport:   true,
	}
}

func NewRealityGlitch(x, y float64) *SpecialItem {
	return &SpecialItem{
		X:                 x,
		Y:                 y,
		Width:             20,
		Height:            20,
		ItemType:          ItemRealityGlitch,
		IsActive:          true,
		Collected:         false,
		PulsePhase:        0,
		Color:             color.RGBA{255, 255, 0, 255},
		GlowColor:         color.RGBA{255, 255, 100, 80},
		Name:              "GLITCH IN THE MATRIX",
		Description:       "ERROR 404: SANITY NOT FOUND",
		ParticleSystem:    NewParticleSystem(15),
		IntensityLevel:    0.6,
		MadnessRadius:     80,
		LastParticleSpawn: 0,
		Health:            6,
		MaxHealth:         6,
		HitFlashTimer:     0,
		IsBeingHit:        false,

		VelocityX:     (rand.Float64() - 0.5) * 40.0,
		VelocityY:     (rand.Float64() - 0.5) * 40.0,
		OriginalX:     x,
		OriginalY:     y,
		MovementTimer: 0,
		MovementType:  rand.Intn(4),
		TeleportTimer: rand.Float64() * 5.0,
		CanTeleport:   true,
	}
}

func NewMadnessCore(x, y float64) *SpecialItem {
	return &SpecialItem{
		X:                 x,
		Y:                 y,
		Width:             32,
		Height:            32,
		ItemType:          ItemMadnessCore,
		IsActive:          true,
		Collected:         false,
		PulsePhase:        0,
		Color:             color.RGBA{255, 0, 0, 255},
		GlowColor:         color.RGBA{255, 50, 50, 120},
		Name:              "CORE OF INSANITY",
		Description:       "THE VOICES ARE GETTING LOUDER...",
		ParticleSystem:    NewParticleSystem(20),
		IntensityLevel:    1.0,
		MadnessRadius:     120,
		LastParticleSpawn: 0,
		Health:            15,
		MaxHealth:         15,
		HitFlashTimer:     0,
		IsBeingHit:        false,

		VelocityX:     (rand.Float64() - 0.5) * 10.0,
		VelocityY:     (rand.Float64() - 0.5) * 10.0,
		OriginalX:     x,
		OriginalY:     y,
		MovementTimer: 0,
		MovementType:  1,
		TeleportTimer: rand.Float64() * 8.0,
		CanTeleport:   true,
	}
}

func NewUnionCrystal(x, y float64) *SpecialItem {
	return &SpecialItem{
		X:                 x,
		Y:                 y,
		Width:             22,
		Height:            22,
		ItemType:          ItemUnionCrystal,
		IsActive:          true,
		Collected:         false,
		PulsePhase:        0,
		Color:             color.RGBA{240, 240, 255, 255},
		GlowColor:         color.RGBA{200, 200, 255, 160},
		Name:              "CRYSTAL OF UNION",
		Description:       "The final piece... Unity of mind and matter",
		ParticleSystem:    NewParticleSystem(10),
		IntensityLevel:    -1.0,
		MadnessRadius:     70,
		LastParticleSpawn: 0,
		Health:            70,
		MaxHealth:         70,
		HitFlashTimer:     0,
		IsBeingHit:        false,

		VelocityX:     0,
		VelocityY:     0,
		OriginalX:     x,
		OriginalY:     y,
		MovementTimer: 0,
		MovementType:  0,
		TeleportTimer: 15.0,
		CanTeleport:   true,
	}
}

func (si *SpecialItem) Update(deltaTime float64) {
	if !si.IsActive || si.Collected {
		return
	}

	if si.HitFlashTimer > 0 {
		si.HitFlashTimer -= deltaTime
		if si.HitFlashTimer <= 0 {
			si.IsBeingHit = false
		}
	}

	si.PulsePhase += deltaTime * 4.0
	si.AuraTimer += deltaTime
	if si.PulsePhase > 2*math.Pi {
		si.PulsePhase -= 2 * math.Pi
	}

	if si.ParticleSystem != nil {
		si.ParticleSystem.Update(deltaTime, si.IntensityLevel)
	}

	si.LastParticleSpawn += deltaTime
	spawnRate := 0.5 / si.IntensityLevel

	if si.LastParticleSpawn > spawnRate {
		si.spawnAuraParticles()
		si.LastParticleSpawn = 0
	}

	if rand.Float64() < 0.02*si.IntensityLevel {
		switch si.ItemType {
		case ItemSchizophrenicFragment:
			si.Color.R = uint8(180 + rand.Intn(50))
			si.Color.G = uint8(rand.Intn(50))
			si.Color.B = uint8(220 + rand.Intn(35))

		case ItemRealityGlitch:
			si.Color.R = uint8(220 + rand.Intn(35))
			si.Color.G = uint8(220 + rand.Intn(35))
			si.Color.B = uint8(rand.Intn(80))

		case ItemMadnessCore:
			si.Color.R = uint8(220 + rand.Intn(35))
			si.Color.G = uint8(rand.Intn(30))
			si.Color.B = uint8(rand.Intn(30))
		default:
		}

	}

	si.UpdateMovement(deltaTime)
}

func (si *SpecialItem) UpdateMovement(deltaTime float64) {
	si.MovementTimer += deltaTime

	switch si.MovementType {
	case 0:
		if si.ItemType == ItemUnionCrystal {
			si.VelocityX = math.Sin(si.MovementTimer) * 5.0
			si.VelocityY = math.Cos(si.MovementTimer) * 5.0
		} else {
			si.VelocityX, si.VelocityY = 0, 0
		}

	case 1:
		if int(si.MovementTimer*10)%20 == 0 {
			si.VelocityX = (rand.Float64()*2 - 1) * 80.0
			si.VelocityY = (rand.Float64()*2 - 1) * 80.0
		}

	case 2:
		angle := si.MovementTimer * 2 * math.Pi / 2
		radius := 50 + 30*math.Sin(si.MovementTimer*2*math.Pi/3)
		si.X = si.OriginalX + math.Cos(angle)*radius
		si.Y = si.OriginalY + math.Sin(angle)*radius
		return

	case 3:
		if int(si.MovementTimer*20)%10 == 0 {
			si.VelocityX = (rand.Float64() - 0.5) * 120.0
			si.VelocityY = (rand.Float64() - 0.5) * 120.0
		}
	}

	si.X += si.VelocityX * deltaTime
	si.Y += si.VelocityY * deltaTime

	maxDist := 200.0
	if si.ItemType == ItemMadnessCore {
		maxDist = 150.0
	}

	distFromOrigin := math.Sqrt(math.Pow(si.X-si.OriginalX, 2) + math.Pow(si.Y-si.OriginalY, 2))
	if distFromOrigin > maxDist {
		si.VelocityX = (si.OriginalX - si.X) * 0.1
		si.VelocityY = (si.OriginalY - si.Y) * 0.1
	}

	if si.CanTeleport {
		si.TeleportTimer += deltaTime
		teleportInterval := 15.0

		switch si.ItemType {
		case ItemRealityGlitch:
			teleportInterval = 8.0
		case ItemMadnessCore:
			teleportInterval = 12.0
		case ItemSchizophrenicFragment:
			teleportInterval = 10.0
		case ItemUnionCrystal:
			teleportInterval = 20.0
		default:
		}

		if si.TeleportTimer > teleportInterval {
			si.Teleport()
			si.TeleportTimer = 0
		}
	}
}

func (si *SpecialItem) Teleport() {
	offsetX := rand.Float64()*200 - 100
	offsetY := rand.Float64()*200 - 100
	si.X = si.OriginalX + offsetX
	si.Y = si.OriginalY + offsetY

	if si.X < -50 {
		si.X = -50
	} else if si.X > 30500 {
		si.X = 30500
	}

	if si.Y < 200 {
		si.Y = 200
	} else if si.Y > 500 {
		si.Y = 500
	}
}

func (si *SpecialItem) spawnAuraParticles() {
	if si.ParticleSystem == nil {
		return
	}

	count := int(si.IntensityLevel*3) + 1

	for i := 0; i < count; i++ {
		angle := rand.Float64() * math.Pi * 2
		distance := 20 + rand.Float64()*si.MadnessRadius

		particleX := si.X + si.Width/2 + math.Cos(angle)*distance
		particleY := si.Y + si.Height/2 + math.Sin(angle)*distance

		switch si.ItemType {
		case ItemSchizophrenicFragment:
			if i%2 == 0 {
				si.ParticleSystem.SpawnParticle(particleX, particleY, ParticleTypeMadness)
			} else {
				si.ParticleSystem.SpawnParticle(particleX, particleY, ParticleTypeHallucinationSpark)
			}

		case ItemRealityGlitch:
			if i%3 == 0 {
				si.ParticleSystem.SpawnParticle(particleX, particleY, ParticleTypeGlitch)
			} else if i%3 == 1 {
				si.ParticleSystem.SpawnParticle(particleX, particleY, ParticleTypeDimensionRip)
			} else {
				si.ParticleSystem.SpawnParticle(particleX, particleY, ParticleTypeMadness)
			}

		case ItemMadnessCore:
			particleTypes := []ParticleType{
				ParticleTypeMadness,
				ParticleTypeGlitch,
				ParticleTypeDimensionRip,
				ParticleTypeChaosOrb,
				ParticleTypeEnergyBeam,
			}

			selectedType := particleTypes[rand.Intn(len(particleTypes))]

			if selectedType == ParticleTypeChaosOrb {
				si.ParticleSystem.SpawnParticle(particleX, particleY, selectedType)
				if len(si.ParticleSystem.Particles) > 0 {
					lastParticle := si.ParticleSystem.Particles[len(si.ParticleSystem.Particles)-1]
					lastParticle.TargetX = si.X + si.Width/2
					lastParticle.TargetY = si.Y + si.Height/2
				}
			} else {
				si.ParticleSystem.SpawnParticle(particleX, particleY, selectedType)
			}

		case ItemHarmonyFragment:
			if i%2 == 0 {
				si.ParticleSystem.SpawnParticle(particleX, particleY, ParticleTypeHealingLight)
			} else {
				si.ParticleSystem.SpawnParticle(particleX, particleY, ParticleTypeHarmonyOrb)
				if len(si.ParticleSystem.Particles) > 0 {
					lastParticle := si.ParticleSystem.Particles[len(si.ParticleSystem.Particles)-1]
					lastParticle.TargetX = si.X + si.Width/2
					lastParticle.TargetY = si.Y + si.Height/2
				}
			}

		case ItemStabilityCore:
			if i%3 == 0 {
				si.ParticleSystem.SpawnParticle(particleX, particleY, ParticleTypeStabilityWave)
			} else if i%3 == 1 {
				si.ParticleSystem.SpawnParticle(particleX, particleY, ParticleTypeRealityRestore)
			} else {
				si.ParticleSystem.SpawnParticle(particleX, particleY, ParticleTypeHealingLight)
			}

		case ItemUnionCrystal:
			unionTypes := []ParticleType{
				ParticleTypeUnionBeam,
				ParticleTypeRealityRestore,
				ParticleTypeHealingLight,
				ParticleTypeStabilityWave,
			}

			selectedType := unionTypes[rand.Intn(len(unionTypes))]
			si.ParticleSystem.SpawnParticle(particleX, particleY, selectedType)

			if selectedType == ParticleTypeUnionBeam && len(si.ParticleSystem.Particles) > 0 {
				lastParticle := si.ParticleSystem.Particles[len(si.ParticleSystem.Particles)-1]
				lastParticle.TargetX = si.X + si.Width/2
				lastParticle.TargetY = si.Y + si.Height/2
			}
		}
	}
}

func (si *SpecialItem) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if !si.IsActive || si.Collected {
		return
	}

	screenX := si.X - cameraX
	screenY := si.Y - cameraY + math.Sin(si.PulsePhase)*3

	if si.ParticleSystem != nil {
		si.ParticleSystem.Draw(screen, cameraX, cameraY)
	}

	for layer := 0; layer < 3; layer++ {
		glowMultiplier := 1.8 + float64(layer)*0.4
		glowSize := si.Width * (glowMultiplier + math.Sin(si.PulsePhase+float64(layer)*math.Pi/3)*0.3)
		glowAlpha := uint8(float64(si.GlowColor.A) / (1.0 + float64(layer)*0.5))

		glowColor := si.GlowColor
		glowColor.A = glowAlpha

		vector.DrawFilledCircle(screen, float32(screenX), float32(screenY), float32(glowSize), glowColor, false)
	}

	itemSize := si.Width * (0.8 + math.Sin(si.PulsePhase)*0.2)

	flashIntensity := 0.7
	if si.IsBeingHit && si.HitFlashTimer > 0 {
		flashIntensity = 1.0 + math.Sin(si.HitFlashTimer*10)*0.2
	}

	flashedColor := si.Color
	if si.IsBeingHit {
		flashedColor.R = uint8(math.Min(255, float64(flashedColor.R)*flashIntensity))
		flashedColor.G = uint8(math.Min(255, float64(flashedColor.G)*flashIntensity))
		flashedColor.B = uint8(math.Min(255, float64(flashedColor.B)*flashIntensity))
	}

	switch si.ItemType {
	case ItemSchizophrenicFragment:
		for i := 0; i < 10; i++ {
			angle := si.PulsePhase + float64(i)*math.Pi/3
			fragmentX := screenX + math.Cos(angle)*itemSize*0.3
			fragmentY := screenY + math.Sin(angle)*itemSize*0.3

			vector.DrawFilledRect(screen,
				float32(fragmentX-itemSize/6), float32(fragmentY-itemSize/6),
				float32(itemSize/3), float32(itemSize/3), flashedColor, false)
		}

	case ItemRealityGlitch:
		for i := 0; i < 5; i++ {
			glitchOffset := (rand.Float64() - 0.5) * 8
			glitchSize := itemSize * (0.7 + 0.3*rand.Float64())

			glitchColor := flashedColor
			if rand.Float64() < 0.1 {
				glitchColor = color.RGBA{220, 220, 220, flashedColor.A}
			}

			vector.DrawFilledRect(screen,
				float32(screenX-glitchSize/2+glitchOffset),
				float32(screenY-glitchSize/2+glitchOffset),
				float32(glitchSize), float32(glitchSize), glitchColor, false)
		}

	case ItemMadnessCore:
		vector.DrawFilledCircle(screen, float32(screenX), float32(screenY), float32(itemSize), flashedColor, false)

		for ring := 1; ring <= 5; ring++ {
			ringRadius := itemSize * (1.2 + float64(ring)*0.4)
			ringRadius *= (1.0 + math.Sin(si.PulsePhase*float64(ring+2))*0.3)

			ringColor := color.RGBA{
				si.Color.R,
				uint8(si.Color.G + 50),
				uint8(si.Color.B + 50),
				uint8(100 / ring),
			}

			vector.StrokeCircle(screen, float32(screenX), float32(screenY), float32(ringRadius), 2, ringColor, false)
		}

	case ItemHarmonyFragment:
		vector.DrawFilledCircle(screen, float32(screenX), float32(screenY), float32(itemSize), si.Color, false)

		for aura := 1; aura <= 5; aura++ {
			auraSize := itemSize * (1.1 + float64(aura)*0.3)
			auraColor := si.GlowColor
			auraColor.A = uint8(float64(si.GlowColor.A) / (1.0 + float64(aura)*0.5))

			vector.DrawFilledCircle(screen, float32(screenX), float32(screenY), float32(auraSize), auraColor, false)
		}

	case ItemStabilityCore:
		for i := 0; i < 6; i++ {
			angle := si.PulsePhase + float64(i)*math.Pi/3
			fragmentX := screenX + math.Cos(angle)*itemSize*0.4
			fragmentY := screenY + math.Sin(angle)*itemSize*0.4

			vector.DrawFilledRect(screen,
				float32(fragmentX-itemSize/8), float32(fragmentY-itemSize/8),
				float32(itemSize/4), float32(itemSize/4), si.Color, false)
		}

	case ItemUnionCrystal:
		vector.DrawFilledCircle(screen, float32(screenX), float32(screenY), float32(itemSize), si.Color, false)

		for beam := 0; beam < 8; beam++ {
			beamAngle := si.PulsePhase + float64(beam)*math.Pi/4
			beamLength := itemSize * (1.5 + 0.5*math.Sin(si.PulsePhase*2+float64(beam)))

			vector.StrokeLine(screen,
				float32(screenX), float32(screenY),
				float32(screenX+math.Cos(beamAngle)*beamLength), float32(screenY+math.Sin(beamAngle)*beamLength),
				2, si.Color, false)
		}

	default:
		vector.DrawFilledRect(screen, float32(screenX-itemSize/2), float32(screenY-itemSize/2), float32(itemSize), float32(itemSize), si.Color, false)
	}

	sparkleCount := int(8 * si.IntensityLevel)
	for i := 0; i < sparkleCount; i++ {
		sparkleAngle := si.PulsePhase + float64(i)*math.Pi*2/float64(sparkleCount)
		sparkleDistance := si.Width * (1.2 + 0.5*math.Sin(si.AuraTimer*2+float64(i)))
		sparkleX := screenX + math.Cos(sparkleAngle)*sparkleDistance
		sparkleY := screenY + math.Sin(sparkleAngle)*sparkleDistance

		sparkleIntensity := math.Sin(si.PulsePhase*3 + float64(i))
		sparkleColor := color.RGBA{
			255,
			255,
			255,
			uint8(50 + 205*math.Abs(sparkleIntensity)),
		}

		sparkleSize := 1.5 + 2*math.Abs(sparkleIntensity)
		vector.DrawFilledCircle(screen, float32(sparkleX), float32(sparkleY), float32(sparkleSize), sparkleColor, false)
	}

	if si.IntensityLevel > 0.7 && math.Sin(si.AuraTimer*4) > 0.8 {
		for i := 0; i < 5; i++ {
			tearAngle := rand.Float64() * math.Pi * 2
			tearDistance := rand.Float64() * si.MadnessRadius
			tearX := screenX + math.Cos(tearAngle)*tearDistance
			tearY := screenY + math.Sin(tearAngle)*tearDistance

			tearColor := color.RGBA{200, 200, 200, 80}
			vector.StrokeLine(screen,
				float32(tearX-5), float32(tearY),
				float32(tearX+5), float32(tearY),
				1, tearColor, false)
		}
	}

	if si.MaxHealth > 1 {
		healthBarWidth := float32(si.Width)
		healthBarHeight := float32(4)
		healthBarY := float32(screenY - si.Height/2 - 10)

		vector.DrawFilledRect(screen,
			float32(screenX-si.Width/2), healthBarY,
			healthBarWidth, healthBarHeight,
			color.RGBA{100, 100, 100, 200}, false)

		healthPercentage := si.GetHealthPercentage()
		fillWidth := healthBarWidth * float32(healthPercentage)

		healthColor := color.RGBA{255, 100, 100, 255}
		if si.ItemType >= ItemHarmonyFragment {
			healthColor = color.RGBA{100, 255, 150, 255}
		}

		vector.DrawFilledRect(screen,
			float32(screenX-si.Width/2), healthBarY,
			fillWidth, healthBarHeight,
			healthColor, false)
	}
}

func (si *SpecialItem) CheckCollision(playerX, playerY, playerW, playerH float64) bool {
	if !si.IsActive || si.Collected {
		return false
	}

	return playerX < si.X+si.Width &&
		playerX+playerW > si.X &&
		playerY < si.Y+si.Height &&
		playerY+playerH > si.Y
}

func (si *SpecialItem) Collect() {
	si.Collected = true
	si.IsActive = false
}

func (si *SpecialItem) TakeHit() bool {
	if !si.IsActive || si.Collected {
		return false
	}

	si.Health--
	si.IsBeingHit = true
	si.HitFlashTimer = 0.05

	if si.ParticleSystem != nil {
		centerX := si.X + si.Width/2
		centerY := si.Y + si.Height/2

		switch si.ItemType {
		case ItemSchizophrenicFragment, ItemRealityGlitch, ItemMadnessCore:
			si.ParticleSystem.SpawnBurst(centerX, centerY, ParticleTypeMadness, 5)
		case ItemHarmonyFragment, ItemStabilityCore:
			si.ParticleSystem.SpawnBurst(centerX, centerY, ParticleTypeHealingLight, 3)
		case ItemUnionCrystal:
			si.ParticleSystem.SpawnBurst(centerX, centerY, ParticleTypeUnionBeam, 4)
		}
	}

	if si.CanTeleport && si.Health > 0 {
		teleportChance := 1.0 - (float64(si.Health)/float64(si.MaxHealth))*0.5
		if rand.Float64() < teleportChance {
			si.Teleport()
			si.TeleportTimer = 0
		}
	}

	if si.Health <= 0 {
		si.Collected = true
		si.IsActive = false
		return true
	}

	return false
}

func (si *SpecialItem) CheckHitCollision(attackX, attackY, attackW, attackH float64) bool {
	if !si.IsActive || si.Collected {
		return false
	}

	return attackX < si.X+si.Width &&
		attackX+attackW > si.X &&
		attackY < si.Y+si.Height &&
		attackY+attackH > si.Y
}

func (si *SpecialItem) GetHealthPercentage() float64 {
	if si.MaxHealth == 0 {
		return 0
	}
	return float64(si.Health) / float64(si.MaxHealth)
}
