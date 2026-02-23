package main

import (
	"database/sql"
	"github.com/burhanarif4211/rafta/internal/db"
	"github.com/burhanarif4211/rafta/internal/models"
	"github.com/burhanarif4211/rafta/internal/repository"
	"log"
)

func main() {
	// Initialize database
	database, err := db.InitDB("./myapp.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// Example: create a note folder
	folderRepo := repository.NewNoteFolderRepository(database)
	rootFolder := models.NewNoteFolder("Nun Notes", sql.NullString{Valid: false})
	if err := folderRepo.Create(rootFolder); err != nil {
		log.Fatal(err)
	}
	log.Printf("Created folder with ID: %s", rootFolder.ID)

	// ... later we will launch the UI
}
