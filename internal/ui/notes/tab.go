package notes

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"github.com/burhanarif4211/rafta/internal/models"
	"github.com/burhanarif4211/rafta/internal/repository"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type NotesTab struct {
	noteFolderRepo repository.NoteFolderRepository
	noteRepo       repository.NoteRepository

	// UI components
	folderTree    *widget.Tree
	noteList      *widget.List
	editor        *widget.Entry
	preview       *widget.RichText
	editorStack   *fyne.Container ////this is deprecated
	folderToolbar *fyne.Container
	mainContent   fyne.CanvasObject // renamed from Content

	// Data
	folders         map[string]*models.NoteFolder
	folderChildren  map[string][]string // parentID -> childIDs; "" for roots
	notes           []*models.Note
	selectedNoteID  string
	currentFolderID string

	win fyne.Window
	mu  sync.RWMutex
}

func NewNotesTab(
	noteFolderRepo repository.NoteFolderRepository,
	noteRepo repository.NoteRepository,
	win fyne.Window,
) *NotesTab {
	nt := &NotesTab{
		noteFolderRepo: noteFolderRepo,
		noteRepo:       noteRepo,
		win:            win,
		folders:        make(map[string]*models.NoteFolder),
		folderChildren: make(map[string][]string),
	}
	nt.buildUI()
	nt.refreshFolders()
	return nt
}

func (nt *NotesTab) buildUI() {
	// ----- Folder tree with toolbar -----
	nt.folderTree = widget.NewTree(
		// child IDs function: given a node ID, return its children
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			nt.mu.RLock()
			defer nt.mu.RUnlock()
			return nt.folderChildren[id]
		},
		// is branch function: true if node has children
		func(id widget.TreeNodeID) bool {
			nt.mu.RLock()
			defer nt.mu.RUnlock()
			return len(nt.folderChildren[id]) > 0
		},
		// create node function
		func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Folder")
		},
		// update node function
		func(id widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
			nt.mu.RLock()
			defer nt.mu.RUnlock()
			if f, ok := nt.folders[id]; ok {
				obj.(*widget.Label).SetText(f.Name)
			} else {
				obj.(*widget.Label).SetText("?")
			}
		},
	)
	nt.folderTree.OnSelected = func(id widget.TreeNodeID) {
		nt.currentFolderID = id
		nt.refreshNotes()
	}

	newFolderBtn := widget.NewButtonWithIcon("New Folder", theme.ContentAddIcon(), nt.createFolder)
	renameFolderBtn := widget.NewButtonWithIcon("Rename", theme.DocumentCreateIcon(), nt.renameFolder)
	deleteFolderBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), nt.deleteFolder)
	nt.folderToolbar = container.NewHBox(newFolderBtn, renameFolderBtn, deleteFolderBtn)

	leftPanel := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Folders", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			nt.folderToolbar,
		),
		nil, nil, nil,
		container.NewScroll(nt.folderTree),
	)

	// ----- Note list -----
	nt.noteList = widget.NewList(
		func() int {
			nt.mu.RLock()
			defer nt.mu.RUnlock()
			return len(nt.notes)
		},
		func() fyne.CanvasObject { return widget.NewLabel("Note") },
		func(i int, obj fyne.CanvasObject) {
			nt.mu.RLock()
			defer nt.mu.RUnlock()
			if i < len(nt.notes) {
				obj.(*widget.Label).SetText(nt.notes[i].Title)
			}
		},
	)
	nt.noteList.OnSelected = func(id widget.ListItemID) {
		nt.mu.RLock()
		if id >= len(nt.notes) {
			nt.mu.RUnlock()
			return
		}
		noteID := nt.notes[id].ID
		nt.mu.RUnlock()
		nt.loadNoteContent(noteID)
	}

	centerPanel := container.NewBorder(
		widget.NewLabelWithStyle("Notes", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		container.NewScroll(nt.noteList),
	)

	// ----- Editor and preview -----
	nt.editor = widget.NewMultiLineEntry()
	nt.editor.Wrapping = fyne.TextWrapWord
	nt.preview = widget.NewRichText()
	nt.preview.Wrapping = fyne.TextWrapWord

	nt.editorStack = container.NewStack(nt.editor, nt.preview)
	nt.preview.Hide()

	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), nt.saveCurrentNote)
	previewBtn := widget.NewButton("Preview", nt.togglePreview)
	newNoteBtn := widget.NewButtonWithIcon("New Note", theme.ContentAddIcon(), nt.createNewNote)

	editorToolbar := container.NewHBox(saveBtn, previewBtn, newNoteBtn)
	editorPanel := container.NewBorder(editorToolbar, nil, nil, nil, nt.editorStack)

	// ----- Assemble main layout -----
	leftCenter := container.NewHSplit(leftPanel, centerPanel)
	leftCenter.SetOffset(0.3)
	mainSplit := container.NewHSplit(leftCenter, editorPanel)
	mainSplit.SetOffset(0.4)

	nt.mainContent = mainSplit
}

func (nt *NotesTab) Content() fyne.CanvasObject {
	return nt.mainContent
}

func (nt *NotesTab) Reload() {
	nt.refreshFolders()
}

// ---------- Folder operations ----------
func (nt *NotesTab) refreshFolders() {
	go func() {
		allFolders, err := nt.noteFolderRepo.GetAll()
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("failed to load folders: %v", err), nt.win)
			})
			return
		}

		newFolders := make(map[string]*models.NoteFolder)
		newChildren := make(map[string][]string)
		for _, f := range allFolders {
			newFolders[f.ID] = f
			parent := ""
			if f.ParentID.Valid {
				parent = f.ParentID.String
			}
			newChildren[parent] = append(newChildren[parent], f.ID)
		}

		fyne.Do(func() {
			nt.mu.Lock()
			nt.folders = newFolders
			nt.folderChildren = newChildren
			nt.mu.Unlock()
			nt.folderTree.Refresh()
		})
	}()
}

func (nt *NotesTab) createFolder() {
	entry := widget.NewEntry()
	dialog.ShowForm("New Folder", "Create", "Cancel",
		[]*widget.FormItem{widget.NewFormItem("Name", entry)},
		func(confirmed bool) {
			if !confirmed || entry.Text == "" {
				return
			}
			parentID := sql.NullString{Valid: false}
			if nt.currentFolderID != "" {
				parentID = sql.NullString{String: nt.currentFolderID, Valid: true}
			}
			folder := models.NewNoteFolder(entry.Text, parentID)
			go func() {
				if err := nt.noteFolderRepo.Create(folder); err != nil {
					fyne.Do(func() {
						dialog.ShowError(fmt.Errorf("failed to create folder: %v", err), nt.win)
					})
					return
				}
				nt.refreshFolders()
			}()
		}, nt.win)
}

func (nt *NotesTab) renameFolder() {
	if nt.currentFolderID == "" {
		dialog.ShowInformation("No folder selected", "Select a folder to rename.", nt.win)
		return
	}
	nt.mu.RLock()
	folder := nt.folders[nt.currentFolderID]
	nt.mu.RUnlock()
	if folder == nil {
		return
	}
	entry := widget.NewEntry()
	entry.SetText(folder.Name)
	dialog.ShowForm("Rename Folder", "Rename", "Cancel",
		[]*widget.FormItem{widget.NewFormItem("Name", entry)},
		func(confirmed bool) {
			if !confirmed || entry.Text == "" || entry.Text == folder.Name {
				return
			}
			folder.Name = entry.Text
			go func() {
				if err := nt.noteFolderRepo.Update(folder); err != nil {
					fyne.Do(func() {
						dialog.ShowError(fmt.Errorf("failed to rename folder: %v", err), nt.win)
					})
					return
				}
				nt.refreshFolders()
			}()
		}, nt.win)
}

func (nt *NotesTab) deleteFolder() {
	if nt.currentFolderID == "" {
		dialog.ShowInformation("No folder selected", "Select a folder to delete.", nt.win)
		return
	}
	nt.mu.RLock()
	folder := nt.folders[nt.currentFolderID]
	nt.mu.RUnlock()
	if folder == nil {
		return
	}
	dialog.ShowConfirm("Delete Folder",
		fmt.Sprintf("Delete '%s' and all its contents?", folder.Name),
		func(ok bool) {
			if !ok {
				return
			}
			go func() {
				if err := nt.noteFolderRepo.Delete(nt.currentFolderID); err != nil {
					fyne.Do(func() {
						dialog.ShowError(fmt.Errorf("failed to delete folder: %v", err), nt.win)
					})
					return
				}
				nt.refreshFolders()
				// Clear current folder selection
				nt.currentFolderID = ""
				fyne.Do(func() {
					nt.notes = nil
					nt.noteList.Refresh()
					nt.editor.SetText("")
				})
			}()
		}, nt.win)
}

// ---------- Note operations ----------
func (nt *NotesTab) refreshNotes() {
	if nt.currentFolderID == "" {
		return
	}
	go func() {
		notes, err := nt.noteRepo.GetByFolder(nt.currentFolderID)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("failed to load notes: %v", err), nt.win)
			})
			return
		}
		fyne.Do(func() {
			nt.mu.Lock()
			nt.notes = notes
			nt.selectedNoteID = ""
			nt.mu.Unlock()
			nt.noteList.Refresh()
			nt.editor.SetText("")
		})
	}()
}

func (nt *NotesTab) loadNoteContent(noteID string) {
	go func() {
		note, err := nt.noteRepo.GetByID(noteID)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("failed to load note: %v", err), nt.win)
			})
			return
		}
		fyne.Do(func() {
			nt.selectedNoteID = noteID
			nt.editor.SetText(note.Content)
			if nt.preview.Visible() {
				nt.preview.ParseMarkdown(note.Content)
			}
		})
	}()
}

func (nt *NotesTab) saveCurrentNote() {
	if nt.selectedNoteID == "" {
		dialog.ShowInformation("No note selected", "Select a note to save.", nt.win)
		return
	}
	content := nt.editor.Text
	go func() {
		note, err := nt.noteRepo.GetByID(nt.selectedNoteID)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("failed to get note: %v", err), nt.win)
			})
			return
		}
		note.Content = content
		note.Title = extractTitle(content)
		if note.Title == "" {
			note.Title = "Untitled"
		}
		if err := nt.noteRepo.Update(note); err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("failed to save note: %v", err), nt.win)
			})
			return
		}
		fyne.Do(func() {
			nt.mu.Lock()
			for _, n := range nt.notes {
				if n.ID == note.ID {
					n.Title = note.Title
					break
				}
			}
			nt.mu.Unlock()
			nt.noteList.Refresh()
			dialog.ShowInformation("Saved", "Note saved.", nt.win)
		})
	}()
}

func (nt *NotesTab) createNewNote() {
	if nt.currentFolderID == "" {
		dialog.ShowInformation("No folder selected", "Select a folder first.", nt.win)
		return
	}
	go func() {
		newNote := models.NewNote("Untitled", "", nt.currentFolderID)
		if err := nt.noteRepo.Create(newNote); err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("failed to create note: %v", err), nt.win)
			})
			return
		}
		fyne.Do(func() {
			nt.refreshNotes()
		})
	}()
}

func (nt *NotesTab) togglePreview() {
	if nt.preview.Visible() {
		nt.preview.Hide()
		nt.editor.Show()
	} else {
		nt.preview.ParseMarkdown(nt.editor.Text)
		nt.editor.Hide()
		nt.preview.Show()
	}
}

// extractTitle returns the first non-empty line of markdown.
func extractTitle(md string) string {
	lines := strings.Split(strings.TrimSpace(md), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// Remove leading '#' for markdown headings
			return strings.TrimLeft(line, "# ")
		}
	}
	return ""
}

// type NotesTab struct {
// 	// Repositories
// 	noteFolderRepo repository.NoteFolderRepository
// 	noteRepo       repository.NoteRepository

// 	// UI components
// 	folderTree     *widget.Tree
// 	noteList       *widget.List
// 	editor         *widget.Entry
// 	preview        *widget.RichText
// 	previewVisible bool
// 	split          *container.Split

// 	// Data
// 	folders        map[string]*models.NoteFolder // id -> folder
// 	folderChildren map[string][]string           // parentID -> childIDs
// 	rootIDs        []string

// 	notes          []*models.Note
// 	selectedNoteID string

// 	// State
// 	currentFolderID string
// 	win             fyne.Window
// 	mu              sync.Mutex // protect notes slice
// }

// func NewNotesTab(
// 	noteFolderRepo repository.NoteFolderRepository,
// 	noteRepo repository.NoteRepository,
// 	win fyne.Window,
// ) *NotesTab {
// 	nt := &NotesTab{
// 		noteFolderRepo: noteFolderRepo,
// 		noteRepo:       noteRepo,
// 		win:            win,
// 		folders:        make(map[string]*models.NoteFolder),
// 		folderChildren: make(map[string][]string),
// 	}

// 	// Build UI
// 	nt.buildUI()
// 	// Load folders asynchronously
// 	nt.refreshFolders()
// 	return nt
// }

// func (nt *NotesTab) refreshFolders() {
// 	go func() {
// 		_, err := nt.noteFolderRepo.GetRoots()
// 		if err != nil {
// 			fyne.Do(func() {
// 				dialog.ShowError(fmt.Errorf("failed to load folders: %v", err), nt.win)
// 			})
// 			return
// 		}
// 		// Build maps
// 		newFolders := make(map[string]*models.NoteFolder)
// 		newChildren := make(map[string][]string)
// 		var rootIDs []string

// 		// First collect all folders by traversing from roots? But we need all folders to build children.
// 		// Simpler: fetch all folders (or we can fetch recursively, but GetAll is easier)
// 		allFolders, err := nt.noteFolderRepo.GetAll()
// 		if err != nil {
// 			fyne.Do(func() {
// 				dialog.ShowError(fmt.Errorf("failed to load all folders: %v", err), nt.win)
// 			})
// 			return
// 		}
// 		for _, f := range allFolders {
// 			newFolders[f.ID] = f
// 			parentID := ""
// 			if f.ParentID.Valid {
// 				parentID = f.ParentID.String
// 			}
// 			newChildren[parentID] = append(newChildren[parentID], f.ID)
// 		}
// 		// rootIDs are those with parentID = "" (i.e., nil parent)
// 		rootIDs = newChildren[""]

// 		fyne.Do(func() {
// 			nt.mu.Lock()
// 			nt.folders = newFolders
// 			nt.folderChildren = newChildren
// 			nt.rootIDs = rootIDs
// 			nt.mu.Unlock()
// 			nt.folderTree.Refresh()
// 		})
// 	}()
// }

// func (nt *NotesTab) refreshNotes() {
// 	if nt.currentFolderID == "" {
// 		return
// 	}
// 	go func() {
// 		notes, err := nt.noteRepo.GetByFolder(nt.currentFolderID)
// 		if err != nil {
// 			fyne.Do(func() {
// 				dialog.ShowError(fmt.Errorf("failed to load notes: %v", err), nt.win)
// 			})
// 			return
// 		}
// 		fyne.Do(func() {
// 			nt.mu.Lock()
// 			nt.notes = notes
// 			nt.selectedNoteID = ""
// 			nt.mu.Unlock()
// 			nt.noteList.Refresh()
// 			nt.editor.SetText("") // clear editor
// 		})
// 	}()
// }

// func (nt *NotesTab) loadNoteContent(noteID string) {
// 	go func() {
// 		note, err := nt.noteRepo.GetByID(noteID)
// 		if err != nil {
// 			fyne.Do(func() {
// 				dialog.ShowError(fmt.Errorf("failed to load note: %v", err), nt.win)
// 			})
// 			return
// 		}
// 		fyne.Do(func() {
// 			nt.editor.SetText(note.Content)
// 		})
// 	}()
// }

// func (nt *NotesTab) saveCurrentNote() {
// 	if nt.selectedNoteID == "" {
// 		dialog.ShowInformation("No note selected", "Please select a note to save.", nt.win)
// 		return
// 	}
// 	content := nt.editor.Text
// 	go func() {
// 		note, err := nt.noteRepo.GetByID(nt.selectedNoteID)
// 		if err != nil {
// 			fyne.Do(func() {
// 				dialog.ShowError(fmt.Errorf("failed to get note: %v", err), nt.win)
// 			})
// 			return
// 		}
// 		note.Content = content
// 		note.Title = extractTitle(content) // simple: first line or "Untitled"
// 		if note.Title == "" {
// 			note.Title = "Untitled"
// 		}
// 		if err := nt.noteRepo.Update(note); err != nil {
// 			fyne.Do(func() {
// 				dialog.ShowError(fmt.Errorf("failed to save note: %v", err), nt.win)
// 			})
// 			return
// 		}
// 		// Update title in list
// 		fyne.Do(func() {
// 			// find note in notes slice and update title
// 			nt.mu.Lock()
// 			for _, n := range nt.notes {
// 				if n.ID == note.ID {
// 					n.Title = note.Title
// 					break
// 				}
// 			}
// 			nt.mu.Unlock()
// 			nt.noteList.Refresh()
// 			dialog.ShowInformation("Saved", "Note saved successfully.", nt.win)
// 		})
// 	}()
// }

// // extractTitle returns first non-empty line or empty string.
// func extractTitle(md string) string {
// 	lines := strings.Split(strings.TrimSpace(md), "\n")
// 	for _, line := range lines {
// 		line = strings.TrimSpace(line)
// 		if line != "" {
// 			// Remove markdown heading symbols? For now just return line.
// 			return line
// 		}
// 	}
// 	return ""
// }

// func (nt *NotesTab) createNewNote() {
// 	if nt.currentFolderID == "" {
// 		dialog.ShowInformation("No folder selected", "Please select a folder first.", nt.win)
// 		return
// 	}
// 	go func() {
// 		newNote := models.NewNote("Untitled", "", nt.currentFolderID)
// 		if err := nt.noteRepo.Create(newNote); err != nil {
// 			fyne.Do(func() {
// 				dialog.ShowError(fmt.Errorf("failed to create note: %v", err), nt.win)
// 			})
// 			return
// 		}
// 		// Refresh notes list and select the new note
// 		fyne.Do(func() {
// 			nt.refreshNotes() // will load notes and clear editor
// 			// after refresh, we need to select the new note? we could find it and load.
// 			// For simplicity, just refresh and let user click.
// 		})
// 	}()
// }

// func (nt *NotesTab) togglePreview() {
// 	if nt.previewVisible {
// 		nt.preview.Hide()
// 		nt.split.Hide()
// 		nt.editor.Show()
// 	} else {
// 		// Render markdown
// 		content := nt.editor.Text
// 		_ = goldmark.Convert([]byte(content), &bytes.Buffer{}) // need import
// 		// Actually we need to convert to RichText segments. Use a simple renderer.
// 		// For now, just show raw markdown in preview.
// 		nt.preview.ParseMarkdown(content) // Fyne's RichText has ParseMarkdown?
// 		// Fyne's RichText does have ParseMarkdown (since v2.3)
// 		nt.preview.Show()
// 		nt.editor.Hide()
// 		nt.split.Show()
// 	}
// 	nt.previewVisible = !nt.previewVisible
// }

// func (nt *NotesTab) buildUI() {
// 	// Folder tree
// 	nt.folderTree = widget.NewTree(
// 		// Return IDs of root nodes
// 		func() (roots []string) {
// 			return nt.rootIDs
// 		},
// 		// Return true if node is branch (has children)
// 		func(id string) bool {
// 			return len(nt.folderChildren[id]) > 0
// 		},
// 		// Return children of given node
// 		func(id string) (children []string) {
// 			return nt.folderChildren[id]
// 		},
// 		// Create node UI
// 		func(branch bool) fyne.CanvasObject {
// 			return widget.NewLabel("Folder")
// 		},
// 		// Update node UI
// 		func(id string, branch bool, obj fyne.CanvasObject) {
// 			label := obj.(*widget.Label)
// 			if folder, ok := nt.folders[id]; ok {
// 				label.SetText(folder.Name)
// 			} else {
// 				label.SetText("?")
// 			}
// 		},
// 	)
// 	nt.folderTree.OnSelected = func(id string) {
// 		nt.currentFolderID = id
// 		nt.refreshNotes() // load notes for selected folder
// 	}

// 	// Note list
// 	nt.noteList = widget.NewList(
// 		func() int {
// 			nt.mu.Lock()
// 			defer nt.mu.Unlock()
// 			return len(nt.notes)
// 		},
// 		func() fyne.CanvasObject {
// 			return widget.NewLabel("Note")
// 		},
// 		func(i int, obj fyne.CanvasObject) {
// 			nt.mu.Lock()
// 			defer nt.mu.Unlock()
// 			if i >= len(nt.notes) {
// 				return
// 			}
// 			note := nt.notes[i]
// 			obj.(*widget.Label).SetText(note.Title)
// 		},
// 	)
// 	nt.noteList.OnSelected = func(id widget.ListItemID) {
// 		nt.mu.Lock()
// 		if id >= len(nt.notes) {
// 			nt.mu.Unlock()
// 			return
// 		}
// 		note := nt.notes[id]
// 		nt.selectedNoteID = note.ID
// 		nt.mu.Unlock()
// 		nt.loadNoteContent(note.ID)
// 	}

// 	// Editor and preview
// 	nt.editor = widget.NewMultiLineEntry()
// 	nt.editor.Wrapping = fyne.TextWrapWord
// 	nt.preview = widget.NewRichText()
// 	nt.preview.Wrapping = fyne.TextWrapWord

// 	// Buttons for editor
// 	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
// 		nt.saveCurrentNote()
// 	})
// 	previewBtn := widget.NewButton("Preview", func() {
// 		nt.togglePreview()
// 	})
// 	newNoteBtn := widget.NewButtonWithIcon("New Note", theme.ContentAddIcon(), func() {
// 		nt.createNewNote()
// 	})

// 	editorToolbar := container.NewHBox(saveBtn, previewBtn, newNoteBtn)

// 	// Editor area: either just editor, or split between editor and preview
// 	nt.split = container.NewHSplit(nt.editor, nt.preview)
// 	nt.split.SetOffset(0.5)
// 	nt.previewVisible = false
// 	nt.preview.Hide()

// 	editorContainer := container.NewBorder(editorToolbar, nil, nil, nil, nt.split)

// 	// Main layout: left (folder tree), center (note list), right (editor)
// 	leftPanel := container.NewBorder(
// 		widget.NewLabelWithStyle("Folders", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
// 		nil, nil, nil,
// 		container.NewScroll(nt.folderTree),
// 	)
// 	centerPanel := container.NewBorder(
// 		widget.NewLabelWithStyle("Notes", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
// 		nil, nil, nil,
// 		container.NewScroll(nt.noteList),
// 	)

// 	// Use a horizontal split: left (folders) + center (note list) + right (editor)
// 	// Actually we can nest splits: left | (center | right)
// 	leftCenter := container.NewHSplit(leftPanel, centerPanel)
// 	leftCenter.SetOffset(0.3)
// 	mainSplit := container.NewHSplit(leftCenter, editorContainer)
// 	mainSplit.SetOffset(0.4) // 40% for left+center, 60% for editor

// 	nt.Content = mainSplit
// }

// // Content returns the main UI object for the tab.
// func (nt *NotesTab) Content() fyne.CanvasObject {
// 	return nt.Content
// }
