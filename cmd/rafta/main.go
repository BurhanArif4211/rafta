package main

import (
	"database/sql"
	"fmt"
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/burhanarif4211/rafta/internal/db"
	"github.com/burhanarif4211/rafta/internal/repository"
	"github.com/burhanarif4211/rafta/internal/sync"
	"github.com/burhanarif4211/rafta/internal/ui/notes"
	"github.com/burhanarif4211/rafta/internal/ui/todos"
)

// forestTheme implements the forest aesthetic with variant support.
type forestTheme struct {
	fyne.Theme
}

func NewForestTheme() fyne.Theme {
	return &forestTheme{Theme: theme.DefaultTheme()}
}

func (t *forestTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Light variant
	if variant == theme.VariantLight {
		switch name {
		case theme.ColorNameBackground:
			return color.NRGBA{R: 0xf5, G: 0xf7, B: 0xec, A: 0xff}
		case theme.ColorNameForeground:
			return color.NRGBA{R: 0x2c, G: 0x3e, B: 0x2c, A: 0xff}
		case theme.ColorNameButton:
			return color.NRGBA{R: 0xa5, G: 0xb8, B: 0x77, A: 0xff}
		case theme.ColorNamePrimary:
			return color.NRGBA{R: 0x5a, G: 0x7d, B: 0x5a, A: 0xff}
		case theme.ColorNameHover:
			return color.NRGBA{R: 0x6b, G: 0x8e, B: 0x6b, A: 0x99}
		case theme.ColorNamePlaceHolder:
			return color.NRGBA{R: 0x7f, G: 0x8c, B: 0x6b, A: 0xcc}
		case theme.ColorNameDisabled:
			return color.NRGBA{R: 0xaa, G: 0xaa, B: 0xaa, A: 0x99}
		}
	} else { // Dark variant
		switch name {
		case theme.ColorNameBackground:
			return color.NRGBA{R: 0x1e, G: 0x2a, B: 0x1e, A: 0xff}
		case theme.ColorNameForeground:
			return color.NRGBA{R: 0xe0, G: 0xe0, B: 0xc0, A: 0xff}
		case theme.ColorNameButton:
			return color.NRGBA{R: 0x3a, G: 0x4e, B: 0x3a, A: 0xff}
		case theme.ColorNamePrimary:
			return color.NRGBA{R: 0x8b, G: 0x9a, B: 0x5e, A: 0xff}
		case theme.ColorNameHover:
			return color.NRGBA{R: 0x4a, G: 0x5e, B: 0x4a, A: 0xcc}
		case theme.ColorNamePlaceHolder:
			return color.NRGBA{R: 0x8a, G: 0x99, B: 0x7a, A: 0xaa}
		case theme.ColorNameDisabled:
			return color.NRGBA{R: 0x55, G: 0x55, B: 0x55, A: 0x99}
		}
	}
	return t.Theme.Color(name, variant)
}

func (t *forestTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 12
	case theme.SizeNameCaptionText:
		return 10
	case theme.SizeNameHeadingText:
		return 18
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 10
	case theme.SizeNameScrollBar:
		return 8
	case theme.SizeNameInlineIcon:
		return 12
	}
	return t.Theme.Size(name)
}

// forcedVariantTheme wraps a theme and forces a specific variant.
type forcedVariantTheme struct {
	fyne.Theme
	variant fyne.ThemeVariant
}

func (f forcedVariantTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return f.Theme.Color(name, f.variant)
}

func main() {
	// Initialize databases
	dbPath := db.GetDatabasePath()
	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// Create repositories
	noteFolderRepo := repository.NewNoteFolderRepository(database)
	noteRepo := repository.NewNoteRepository(database)
	todoFolderRepo := repository.NewTodoFolderRepository(database)
	todoRepo := repository.NewTodoRepository(database)
	todoStepRepo := repository.NewTodoStepRepository(database)

	// Start sync server
	syncServer := sync.NewServer(noteFolderRepo, noteRepo, todoFolderRepo, todoRepo, todoStepRepo)
	if err := syncServer.Start("4211"); err != nil {
		log.Printf("Failed to start sync server: %v", err)
	}
	defer syncServer.Stop()
	// Create Fyne app
	a := app.New()
	w := a.NewWindow("Rafta")

	baseTheme := NewForestTheme()
	lightTheme := forcedVariantTheme{baseTheme, theme.VariantLight}
	darkTheme := forcedVariantTheme{baseTheme, theme.VariantDark}

	a.Settings().SetTheme(darkTheme)

	// Create notes tab
	notesTab := notes.NewNotesTab(noteFolderRepo, noteRepo, w)

	// Create todos tab (we'll implement later)
	todosTab := todos.NewTodosTab(todoFolderRepo, todoRepo, todoStepRepo, w)

	tabs := container.NewAppTabs(
		container.NewTabItem("Notes", notesTab.Content()),
		container.NewTabItem("Todos", todosTab.Content()),
	)

	// Sync menu item
	syncItem := fyne.NewMenuItem("Sync from device...", func() {
		showSyncDialog(w, database)
	})
	themeLightItem := fyne.NewMenuItem("Switch To Light", func() {
		a.Settings().SetTheme(lightTheme)
	})
	themeDarkItem := fyne.NewMenuItem("Switch To Dark", func() {
		a.Settings().SetTheme(darkTheme)
	})
	fileMenu := fyne.NewMenu("File", syncItem)
	themeMenu := fyne.NewMenu("Theme", themeLightItem, themeDarkItem)
	mainMenu := fyne.NewMainMenu(fileMenu, themeMenu)
	w.SetMainMenu(mainMenu)

	//main ui setter
	w.SetContent(tabs)

	w.Resize(fyne.NewSize(900, 600))
	w.ShowAndRun()
}

func showSyncDialog(parent fyne.Window, db *sql.DB) {
	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("ENTER IP ADDRESS")

	items := []*widget.FormItem{
		widget.NewFormItem("Device IP", ipEntry),
	}
	dialog.ShowForm("Sync from device", "Pull", "Cancel", items, func(confirmed bool) {

		if !confirmed {
			return
		}
		addr := ipEntry.Text
		if addr == "" {
			dialog.ShowError(fmt.Errorf("please enter an address"), parent)
			return
		}

		client := sync.NewClient(db)
		go func() {
			err := client.Pull(addr)
			// Run UI update on main thread
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("sync failed: %v", err), parent)
				})
			} else {
				fyne.Do(func() {
					dialog.ShowInformation("Sync completed", "Data successfully pulled from device.", parent)
				})
			}
		}()
	}, parent)
}
