package display

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// all shortcuts have associated commands in the menu bar, so they just call
// their counterparts from menuActions.go.
var (
	shortOneInst   = desktop.CustomShortcut{KeyName: fyne.KeyTab, Modifier: fyne.KeyModifierControl}
	shortUntilHalt = desktop.CustomShortcut{KeyName: fyne.KeyH, Modifier: fyne.KeyModifierControl}
	shortReset     = desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: fyne.KeyModifierControl}
	shortStop      = desktop.CustomShortcut{KeyName: fyne.KeyP, Modifier: fyne.KeyModifierControl}
)

// handleShortcuts runs whem every shortcut is triggered, responsible for
// calling the associated menu action.
func handleShortcuts(sh fyne.Shortcut) {
	desktopSh, ok := sh.(*desktop.CustomShortcut)
	if !ok {
		fmt.Println("shortcut is not of expected type")
		return
	}

	switch *desktopSh {
	case shortOneInst:
		runOneInst()
	case shortUntilHalt:
		runUntilHalt()
	case shortReset:
		restartCode()
	case shortStop:
		stopSim()
	default:
		fmt.Println("invalid shortcut")
	}
}

// setupShortcuts adds all shortcuts from the simulator to a window.
func setupShortcuts() {
	window.Canvas().AddShortcut(&shortOneInst, func(sh fyne.Shortcut) { handleShortcuts(sh) })
	window.Canvas().AddShortcut(&shortUntilHalt, func(sh fyne.Shortcut) { handleShortcuts(sh) })
	window.Canvas().AddShortcut(&shortReset, func(sh fyne.Shortcut) { handleShortcuts(sh) })
	window.Canvas().AddShortcut(&shortStop, func(sh fyne.Shortcut) { handleShortcuts(sh) })
}
