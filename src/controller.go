package src

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type ControllerConfig struct {
	DeadZoneLeft     float64
	StickSensitivity float64
}

type ControllerInput struct {
	gamepadID         ebiten.GamepadID
	isActive          bool
	config            ControllerConfig
	hasStandardLayout bool
}

func NewControllerInput() *ControllerInput {
	return &ControllerInput{
		config: ControllerConfig{
			DeadZoneLeft:     0.15,
			StickSensitivity: 1.0,
		},
	}
}

func (c *ControllerInput) Update() {
	gamepadIDs := inpututil.AppendJustConnectedGamepadIDs(nil)
	if len(gamepadIDs) == 0 {
		gamepadIDs = ebiten.AppendGamepadIDs(nil)
	}

	if len(gamepadIDs) > 0 {
		found := false
		if c.isActive {
			for _, id := range gamepadIDs {
				if id == c.gamepadID {
					found = true
					break
				}
			}
		}

		if !found {
			c.gamepadID = gamepadIDs[0]
		}
		c.isActive = true
		c.hasStandardLayout = ebiten.IsStandardGamepadLayoutAvailable(c.gamepadID)
	} else {
		c.isActive = false
	}
}

func (c *ControllerInput) IsActive() bool {
	return c.isActive
}

func (c *ControllerInput) applyDeadZone(x, y, deadZone float64) (float64, float64) {
	magnitude := math.Sqrt(x*x + y*y)
	if magnitude < deadZone {
		return 0, 0
	}

	normalizedMagnitude := (magnitude - deadZone) / (1.0 - deadZone)
	if normalizedMagnitude > 1.0 {
		normalizedMagnitude = 1.0
	}

	if magnitude > 0 {
		ratio := normalizedMagnitude / magnitude
		return x * ratio, y * ratio
	}
	return 0, 0
}

func (c *ControllerInput) GetLeftStick() (float64, float64) {
	if !c.isActive {
		return 0, 0
	}

	var x, y float64
	if c.hasStandardLayout {
		x = ebiten.StandardGamepadAxisValue(c.gamepadID, ebiten.StandardGamepadAxisLeftStickHorizontal)
		y = ebiten.StandardGamepadAxisValue(c.gamepadID, ebiten.StandardGamepadAxisLeftStickVertical)
	} else {
		x = ebiten.GamepadAxisValue(c.gamepadID, 0)
		y = ebiten.GamepadAxisValue(c.gamepadID, 1)
	}

	return c.applyDeadZone(x, y, c.config.DeadZoneLeft)
}

func (c *ControllerInput) GetHorizontalAxis() float64 {
	x, _ := c.GetLeftStick()
	return x * c.config.StickSensitivity
}

func (c *ControllerInput) IsLeftPressed() bool {
	if !c.isActive {
		return false
	}

	if c.hasStandardLayout {
		if ebiten.IsStandardGamepadButtonPressed(c.gamepadID, ebiten.StandardGamepadButtonLeftLeft) {
			return true
		}
	} else {
		if ebiten.IsGamepadButtonPressed(c.gamepadID, 14) {
			return true
		}
	}

	x, _ := c.GetLeftStick()
	return x < -c.config.DeadZoneLeft
}

func (c *ControllerInput) IsRightPressed() bool {
	if !c.isActive {
		return false
	}

	if c.hasStandardLayout {
		if ebiten.IsStandardGamepadButtonPressed(c.gamepadID, ebiten.StandardGamepadButtonLeftRight) {
			return true
		}
	} else {
		if ebiten.IsGamepadButtonPressed(c.gamepadID, 15) {
			return true
		}
	}

	x, _ := c.GetLeftStick()
	return x > c.config.DeadZoneLeft
}

func (c *ControllerInput) IsUpJustPressed() bool {
	if !c.isActive {
		return false
	}

	if c.hasStandardLayout {
		return inpututil.IsStandardGamepadButtonJustPressed(c.gamepadID, ebiten.StandardGamepadButtonLeftTop)
	}
	return inpututil.IsGamepadButtonJustPressed(c.gamepadID, 12)
}

func (c *ControllerInput) IsDownJustPressed() bool {
	if !c.isActive {
		return false
	}

	if c.hasStandardLayout {
		return inpututil.IsStandardGamepadButtonJustPressed(c.gamepadID, ebiten.StandardGamepadButtonLeftBottom)
	}
	return inpututil.IsGamepadButtonJustPressed(c.gamepadID, 13)
}

func (c *ControllerInput) IsJumpJustPressed() bool {
	if !c.isActive {
		return false
	}

	if c.hasStandardLayout {
		return inpututil.IsStandardGamepadButtonJustPressed(c.gamepadID, ebiten.StandardGamepadButtonRightRight)
	}

	return inpututil.IsGamepadButtonJustPressed(c.gamepadID, 1)
}

func (c *ControllerInput) IsSelectJustPressed() bool {
	if !c.isActive {
		return false
	}

	if c.hasStandardLayout {
		return inpututil.IsStandardGamepadButtonJustPressed(c.gamepadID, ebiten.StandardGamepadButtonRightBottom)
	}

	return inpututil.IsGamepadButtonJustPressed(c.gamepadID, 0)
}

func (c *ControllerInput) IsBackJustPressed() bool {
	if !c.isActive {
		return false
	}

	if c.hasStandardLayout {
		return inpututil.IsStandardGamepadButtonJustPressed(c.gamepadID, ebiten.StandardGamepadButtonRightRight) ||
			inpututil.IsStandardGamepadButtonJustPressed(c.gamepadID, ebiten.StandardGamepadButtonCenterLeft) ||
			inpututil.IsStandardGamepadButtonJustPressed(c.gamepadID, ebiten.StandardGamepadButtonCenterRight)
	}

	return inpututil.IsGamepadButtonJustPressed(c.gamepadID, 1) ||
		inpututil.IsGamepadButtonJustPressed(c.gamepadID, 6) ||
		inpututil.IsGamepadButtonJustPressed(c.gamepadID, 7)
}

func (c *ControllerInput) IsPauseJustPressed() bool {
	if !c.isActive {
		return false
	}

	if c.hasStandardLayout {
		return inpututil.IsStandardGamepadButtonJustPressed(c.gamepadID, ebiten.StandardGamepadButtonCenterRight) ||
			inpututil.IsStandardGamepadButtonJustPressed(c.gamepadID, ebiten.StandardGamepadButtonCenterLeft)
	}

	return inpututil.IsGamepadButtonJustPressed(c.gamepadID, 7) ||
		inpututil.IsGamepadButtonJustPressed(c.gamepadID, 6)
}

func (c *ControllerInput) IsRollJustPressed() bool {
	if !c.isActive {
		return false
	}
	if c.hasStandardLayout {
		return inpututil.IsStandardGamepadButtonJustPressed(c.gamepadID, ebiten.StandardGamepadButtonRightBottom)
	}
	return inpututil.IsGamepadButtonJustPressed(c.gamepadID, 0)
}
