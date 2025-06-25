package src

import (
	"image/color"
	"math"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/temidaradev/ebijam25/assets"
	"github.com/temidaradev/esset/v2"
)

type MenuState int

const (
	MenuStateMain MenuState = iota
	MenuStatePause
	MenuStateRespawn
)

type MenuItem struct {
	Text     string
	Action   func() MenuState
	Selected bool
}

type Menu struct {
	state                     MenuState
	previousState             MenuState
	selectedIndex             int
	menuItems                 []MenuItem
	pauseItems                []MenuItem
	respawnItems              []MenuItem
	animationTime             float64
	transitionAlpha           float64
	backgroundAlpha           float64
	startGameRequested        bool
	exitRequested             bool
	continueRequested         bool
	restartRequested          bool
	fullscreenToggleRequested bool
	controller                *ControllerInput
}

func NewMenu() *Menu {
	m := &Menu{
		state:           MenuStateMain,
		previousState:   MenuStateMain,
		selectedIndex:   0,
		animationTime:   0,
		transitionAlpha: 1.0,
		backgroundAlpha: 0.8,
		controller:      NewControllerInput(),
	}

	m.menuItems = []MenuItem{
		{Text: "START GAME", Action: func() MenuState {
			m.startGameRequested = true
			return MenuStateMain
		}},
		{Text: "EXIT", Action: func() MenuState {
			m.exitRequested = true
			return MenuStateMain
		}},
	}

	m.pauseItems = []MenuItem{
		{Text: "CONTINUE", Action: func() MenuState {
			m.continueRequested = true
			return MenuStatePause
		}},
		{Text: "EXIT GAME", Action: func() MenuState {
			os.Exit(0)
			return MenuStatePause
		}},
	}

	m.respawnItems = []MenuItem{
		{Text: "QUIT GAME", Action: func() MenuState {
			os.Exit(1)
			return MenuStateRespawn
		}},
	}

	return m
}

func (m *Menu) Update() error {
	m.animationTime += 1.0 / 60.0

	m.controller.Update()

	currentItems := m.getCurrentMenuItems()

	upPressed := inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW)
	downPressed := inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS)
	selectPressed := inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace)

	upPressed = upPressed || m.controller.IsUpJustPressed()
	downPressed = downPressed || m.controller.IsDownJustPressed()
	selectPressed = selectPressed || m.controller.IsSelectJustPressed()

	if upPressed {
		m.selectedIndex--
		if m.selectedIndex < 0 {
			m.selectedIndex = len(currentItems) - 1
		}
	}

	if downPressed {
		m.selectedIndex++
		if m.selectedIndex >= len(currentItems) {
			m.selectedIndex = 0
		}
	}

	if selectPressed {
		if m.selectedIndex < len(currentItems) {
			newState := currentItems[m.selectedIndex].Action()
			if newState != m.state {
				m.state = newState
				m.selectedIndex = 0
			}
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
	case MenuStatePause:
		m.drawPauseMenu(screen, screenWidth, screenHeight)
	case MenuStateRespawn:
		m.drawRespawnMenu(screen, screenWidth, screenHeight)
	}
}

func (m *Menu) drawMainMenu(screen *ebiten.Image, screenWidth, screenHeight int) {
	titleText := "FIGHT FOR UNION"
	titleX := float64(screenWidth) * 0.025
	titleY := float64(screenHeight) * 0.15

	glitchPhase := m.animationTime * 3.0
	titleAlpha := 0.8 + 0.2*math.Sin(glitchPhase)

	if rand.Float64() < 0.1 {
		titleColor := color.RGBA{
			uint8(200 + rand.Intn(56)),
			uint8(50 + rand.Intn(100)),
			uint8(200 + rand.Intn(56)),
			uint8(titleAlpha * 255),
		}
		esset.DrawText(screen, titleText, titleX, titleY, assets.FontFaceM, titleColor)
	} else {
		titleColor := color.RGBA{255, 100, 255, uint8(titleAlpha * 255)}
		esset.DrawText(screen, titleText, titleX, titleY, assets.FontFaceM, titleColor)
	}

	subtitleText := "Find the CURSED ARTIFACTS to break your mind"
	subtitleX := float64(screenWidth) * 0.025
	subtitleY := float64(screenHeight) * 0.25
	esset.DrawText(screen, subtitleText, subtitleX, subtitleY, assets.FontFaceS, color.RGBA{200, 150, 200, 255})

	disclaimerText := "WARNING: Contains disturbing visual effects and reality distortion"
	disclaimerX := float64(screenWidth) * 0.025
	disclaimerY := float64(screenHeight) * 0.32
	esset.DrawText(screen, disclaimerText, disclaimerX, disclaimerY, assets.FontFaceS, color.RGBA{255, 200, 100, 255})

	m.drawMenuItems(screen, m.menuItems, screenWidth, screenHeight)
}

func (m *Menu) drawPauseMenu(screen *ebiten.Image, screenWidth, screenHeight int) {
	titleText := "PAUSED"
	titleX := float64(screenWidth) * 0.025
	titleY := float64(screenHeight) * 0.25

	esset.DrawText(screen, titleText, titleX, titleY, assets.FontFaceM, color.RGBA{255, 255, 255, 255})

	m.drawMenuItems(screen, m.pauseItems, screenWidth, screenHeight)
}

func (m *Menu) drawRespawnMenu(screen *ebiten.Image, screenWidth, screenHeight int) {
	vector.DrawFilledRect(screen, 0, 0, float32(screenWidth), float32(screenHeight),
		color.RGBA{0, 0, 0, 150}, false)

	titleText := "YOU DIED"
	titleX := float64(screenWidth)*0.5 - 100
	titleY := float64(screenHeight) * 0.25

	pulse := 0.7 + 0.3*math.Sin(m.animationTime*3)
	titleColor := color.RGBA{255, uint8(100 * pulse), uint8(100 * pulse), 255}

	esset.DrawText(screen, titleText, titleX, titleY, assets.FontFaceM, titleColor)

	subtitleText := "Choose your next action:"
	subtitleX := float64(screenWidth) * 0.025
	subtitleY := titleY + 60

	esset.DrawText(screen, subtitleText, subtitleX, subtitleY, assets.FontFaceS, color.RGBA{200, 200, 200, 255})

	m.drawMenuItems(screen, m.respawnItems, screenWidth, screenHeight)
}

func (m *Menu) drawMenuItems(screen *ebiten.Image, items []MenuItem, screenWidth, screenHeight int) {
	menuX := float64(screenWidth) * 0.025
	startY := float64(screenHeight) * 0.4
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

}

func (m *Menu) getCurrentMenuItems() []MenuItem {
	switch m.state {
	case MenuStatePause:
		return m.pauseItems
	case MenuStateRespawn:
		return m.respawnItems
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

func (m *Menu) IsContinueRequested() bool {
	if m.continueRequested {
		m.continueRequested = false
		return true
	}
	return false
}

func (m *Menu) IsFullscreenToggleRequested() bool {
	if m.fullscreenToggleRequested {
		m.fullscreenToggleRequested = false
		return true
	}
	return false
}

func (m *Menu) IsRestartRequested() bool {
	if m.restartRequested {
		m.restartRequested = false
		return true
	}
	return false
}

func (m *Menu) SetPauseState() {
	m.state = MenuStatePause
	m.selectedIndex = 0
}

func (m *Menu) SetRespawnState() {
	m.state = MenuStateRespawn
	m.selectedIndex = 0
}
