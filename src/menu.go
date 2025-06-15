package src

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/temidaradev/ebijam25/assets"
	"github.com/temidaradev/esset/v2"
)

type MenuState int

const (
	MenuStateMain MenuState = iota
	MenuStateSettings
)

type MenuItem struct {
	Text     string
	Action   func() MenuState
	Selected bool
}

type Menu struct {
	state              MenuState
	selectedIndex      int
	menuItems          []MenuItem
	settingsItems      []MenuItem
	animationTime      float64
	transitionAlpha    float64
	backgroundAlpha    float64
	startGameRequested bool
	exitRequested      bool
	isFullscreen       bool
}

func NewMenu() *Menu {
	m := &Menu{
		state:           MenuStateMain,
		selectedIndex:   0,
		animationTime:   0,
		transitionAlpha: 1.0,
		backgroundAlpha: 0.8,
		isFullscreen:    ebiten.IsFullscreen(),
	}

	m.menuItems = []MenuItem{
		{Text: "START GAME", Action: func() MenuState {
			m.startGameRequested = true
			return MenuStateMain
		}},
		{Text: "SETTINGS", Action: func() MenuState { return MenuStateSettings }},
		{Text: "EXIT", Action: func() MenuState {
			m.exitRequested = true
			return MenuStateMain
		}},
	}

	fullscreenText := "FULLSCREEN: OFF"
	if m.isFullscreen {
		fullscreenText = "FULLSCREEN: ON"
	}

	m.settingsItems = []MenuItem{
		{Text: "MUSIC VOLUME: 100%", Action: func() MenuState { return MenuStateSettings }},
		{Text: "SOUND EFFECTS: 100%", Action: func() MenuState { return MenuStateSettings }},
		{Text: fullscreenText, Action: func() MenuState {
			m.toggleFullscreen()
			return MenuStateSettings
		}},
		{Text: "BACK", Action: func() MenuState { return MenuStateMain }},
	}
	return m
}

func (m *Menu) toggleFullscreen() {
	m.isFullscreen = !m.isFullscreen
	ebiten.SetFullscreen(m.isFullscreen)

	if m.isFullscreen {
		m.settingsItems[2].Text = "FULLSCREEN: ON"
	} else {
		m.settingsItems[2].Text = "FULLSCREEN: OFF"
	}
}

func (m *Menu) Update() error {
	m.animationTime += 1.0 / 60.0

	currentItems := m.getCurrentMenuItems()

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		m.selectedIndex--
		if m.selectedIndex < 0 {
			m.selectedIndex = len(currentItems) - 1
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		m.selectedIndex++
		if m.selectedIndex >= len(currentItems) {
			m.selectedIndex = 0
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if m.selectedIndex < len(currentItems) {
			newState := currentItems[m.selectedIndex].Action()
			if newState != m.state {
				m.state = newState
				m.selectedIndex = 0
			}
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if m.state != MenuStateMain {
			m.state = MenuStateMain
			m.selectedIndex = 0
		}
	}

	return nil
}

func (m *Menu) Draw(screen *ebiten.Image) {
	screenWidth, screenHeight := screen.Bounds().Dx(), screen.Bounds().Dy()

	vector.DrawFilledRect(screen, 0, 0, float32(screenWidth), float32(screenHeight),
		color.RGBA{0, 0, 0, uint8(m.backgroundAlpha * 255)}, false)

	switch m.state {
	case MenuStateMain:
		m.drawMainMenu(screen, screenWidth, screenHeight)
	case MenuStateSettings:
		m.drawSettingsMenu(screen, screenWidth, screenHeight)
	}
}

func (m *Menu) drawMainMenu(screen *ebiten.Image, screenWidth, screenHeight int) {
	titleText := "Between Layers"
	titleX := float64(screenWidth) * 0.025
	titleY := float64(screenHeight) * 0.2

	titleAlpha := 0.8 + 0.2*math.Sin(m.animationTime*2)
	titleColor := color.RGBA{255, 255, 255, uint8(titleAlpha * 255)}

	esset.DrawText(screen, titleText, titleX, titleY, assets.FontFaceM, titleColor)

	m.drawMenuItems(screen, m.menuItems, screenWidth, screenHeight)
}

func (m *Menu) drawSettingsMenu(screen *ebiten.Image, screenWidth, screenHeight int) {
	titleText := "SETTINGS"
	titleX := float64(screenWidth) * 0.025
	titleY := float64(screenHeight) * 0.2

	esset.DrawText(screen, titleText, titleX, titleY, assets.FontFaceM, color.RGBA{255, 255, 255, 255})

	m.drawMenuItems(screen, m.settingsItems, screenWidth, screenHeight)
}

func (m *Menu) drawMenuItems(screen *ebiten.Image, items []MenuItem, screenWidth, screenHeight int) {
	menuX := float64(screenWidth) * 0.025
	startY := float64(screenHeight) * 0.36
	itemSpacing := float64(screenHeight) * 0.07

	if itemSpacing < 38 {
		itemSpacing = 38
	}
	if itemSpacing > 64 {
		itemSpacing = 64
	}

	for i, item := range items {
		y := startY + float64(i)*itemSpacing
		x := menuX

		var itemColor color.RGBA
		if i == m.selectedIndex {
			pulse := 0.7 + 0.3*math.Sin(m.animationTime*4)
			itemColor = color.RGBA{255, uint8(255 * pulse), 0, 255}

			arrowX := float32(x - 25)
			arrowY := float32(y)
			arrowColor := color.RGBA{255, 255, 100, uint8(255 * pulse)}

			vector.DrawFilledRect(screen, arrowX, arrowY+8, 16, 2, arrowColor, false)
			vector.DrawFilledRect(screen, arrowX+13, arrowY+5, 2, 8, arrowColor, false)
			vector.DrawFilledRect(screen, arrowX+15, arrowY+7, 2, 4, arrowColor, false)
			vector.DrawFilledRect(screen, arrowX+12, arrowY+3, 3, 3, arrowColor, false)
			vector.DrawFilledRect(screen, arrowX+12, arrowY+6, 3, 3, arrowColor, false)
			vector.DrawFilledRect(screen, arrowX+12, arrowY+9, 3, 3, arrowColor, false)
			vector.DrawFilledRect(screen, arrowX+15, arrowY+6, 3, 3, arrowColor, false)

		} else {
			itemColor = color.RGBA{200, 200, 200, 255}
		}

		esset.DrawText(screen, item.Text, x, y, assets.FontFaceS, itemColor)
	}

	hintText := "USE ARROW KEYS OR WASD TO NAVIGATE • ENTER/SPACE TO SELECT • ESC TO GO BACK"
	hintX := float64(screenWidth) * 0.025
	hintY := float64(screenHeight) * 0.95

	esset.DrawText(screen, hintText, hintX, hintY, assets.FontFaceS, color.RGBA{150, 150, 150, 200})
}

func (m *Menu) getCurrentMenuItems() []MenuItem {
	switch m.state {
	case MenuStateSettings:
		return m.settingsItems
	default:
		return m.menuItems
	}
}

func (m *Menu) GetState() MenuState {
	return m.state
}

func (m *Menu) IsStartSelected() bool {
	if m.startGameRequested {
		m.startGameRequested = false
		return true
	}
	return false
}

func (m *Menu) IsExitSelected() bool {
	if m.exitRequested {
		m.exitRequested = false
		return true
	}
	return false
}
