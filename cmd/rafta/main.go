package main

import (
	"database/sql"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/burhanarif4211/rafta/internal/db"
	"github.com/burhanarif4211/rafta/internal/repository"
	"github.com/burhanarif4211/rafta/internal/sync"
	"log"
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

	// Start sync server on port 8080
	syncServer := sync.NewServer(noteFolderRepo, noteRepo, todoFolderRepo, todoRepo, todoStepRepo)
	if err := syncServer.Start("8080"); err != nil {
		log.Printf("Failed to start sync server: %v", err)
	}
	defer syncServer.Stop()

	// Create Fyne app
	a := app.New()
	w := a.NewWindow("MyApp")

	// Build UI (placeholder – we'll replace with actual tabs later)
	content := widget.NewLabel("Main content will go here")

	// Sync menu item
	syncItem := fyne.NewMenuItem("Sync from device...", func() {
		showSyncDialog(w, database)
	})
	fileMenu := fyne.NewMenu("File", syncItem)
	mainMenu := fyne.NewMainMenu(fileMenu)
	w.SetMainMenu(mainMenu)

	w.SetContent(content)
	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}

func showSyncDialog(parent fyne.Window, db *sql.DB) {
	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("192.168.0.2:8080")

	items := []*widget.FormItem{
		widget.NewFormItem("Device IP:port", ipEntry),
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
					// TODO: refresh UI (folders, notes, todos)
				})
			}
		}()
	}, parent)
}
