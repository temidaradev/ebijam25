package src

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type RuneElement struct {
	ID         string
	X, Y       float64
	Width      float64
	Height     float64
	IsActive   bool
	Collected  bool
	SetGroup   string
	Symbol     rune
	Color      color.RGBA
	PulsePhase float64
}

type SetGate struct {
	X, Y           float64
	Width          float64
	Height         float64
	IsOpen         bool
	IsActive       bool
	RequiredSets   []string
	RequiredUnions [][]string
	RequiredItems  map[string]bool
	CollectedSets  map[string][]string
	FormedUnions   [][]string
	ClosedColor    color.RGBA
	OpenColor      color.RGBA
	ProgressColor  color.RGBA
	OpenProgress   float64
	AnimationPhase float64
}

type ItemCollectable struct {
	ID           string
	X, Y         float64
	Width        float64
	Height       float64
	IsActive     bool
	Collected    bool
	ItemType     string
	SetValue     string
	Color        color.RGBA
	Symbol       string
	BobbingPhase float64
}

type SetLogicPuzzle struct {
	ID              string
	X, Y            float64
	IsActive        bool
	IsSolved        bool
	Runes           []*RuneElement
	Gates           []*SetGate
	Items           []*ItemCollectable
	PlayerInventory map[string][]string
	SolutionHint    string
	ShowHint        bool
	HintTimer       float64
}

type SetOperations struct{}

func NewRuneElement(id, setGroup string, x, y float64, symbol rune, color color.RGBA) *RuneElement {
	return &RuneElement{
		ID:         id,
		X:          x,
		Y:          y,
		Width:      24,
		Height:     24,
		IsActive:   true,
		Collected:  false,
		SetGroup:   setGroup,
		Symbol:     symbol,
		Color:      color,
		PulsePhase: 0,
	}
}

func NewSetGate(x, y, width, height float64) *SetGate {
	return &SetGate{
		X:              x,
		Y:              y,
		Width:          width,
		Height:         height,
		IsOpen:         false,
		IsActive:       true,
		RequiredSets:   []string{},
		RequiredUnions: [][]string{},
		RequiredItems:  make(map[string]bool),
		CollectedSets:  make(map[string][]string),
		FormedUnions:   [][]string{},
		ClosedColor:    color.RGBA{150, 50, 50, 255},
		OpenColor:      color.RGBA{50, 150, 50, 255},
		ProgressColor:  color.RGBA{100, 100, 200, 255},
		OpenProgress:   0.0,
		AnimationPhase: 0,
	}
}

func NewItemCollectable(id, itemType, setValue string, x, y float64, itemColor color.RGBA, symbol string) *ItemCollectable {
	return &ItemCollectable{
		ID:           id,
		X:            x,
		Y:            y,
		Width:        20,
		Height:       20,
		IsActive:     true,
		Collected:    false,
		ItemType:     itemType,
		SetValue:     setValue,
		Color:        itemColor,
		Symbol:       symbol,
		BobbingPhase: 0,
	}
}

func NewSetLogicPuzzle(id string, x, y float64) *SetLogicPuzzle {
	return &SetLogicPuzzle{
		ID:              id,
		X:               x,
		Y:               y,
		IsActive:        true,
		IsSolved:        false,
		Runes:           []*RuneElement{},
		Gates:           []*SetGate{},
		Items:           []*ItemCollectable{},
		PlayerInventory: make(map[string][]string),
		SolutionHint:    "",
		ShowHint:        false,
		HintTimer:       0,
	}
}

func (re *RuneElement) Update(deltaTime float64) {
	if !re.IsActive || re.Collected {
		return
	}

	re.PulsePhase += deltaTime * 3.0
	if re.PulsePhase > 2*math.Pi {
		re.PulsePhase -= 2 * math.Pi
	}
}

func (sg *SetGate) Update(deltaTime float64) {
	if !sg.IsActive {
		return
	}

	sg.AnimationPhase += deltaTime * 2.0
	if sg.AnimationPhase > 2*math.Pi {
		sg.AnimationPhase -= 2 * math.Pi
	}

	shouldOpen := sg.checkSetLogicConditions()

	if shouldOpen && !sg.IsOpen {
		sg.OpenProgress = math.Min(1.0, sg.OpenProgress+deltaTime*2.0)
		if sg.OpenProgress >= 1.0 {
			sg.IsOpen = true
		}
	} else if !shouldOpen && sg.IsOpen {
		sg.OpenProgress = math.Max(0.0, sg.OpenProgress-deltaTime*2.0)
		if sg.OpenProgress <= 0.0 {
			sg.IsOpen = false
		}
	}
}

func (ic *ItemCollectable) Update(deltaTime float64) {
	if !ic.IsActive || ic.Collected {
		return
	}

	ic.BobbingPhase += deltaTime * 2.0
	if ic.BobbingPhase > 2*math.Pi {
		ic.BobbingPhase -= 2 * math.Pi
	}
}

func (slp *SetLogicPuzzle) Update(deltaTime float64) {
	if !slp.IsActive || slp.IsSolved {
		return
	}

	for _, rune := range slp.Runes {
		rune.Update(deltaTime)
	}

	for _, gate := range slp.Gates {
		gate.Update(deltaTime)
	}

	for _, item := range slp.Items {
		item.Update(deltaTime)
	}

	if slp.HintTimer > 0 {
		slp.HintTimer -= deltaTime
		if slp.HintTimer <= 0 {
			slp.ShowHint = false
		}
	}

	slp.checkPuzzleSolved()
}

func (sg *SetGate) checkSetLogicConditions() bool {
	for _, requiredSet := range sg.RequiredSets {
		if _, exists := sg.CollectedSets[requiredSet]; !exists {
			return false
		}
		if len(sg.CollectedSets[requiredSet]) == 0 {
			return false
		}
	}

	for _, requiredUnion := range sg.RequiredUnions {
		if !sg.isUnionFormed(requiredUnion) {
			return false
		}
	}

	for item := range sg.RequiredItems {
		if !sg.RequiredItems[item] {
			return false
		}
	}

	return true
}

func (sg *SetGate) isUnionFormed(union []string) bool {
	unionElements := make(map[string]bool)

	for _, setName := range union {
		if elements, exists := sg.CollectedSets[setName]; exists {
			for _, element := range elements {
				unionElements[element] = true
			}
		}
	}

	for _, setName := range union {
		hasElementFromSet := false
		if elements, exists := sg.CollectedSets[setName]; exists && len(elements) > 0 {
			hasElementFromSet = true
		}
		if !hasElementFromSet {
			return false
		}
	}

	return true
}

func (slp *SetLogicPuzzle) checkPuzzleSolved() {
	allGatesOpen := true
	for _, gate := range slp.Gates {
		if gate.IsActive && !gate.IsOpen {
			allGatesOpen = false
			break
		}
	}

	if allGatesOpen && !slp.IsSolved {
		slp.IsSolved = true
	}
}

func (slp *SetLogicPuzzle) CollectRune(runeID string) bool {
	for _, rune := range slp.Runes {
		if rune.ID == runeID && !rune.Collected {
			rune.Collected = true

			if slp.PlayerInventory[rune.SetGroup] == nil {
				slp.PlayerInventory[rune.SetGroup] = []string{}
			}
			slp.PlayerInventory[rune.SetGroup] = append(slp.PlayerInventory[rune.SetGroup], rune.ID)

			slp.updateGatesWithCollection(rune.SetGroup, rune.ID)

			return true
		}
	}
	return false
}

func (slp *SetLogicPuzzle) CollectItem(itemID string) bool {
	for _, item := range slp.Items {
		if item.ID == itemID && !item.Collected {
			item.Collected = true

			for _, gate := range slp.Gates {
				if _, required := gate.RequiredItems[itemID]; required {
					gate.RequiredItems[itemID] = true
				}
			}

			return true
		}
	}
	return false
}

func (slp *SetLogicPuzzle) updateGatesWithCollection(setGroup, elementID string) {
	for _, gate := range slp.Gates {
		if gate.CollectedSets[setGroup] == nil {
			gate.CollectedSets[setGroup] = []string{}
		}

		found := false
		for _, existing := range gate.CollectedSets[setGroup] {
			if existing == elementID {
				found = true
				break
			}
		}
		if !found {
			gate.CollectedSets[setGroup] = append(gate.CollectedSets[setGroup], elementID)
		}
	}
}

func (slp *SetLogicPuzzle) ShowHintFor(duration float64) {
	slp.ShowHint = true
	slp.HintTimer = duration
}

func (so *SetOperations) Union(setA, setB []string) []string {
	unionMap := make(map[string]bool)
	result := []string{}

	for _, element := range setA {
		if !unionMap[element] {
			unionMap[element] = true
			result = append(result, element)
		}
	}

	for _, element := range setB {
		if !unionMap[element] {
			unionMap[element] = true
			result = append(result, element)
		}
	}

	return result
}

func (so *SetOperations) Intersection(setA, setB []string) []string {
	elementCount := make(map[string]int)
	result := []string{}

	for _, element := range setA {
		elementCount[element]++
	}

	for _, element := range setB {
		elementCount[element]++
		if elementCount[element] == 2 {
			result = append(result, element)
		}
	}

	return result
}

func (so *SetOperations) Difference(setA, setB []string) []string {
	setBElements := make(map[string]bool)
	result := []string{}

	for _, element := range setB {
		setBElements[element] = true
	}

	for _, element := range setA {
		if !setBElements[element] {
			result = append(result, element)
		}
	}

	return result
}

func (re *RuneElement) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if re.Collected {
		return
	}

	screenX := re.X - cameraX
	screenY := re.Y - cameraY

	pulseScale := 1.0 + math.Sin(re.PulsePhase)*0.1
	size := re.Width * pulseScale

	vector.DrawFilledCircle(screen, float32(screenX), float32(screenY), float32(size/2), re.Color, false)

	glowColor := color.RGBA{255, 255, 255, uint8(100 + 50*math.Sin(re.PulsePhase))}
	vector.DrawFilledCircle(screen, float32(screenX), float32(screenY), float32(size/3), glowColor, false)
}

func (sg *SetGate) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	screenX := sg.X - cameraX
	screenY := sg.Y - cameraY

	gateColor := sg.ClosedColor
	if sg.IsOpen {
		gateColor = sg.OpenColor
	} else if sg.OpenProgress > 0 {
		progress := sg.OpenProgress
		gateColor = color.RGBA{
			uint8(float64(sg.ClosedColor.R)*(1-progress) + float64(sg.ProgressColor.R)*progress),
			uint8(float64(sg.ClosedColor.G)*(1-progress) + float64(sg.ProgressColor.G)*progress),
			uint8(float64(sg.ClosedColor.B)*(1-progress) + float64(sg.ProgressColor.B)*progress),
			255,
		}
	}

	gateHeight := sg.Height * (1.0 - sg.OpenProgress)
	if gateHeight > 0 {
		vector.DrawFilledRect(screen, float32(screenX-sg.Width/2), float32(screenY-gateHeight/2), float32(sg.Width), float32(gateHeight), gateColor, false)
	}

	if sg.OpenProgress > 0 && sg.OpenProgress < 1.0 {
		energyIntensity := math.Sin(sg.AnimationPhase) * sg.OpenProgress
		energyColor := color.RGBA{100, 200, 255, uint8(50 * energyIntensity)}

		for i := 0; i < 8; i++ {
			angle := float64(i)*math.Pi*2.0/8.0 + sg.AnimationPhase
			particleX := screenX + math.Cos(angle)*sg.Width*0.6
			particleY := screenY + math.Sin(angle)*sg.Height*0.6

			vector.DrawFilledCircle(screen, float32(particleX), float32(particleY), 3*float32(energyIntensity), energyColor, false)
		}
	}
}

func (ic *ItemCollectable) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if ic.Collected {
		return
	}

	screenX := ic.X - cameraX
	screenY := ic.Y - cameraY + math.Sin(ic.BobbingPhase)*3

	vector.DrawFilledRect(screen, float32(screenX-ic.Width/2), float32(screenY-ic.Height/2), float32(ic.Width), float32(ic.Height), ic.Color, false)

	sparkleColor := color.RGBA{255, 255, 255, uint8(100 + 50*math.Sin(ic.BobbingPhase*2))}
	vector.DrawFilledCircle(screen, float32(screenX), float32(screenY), 2, sparkleColor, false)
}
