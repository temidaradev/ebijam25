package src

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type EndingState int

const (
	EndingStateFragmentsRising EndingState = iota
	EndingStateFormingCrystal
	EndingStateCrystalComplete
	EndingStateFormingUnion
	EndingStateUnionComplete
	EndingStateComplete
)

type CrystalFragmentParticle struct {
	X, Y                 float64
	TargetX, TargetY     float64
	StartX, StartY       float64
	VelocityX, VelocityY float64
	Color                color.RGBA
	Size                 float64
	Progress             float64
	DelayTimer           float64
	PulsePhase           float64
	Alpha                float64
	CrystalLayer         int
}

type EndingAnimation struct {
	State           EndingState
	Timer           float64
	TotalDuration   float64
	Fragments       []CrystalFragmentParticle
	UnionFragments  []CrystalFragmentParticle
	CrystalCenterX  float64
	CrystalCenterY  float64
	CrystalScale    float64
	FadeAlpha       float64
	IsActive        bool
	GameCloseTimer  float64
	ScreenShakeX    float64
	ScreenShakeY    float64
	WhiteFlashAlpha float64
	MusicFadeVolume float64
	CrystalPoints   []struct {
		X, Y  float64
		Layer int
	}
	UnionPoints []struct{ X, Y float64 }
}

func NewEndingAnimation(screenWidth, screenHeight int) *EndingAnimation {
	centerX := float64(screenWidth) / 2
	centerY := float64(screenHeight) / 2

	ending := &EndingAnimation{
		State:           EndingStateFragmentsRising,
		Timer:           0,
		TotalDuration:   20.0,
		CrystalCenterX:  centerX,
		CrystalCenterY:  centerY - 50,
		CrystalScale:    80.0,
		IsActive:        false,
		GameCloseTimer:  0,
		MusicFadeVolume: 1.0,
	}

	ending.generateCrystalPoints()
	ending.generateUnionPoints(screenWidth)
	ending.createFragments(screenWidth, screenHeight)

	return ending
}

func (ea *EndingAnimation) generateCrystalPoints() {
	ea.CrystalPoints = make([]struct {
		X, Y  float64
		Layer int
	}, 0, 60)

	corePoints := []struct{ X, Y float64 }{
		{0, -0.8},
		{0.4, -0.4},
		{0.6, 0},
		{0.4, 0.4},
		{0, 0.8},
		{-0.4, 0.4},
		{-0.6, 0},
		{-0.4, -0.4},
	}

	for _, point := range corePoints {
		ea.CrystalPoints = append(ea.CrystalPoints, struct {
			X, Y  float64
			Layer int
		}{
			point.X, point.Y, 0,
		})
	}

	for i := 0; i < 12; i++ {
		angle := float64(i) * 2 * math.Pi / 12
		radius := 0.9
		x := radius * math.Cos(angle)
		y := radius * math.Sin(angle)
		ea.CrystalPoints = append(ea.CrystalPoints, struct {
			X, Y  float64
			Layer int
		}{
			x, y, 1,
		})
	}

	for i := 0; i < 16; i++ {
		angle := float64(i) * 2 * math.Pi / 16
		radius := 1.2
		x := radius * math.Cos(angle)
		y := radius * math.Sin(angle)
		ea.CrystalPoints = append(ea.CrystalPoints, struct {
			X, Y  float64
			Layer int
		}{
			x, y, 2,
		})
	}

	for i := 0; i < 24; i++ {
		angle := float64(i) * 2 * math.Pi / 24
		radius := 1.5 + 0.3*math.Sin(float64(i)*0.5)
		x := radius * math.Cos(angle)
		y := radius * math.Sin(angle)
		ea.CrystalPoints = append(ea.CrystalPoints, struct {
			X, Y  float64
			Layer int
		}{
			x, y, 3,
		})
	}
}

func (ea *EndingAnimation) generateUnionPoints(screenWidth int) {
	ea.UnionPoints = make([]struct{ X, Y float64 }, 0, 80)

	textCenterX := float64(screenWidth) / 2
	textCenterY := ea.CrystalCenterY + 120

	letterSpacing := 60.0
	letterWidth := 25.0
	letterHeight := 35.0

	startX := textCenterX - (4.5*letterSpacing)/2

	ea.addLetterU(startX, textCenterY, letterWidth, letterHeight)
	ea.addLetterN(startX+letterSpacing, textCenterY, letterWidth, letterHeight)
	ea.addLetterI(startX+2*letterSpacing, textCenterY, letterWidth, letterHeight)
	ea.addLetterO(startX+3*letterSpacing, textCenterY, letterWidth, letterHeight)
	ea.addLetterN(startX+4*letterSpacing, textCenterY, letterWidth, letterHeight)
}

func (ea *EndingAnimation) addLetterU(x, y, w, h float64) {
	for i := 0; i < 6; i++ {
		ea.UnionPoints = append(ea.UnionPoints, struct{ X, Y float64 }{
			x - w/2, y - h/2 + float64(i)*h/6,
		})
	}
	for i := 1; i < 4; i++ {
		ea.UnionPoints = append(ea.UnionPoints, struct{ X, Y float64 }{
			x - w/2 + float64(i)*w/4, y + h/2,
		})
	}
	for i := 0; i < 6; i++ {
		ea.UnionPoints = append(ea.UnionPoints, struct{ X, Y float64 }{
			x + w/2, y - h/2 + float64(i)*h/6,
		})
	}
}

func (ea *EndingAnimation) addLetterN(x, y, w, h float64) {
	for i := 0; i < 7; i++ {
		ea.UnionPoints = append(ea.UnionPoints, struct{ X, Y float64 }{
			x - w/2, y - h/2 + float64(i)*h/6,
		})
	}
	for i := 0; i < 6; i++ {
		ea.UnionPoints = append(ea.UnionPoints, struct{ X, Y float64 }{
			x - w/2 + float64(i)*w/5, y - h/2 + float64(i)*h/5,
		})
	}
	for i := 0; i < 7; i++ {
		ea.UnionPoints = append(ea.UnionPoints, struct{ X, Y float64 }{
			x + w/2, y - h/2 + float64(i)*h/6,
		})
	}
}

func (ea *EndingAnimation) addLetterI(x, y, w, h float64) {
	for i := 0; i < 5; i++ {
		ea.UnionPoints = append(ea.UnionPoints, struct{ X, Y float64 }{
			x - w/2 + float64(i)*w/4, y - h/2,
		})
	}
	for i := 1; i < 6; i++ {
		ea.UnionPoints = append(ea.UnionPoints, struct{ X, Y float64 }{
			x, y - h/2 + float64(i)*h/6,
		})
	}
	for i := 0; i < 5; i++ {
		ea.UnionPoints = append(ea.UnionPoints, struct{ X, Y float64 }{
			x - w/2 + float64(i)*w/4, y + h/2,
		})
	}
}

func (ea *EndingAnimation) addLetterO(x, y, w, h float64) {
	for i := 1; i < 4; i++ {
		ea.UnionPoints = append(ea.UnionPoints, struct{ X, Y float64 }{
			x - w/2 + float64(i)*w/4, y - h/2,
		})
	}
	for i := 1; i < 6; i++ {
		ea.UnionPoints = append(ea.UnionPoints, struct{ X, Y float64 }{
			x - w/2, y - h/2 + float64(i)*h/6,
		})
	}
	for i := 1; i < 4; i++ {
		ea.UnionPoints = append(ea.UnionPoints, struct{ X, Y float64 }{
			x - w/2 + float64(i)*w/4, y + h/2,
		})
	}
	for i := 1; i < 6; i++ {
		ea.UnionPoints = append(ea.UnionPoints, struct{ X, Y float64 }{
			x + w/2, y - h/2 + float64(i)*h/6,
		})
	}
}

func (ea *EndingAnimation) createFragments(screenWidth, screenHeight int) {
	numFragments := len(ea.CrystalPoints)
	ea.Fragments = make([]CrystalFragmentParticle, numFragments)

	groundY := float64(screenHeight) - 50

	for i := 0; i < numFragments; i++ {
		startX := float64(screenWidth/10) + rand.Float64()*float64(screenWidth*8/10)
		startY := groundY + rand.Float64()*20

		crystalPoint := ea.CrystalPoints[i]
		targetX := ea.CrystalCenterX + crystalPoint.X*ea.CrystalScale
		targetY := ea.CrystalCenterY + crystalPoint.Y*ea.CrystalScale

		var fragColor color.RGBA
		switch crystalPoint.Layer {
		case 0:
			fragColor = color.RGBA{220, 240, 255, 255}
		case 1:
			fragColor = color.RGBA{100, 200, 255, 255}
		case 2:
			fragColor = color.RGBA{80, 150, 255, 255}
		case 3:
			fragColor = color.RGBA{150, 100, 255, 255}
		}

		ea.Fragments[i] = CrystalFragmentParticle{
			X:            startX,
			Y:            startY,
			StartX:       startX,
			StartY:       startY,
			TargetX:      targetX,
			TargetY:      targetY,
			VelocityX:    0,
			VelocityY:    0,
			Size:         2 + rand.Float64()*3,
			Progress:     0,
			DelayTimer:   rand.Float64() * 2.0,
			PulsePhase:   rand.Float64() * math.Pi * 2,
			Alpha:        1.0,
			Color:        fragColor,
			CrystalLayer: crystalPoint.Layer,
		}
	}
}

func (ea *EndingAnimation) createUnionFragments() {
	numUnionFragments := len(ea.UnionPoints)
	ea.UnionFragments = make([]CrystalFragmentParticle, numUnionFragments)

	for i := 0; i < numUnionFragments; i++ {
		angle := rand.Float64() * 2 * math.Pi
		radius := 100 + rand.Float64()*50
		startX := ea.CrystalCenterX + radius*math.Cos(angle)
		startY := ea.CrystalCenterY + radius*math.Sin(angle)

		unionPoint := ea.UnionPoints[i]
		targetX := unionPoint.X
		targetY := unionPoint.Y

		fragColor := color.RGBA{255, 215, 0, 255}

		ea.UnionFragments[i] = CrystalFragmentParticle{
			X:            startX,
			Y:            startY,
			StartX:       startX,
			StartY:       startY,
			TargetX:      targetX,
			TargetY:      targetY,
			VelocityX:    0,
			VelocityY:    0,
			Size:         2 + rand.Float64()*2,
			Progress:     0,
			DelayTimer:   rand.Float64() * 1.0,
			PulsePhase:   rand.Float64() * math.Pi * 2,
			Alpha:        1.0,
			Color:        fragColor,
			CrystalLayer: 0,
		}
	}
}

func (ea *EndingAnimation) Start() {
	ea.IsActive = true
	ea.State = EndingStateFragmentsRising
	ea.Timer = 0
}

func (ea *EndingAnimation) Update(deltaTime float64) {
	if !ea.IsActive {
		return
	}

	ea.Timer += deltaTime

	switch ea.State {
	case EndingStateFragmentsRising:
		ea.updateFragmentsRising(deltaTime)
		if ea.Timer > 5.0 {
			ea.State = EndingStateFormingCrystal
			ea.Timer = 0
		}
	case EndingStateFormingCrystal:
		ea.updateFormingCrystal()
		if ea.Timer > 4.0 {
			ea.State = EndingStateCrystalComplete
			ea.Timer = 0
		}
	case EndingStateCrystalComplete:
		ea.updateCrystalComplete()
		if ea.Timer > 3.0 {
			ea.State = EndingStateFormingUnion
			ea.Timer = 0
			ea.createUnionFragments()
		}
	case EndingStateFormingUnion:
		ea.updateFormingUnion(deltaTime)
		if ea.Timer > 4.0 {
			ea.State = EndingStateUnionComplete
			ea.Timer = 0
		}
	case EndingStateUnionComplete:
		ea.updateUnionComplete()
		ea.State = EndingStateUnionComplete
	case EndingStateComplete:
		ea.State = EndingStateUnionComplete
	}

	for i := range ea.Fragments {
		ea.Fragments[i].PulsePhase += deltaTime * 3.0
	}
}

func (ea *EndingAnimation) updateFragmentsRising(deltaTime float64) {
	for i := range ea.Fragments {
		fragment := &ea.Fragments[i]

		if fragment.DelayTimer > 0 {
			fragment.DelayTimer -= deltaTime
			continue
		}

		riseFactor := ea.Timer / 5.0
		if riseFactor > 1.0 {
			riseFactor = 1.0
		}

		easedRise := 1.0 - math.Pow(1.0-riseFactor, 3.0)
		lateralOffset := math.Sin(ea.Timer*2.0+float64(i)*0.5) * 20.0
		fragment.Y = fragment.StartY - easedRise*200.0
		fragment.X = fragment.StartX + lateralOffset*easedRise
		ea.ScreenShakeX = math.Sin(ea.Timer*10.0) * 2.0 * riseFactor
		ea.ScreenShakeY = math.Sin(ea.Timer*15.0) * 1.5 * riseFactor
	}
}

func (ea *EndingAnimation) updateFormingCrystal() {
	for i := range ea.Fragments {
		fragment := &ea.Fragments[i]
		progress := ea.Timer / 4.0
		if progress > 1.0 {
			progress = 1.0
		}
		layerDelay := float64(fragment.CrystalLayer) * 0.2
		adjustedProgress := math.Max(0, progress-layerDelay)
		if adjustedProgress > 0 {
			adjustedProgress = adjustedProgress * (1.0 + layerDelay)
			if adjustedProgress > 1.0 {
				adjustedProgress = 1.0
			}
		}
		easedProgress := adjustedProgress * adjustedProgress * (3.0 - 2.0*adjustedProgress)
		startY := fragment.StartY - 200.0
		fragment.X = fragment.StartX + (fragment.TargetX-fragment.StartX)*easedProgress
		fragment.Y = startY + (fragment.TargetY-startY)*easedProgress
		fragment.Progress = easedProgress
		baseSize := 2.0
		if fragment.CrystalLayer == 0 {
			baseSize = 4.0
		} else if fragment.CrystalLayer == 1 {
			baseSize = 3.0
		}
		fragment.Size = baseSize + easedProgress*3
	}
	ea.ScreenShakeX = math.Sin(ea.Timer*20.0) * 1.0
	ea.ScreenShakeY = math.Sin(ea.Timer*25.0) * 0.8
}

func (ea *EndingAnimation) updateCrystalComplete() {
	corePulseScale := 1.0 + math.Sin(ea.Timer*4.0)*0.15
	outerPulseScale := 1.0 + math.Sin(ea.Timer*2.0)*0.08
	for i := range ea.Fragments {
		fragment := &ea.Fragments[i]
		crystalPoint := ea.CrystalPoints[i]
		var pulseScale float64
		switch fragment.CrystalLayer {
		case 0:
			pulseScale = corePulseScale
		case 1:
			pulseScale = 1.0 + math.Sin(ea.Timer*3.0)*0.12
		case 2:
			pulseScale = outerPulseScale
		case 3:
			pulseScale = 1.0 + math.Sin(ea.Timer*6.0)*0.2
		}
		fragment.X = ea.CrystalCenterX + crystalPoint.X*ea.CrystalScale*pulseScale
		fragment.Y = ea.CrystalCenterY + crystalPoint.Y*ea.CrystalScale*pulseScale
		baseSize := 3.0
		if fragment.CrystalLayer == 0 {
			baseSize = 6.0
		} else if fragment.CrystalLayer == 1 {
			baseSize = 4.0
		}
		fragment.Size = baseSize + math.Sin(fragment.PulsePhase)*2
		switch fragment.CrystalLayer {
		case 0:
			hue := math.Mod(ea.Timer*1.5+float64(i)*0.1, 1.0)
			fragment.Color = ea.crystalHsvToRGB(0.6+hue*0.1, 0.3, 1.0)
		case 1:
			hue := math.Mod(ea.Timer*1.0+float64(i)*0.1, 1.0)
			fragment.Color = ea.crystalHsvToRGB(0.5+hue*0.1, 0.7, 0.9)
		case 2:
			hue := math.Mod(ea.Timer*0.8+float64(i)*0.1, 1.0)
			fragment.Color = ea.crystalHsvToRGB(0.6+hue*0.05, 0.8, 0.8)
		case 3:
			hue := math.Mod(ea.Timer*2.0+float64(i)*0.15, 1.0)
			fragment.Color = ea.crystalHsvToRGB(0.7+hue*0.1, 0.9, 0.7)
		}
	}
}

func (ea *EndingAnimation) updateFormingUnion(deltaTime float64) {
	ea.updateCrystalComplete()
	for i := range ea.UnionFragments {
		fragment := &ea.UnionFragments[i]
		if fragment.DelayTimer > 0 {
			fragment.DelayTimer -= deltaTime
			continue
		}
		progress := ea.Timer / 4.0
		if progress > 1.0 {
			progress = 1.0
		}
		easedProgress := progress * progress * (3.0 - 2.0*progress)
		fragment.X = fragment.StartX + (fragment.TargetX-fragment.StartX)*easedProgress
		fragment.Y = fragment.StartY + (fragment.TargetY-fragment.StartY)*easedProgress
		fragment.Progress = easedProgress
		fragment.Size = 2.0 + easedProgress*2
		fragment.Alpha = 0.5 + easedProgress*0.5
	}
	ea.ScreenShakeX = math.Sin(ea.Timer*8.0) * 0.5 * (1.0 - ea.Timer/4.0)
	ea.ScreenShakeY = math.Sin(ea.Timer*10.0) * 0.3 * (1.0 - ea.Timer/4.0)
}

func (ea *EndingAnimation) updateUnionComplete() {
	corePulseScale := 1.0 + math.Sin(ea.Timer*2.0)*0.1
	outerPulseScale := 1.0 + math.Sin(ea.Timer*1.5)*0.05

	for i := range ea.Fragments {
		fragment := &ea.Fragments[i]
		crystalPoint := ea.CrystalPoints[i]
		var pulseScale float64
		switch fragment.CrystalLayer {
		case 0:
			pulseScale = corePulseScale
		case 1:
			pulseScale = 1.0 + math.Sin(ea.Timer*1.8)*0.08
		case 2:
			pulseScale = outerPulseScale
		case 3:
			pulseScale = 1.0 + math.Sin(ea.Timer*3.0)*0.12
		}
		fragment.X = ea.CrystalCenterX + crystalPoint.X*ea.CrystalScale*pulseScale
		fragment.Y = ea.CrystalCenterY + crystalPoint.Y*ea.CrystalScale*pulseScale
		baseSize := 3.0
		if fragment.CrystalLayer == 0 {
			baseSize = 5.0
		} else if fragment.CrystalLayer == 1 {
			baseSize = 4.0
		}
		fragment.Size = baseSize + math.Sin(fragment.PulsePhase)*1
	}
	for i := range ea.UnionFragments {
		fragment := &ea.UnionFragments[i]
		fragment.X = fragment.TargetX
		fragment.Y = fragment.TargetY
		fragment.Size = 3.0 + math.Sin(ea.Timer*2.0+float64(i)*0.1)*0.8
		fragment.Alpha = 0.9 + math.Sin(ea.Timer*3.0+float64(i)*0.2)*0.1
		hue := math.Mod(ea.Timer*0.5+float64(i)*0.05, 1.0)
		fragment.Color = ea.goldenHsvToRGB(0.13+hue*0.02, 1.0, 1.0)
	}
	ea.FadeAlpha = 0
}

func (ea *EndingAnimation) goldenHsvToRGB(h, s, v float64) color.RGBA {
	c := v * s
	x := c * (1.0 - math.Abs(math.Mod(h*6.0, 2.0)-1.0))
	m := v - c

	var r, g, b float64
	if h < 1.0/6.0 {
		r, g, b = c, x, 0
	} else if h < 2.0/6.0 {
		r, g, b = x, c, 0
	} else if h < 3.0/6.0 {
		r, g, b = 0, c, x
	} else if h < 4.0/6.0 {
		r, g, b = 0, x, c
	} else if h < 5.0/6.0 {
		r, g, b = x, 0, c
	} else {
		r, g, b = c, 0, x
	}

	return color.RGBA{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
		A: 255,
	}
}

func (ea *EndingAnimation) Draw(screen *ebiten.Image) {
	if !ea.IsActive {
		return
	}

	shakeOffsetX := int(ea.ScreenShakeX)
	shakeOffsetY := int(ea.ScreenShakeY)

	for _, fragment := range ea.Fragments {
		if fragment.Alpha <= 0 {
			continue
		}
		x := fragment.X + float64(shakeOffsetX)
		y := fragment.Y + float64(shakeOffsetY)
		pulseSize := fragment.Size * (1.0 + math.Sin(fragment.PulsePhase)*0.3)
		glowColor := fragment.Color
		glowColor.A = uint8(float64(glowColor.A) * fragment.Alpha * 0.3)
		vector.DrawFilledCircle(screen, float32(x), float32(y), float32(pulseSize*2), glowColor, false)
		coreColor := fragment.Color
		coreColor.A = uint8(float64(coreColor.A) * fragment.Alpha)
		vector.DrawFilledCircle(screen, float32(x), float32(y), float32(pulseSize), coreColor, false)
	}

	for _, fragment := range ea.UnionFragments {
		if fragment.Alpha <= 0 {
			continue
		}
		x := fragment.X + float64(shakeOffsetX)
		y := fragment.Y + float64(shakeOffsetY)
		pulseSize := fragment.Size * (1.0 + math.Sin(fragment.PulsePhase)*0.3)
		glowColor := fragment.Color
		glowColor.A = uint8(float64(glowColor.A) * fragment.Alpha * 0.4)
		vector.DrawFilledCircle(screen, float32(x), float32(y), float32(pulseSize*2.5), glowColor, false)
		coreColor := fragment.Color
		coreColor.A = uint8(float64(coreColor.A) * fragment.Alpha)
		vector.DrawFilledCircle(screen, float32(x), float32(y), float32(pulseSize), coreColor, false)
	}
}

func (ea *EndingAnimation) IsComplete() bool {
	return ea.State == EndingStateUnionComplete
}

func (ea *EndingAnimation) ShouldCloseGame() bool {
	return false
}

func (ea *EndingAnimation) GetMusicVolume() float64 {
	return ea.MusicFadeVolume
}

func (ea *EndingAnimation) crystalHsvToRGB(h, s, v float64) color.RGBA {
	c := v * s
	x := c * (1.0 - math.Abs(math.Mod(h*6.0, 2.0)-1.0))
	m := v - c

	var r, g, b float64
	if h < 1.0/6.0 {
		r, g, b = c, x, 0
	} else if h < 2.0/6.0 {
		r, g, b = x, c, 0
	} else if h < 3.0/6.0 {
		r, g, b = 0, c, x
	} else if h < 4.0/6.0 {
		r, g, b = 0, x, c
	} else if h < 5.0/6.0 {
		r, g, b = x, 0, c
	} else {
		r, g, b = c, 0, x
	}

	return color.RGBA{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
		A: 255,
	}
}
