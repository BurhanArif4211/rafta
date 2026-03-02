package main

import (
	"database/sql"
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/burhanarif4211/rafta/internal/db"
	"github.com/burhanarif4211/rafta/internal/repository"
	"github.com/burhanarif4211/rafta/internal/sync"
	"github.com/burhanarif4211/rafta/internal/ui/notes"
	"github.com/burhanarif4211/rafta/internal/ui/todos"
)

func main() {
	// Initialize database
	database, err := db.InitDB("./data/rafta-main.db")
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
	fileMenu := fyne.NewMenu("File", syncItem)
	mainMenu := fyne.NewMainMenu(fileMenu)
	w.SetMainMenu(mainMenu)

	//main ui setter
	w.SetContent(tabs)

	// w.Resize(fyne.NewSize(800, 600))
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
