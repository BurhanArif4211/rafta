package notes

import (
	"database/sql"
	"sort"
	"sync"
	"time"

	"github.com/burhanarif4211/rafta/internal/models"
	"github.com/burhanarif4211/rafta/internal/repository"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ItemType int

const (
	TypeFolder ItemType = iota
	TypeNote
)

type Item struct {
	ID        string
	Type      ItemType
	Name      string
	ParentID  string // for notes: folder ID; for folders: parent folder ID (empty if root)
	CreatedAt time.Time
	UpdatedAt time.Time
}
type NotesTab struct {
	noteFolderRepo repository.NoteFolderRepository
	noteRepo       repository.NoteRepository

	// Data
	items   map[string]*Item
	rootIDs []string

	// UI
	tree           *widget.Tree
	editor         *NoteEditor
	win            fyne.Window
	content        fyne.CanvasObject
	leftPanel      *fyne.Container
	splitContainer *container.Split
	editorOnly     *fyne.Container
	mainView       *fyne.Container
	showSidebar    bool

	// Selection
	selectedID string

	// Callbacks
	onAddNote    func(string)
	onRename     func(string, string)
	onDelete     func(string, ItemType)
	onSelectNote func(string)

	mu sync.RWMutex
}

func NewNotesTab(noteFolderRepo repository.NoteFolderRepository, noteRepo repository.NoteRepository, win fyne.Window) *NotesTab {
	nt := &NotesTab{
		noteFolderRepo: noteFolderRepo,
		noteRepo:       noteRepo,
		win:            win,
		items:          make(map[string]*Item),
		showSidebar:    true,
	}
	nt.onAddNote = nt.createNote
	nt.onRename = nt.renameItem
	nt.onDelete = nt.deleteItem
	nt.onSelectNote = nt.loadNote
	nt.buildUI()
	nt.refreshData()
	return nt
}

func (nt *NotesTab) buildUI() {
	// ----- Left panel: folder toolbar + tree -----
	nt.tree = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			// nt.mu.RLock()
			// defer nt.mu.RUnlock()
			if id == "" {
				return nt.rootIDs
			}
			var children []string
			for _, item := range nt.items {
				if item.ParentID == id {
					children = append(children, item.ID)
				}
			}
			nt.sortChildren(children)
			return children
		},
		func(id widget.TreeNodeID) bool {
			if id == "" {
				return true
			}
			item, ok := nt.items[id]
			return ok && item.Type == TypeFolder
		},
		func(branch bool) fyne.CanvasObject {
			return newTreeRow(branch, nt.onAddNote, nt.onRename, nt.onDelete, nt.win)
		},
		func(id widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
			row := obj.(*treeRow)
			if id == "" {
				return
			}
			item, ok := nt.items[id]
			if !ok {
				return
			}
			row.SetItem(id, item.Name, item.Type)
		},
	)
	nt.tree.OnSelected = func(id widget.TreeNodeID) {
		nt.selectedID = id
		item, ok := nt.items[id]
		if ok && item.Type == TypeNote {
			nt.onSelectNote(id)
		}
	}
	nt.tree.OnUnselected = func(id widget.TreeNodeID) {
		nt.selectedID = ""
	}

	// Folder toolbar
	folderToolbar := container.NewHBox(
		widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), nt.refreshData),
		widget.NewButtonWithIcon("", theme.ContentAddIcon(), nt.createFolder),
		widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() {
			if nt.selectedID != "" {
				item := nt.items[nt.selectedID]
				nt.onRename(nt.selectedID, item.Name)
			}
		}),
		widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			if nt.selectedID != "" {
				item := nt.items[nt.selectedID]
				nt.onDelete(nt.selectedID, item.Type)
			}
		}),
	)
	nt.leftPanel = container.NewBorder(folderToolbar, nil, nil, nil, nt.tree)

	// ----- Editor -----
	nt.editor = NewNoteEditor(nt.noteRepo, nt.win)

	// ----- Split container (resizable) -----
	nt.splitContainer = container.NewHSplit(nt.leftPanel, nt.editor.Content())
	nt.splitContainer.SetOffset(0.3)

	// ----- Editor-only view (collapsed) -----
	nt.editorOnly = container.NewStack(nt.editor.Content())

	// ----- Main view: holds either split or editorOnly -----
	nt.mainView = container.NewStack(nt.splitContainer) // start with split

	// Final layout: top toolbar + mainView
	nt.content = container.NewBorder(
		container.NewHBox(widget.NewButtonWithIcon("", theme.NavigateBackIcon(), nt.toggleSidebar)),
		nil, nil, nil,
		nt.mainView)
}

func (nt *NotesTab) toggleSidebar() {
	nt.showSidebar = !nt.showSidebar
	nt.mainView.RemoveAll()
	if nt.showSidebar {
		nt.mainView.Add(nt.splitContainer)
	} else {
		nt.mainView.Add(nt.editorOnly)
	}
	nt.mainView.Refresh()
}

func (nt *NotesTab) Content() fyne.CanvasObject {
	return nt.content
}

func (nt *NotesTab) refreshData() {
	// Fetch all folders
	folders, err := nt.noteFolderRepo.GetAll()
	if err != nil {
		dialog.ShowError(err, nt.win)
		return
	}
	// Fetch all notes
	notes, err := nt.noteRepo.GetAll()
	if err != nil {
		dialog.ShowError(err, nt.win)
		return
	}
	// nt.mu.Lock()
	// defer nt.mu.Unlock()

	// Build items map
	nt.items = make(map[string]*Item)
	for _, f := range folders {
		parentID := ""
		if f.ParentID.Valid {
			parentID = f.ParentID.String
		}
		nt.items[f.ID] = &Item{
			ID:        f.ID,
			Type:      TypeFolder,
			Name:      f.Name,
			ParentID:  parentID,
			CreatedAt: f.CreatedAt,
			UpdatedAt: f.UpdatedAt,
		}
	}
	for _, n := range notes {
		nt.items[n.ID] = &Item{
			ID:        n.ID,
			Type:      TypeNote,
			Name:      n.Title,
			ParentID:  n.FolderID,
			CreatedAt: n.CreatedAt,
			UpdatedAt: n.UpdatedAt,
		}
	}
	// Compute root IDs (items with empty ParentID)
	nt.rootIDs = nil
	for _, item := range nt.items {
		if item.ParentID == "" {
			nt.rootIDs = append(nt.rootIDs, item.ID)
		}
	}
	nt.sortChildren(nt.rootIDs)
	nt.tree.Refresh()
}
func (nt *NotesTab) createFolder() {
	entry := widget.NewEntry()
	dialog.ShowForm("New Folder", "Create", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Name", entry),
	}, func(ok bool) {
		if !ok || entry.Text == "" {
			return
		}
		var parentID sql.NullString
		if nt.selectedID != "" {
			if item, exists := nt.items[nt.selectedID]; exists && item.Type == TypeFolder {
				parentID = sql.NullString{String: nt.selectedID, Valid: true}
			}
		}
		folder := models.NewNoteFolder(entry.Text, parentID)
		err := nt.noteFolderRepo.Create(folder)
		if err != nil {
			dialog.ShowError(err, nt.win)
			return
		}
		nt.refreshData()
	}, nt.win)
}
func (nt *NotesTab) createNote(folderID string) {
	entry := widget.NewEntry()
	dialog.ShowForm("New Note", "Create", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Title", entry),
	}, func(ok bool) {
		if !ok || entry.Text == "" {
			return
		}
		note := models.NewNote(entry.Text, "", folderID)
		err := nt.noteRepo.Create(note)
		if err != nil {
			dialog.ShowError(err, nt.win)
			return
		}
		nt.refreshData()
		// Automatically open the new note
		nt.onSelectNote(note.ID)
	}, nt.win)
}
func (nt *NotesTab) renameItem(id string, currentName string) {
	entry := widget.NewEntry()
	entry.SetText(currentName)
	dialog.ShowForm("Rename", "Save", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Name", entry),
	}, func(ok bool) {
		if !ok || entry.Text == "" || entry.Text == currentName {
			return
		}
		item := nt.items[id]
		var err error
		if item.Type == TypeFolder {
			folder, _ := nt.noteFolderRepo.GetByID(id)
			folder.Name = entry.Text
			err = nt.noteFolderRepo.Update(folder)
		} else {
			note, _ := nt.noteRepo.GetByID(id)
			note.Title = entry.Text
			err = nt.noteRepo.Update(note)
		}
		if err != nil {
			dialog.ShowError(err, nt.win)
			return
		}
		nt.refreshData()
	}, nt.win)
}
func (nt *NotesTab) deleteItem(id string, itemType ItemType) {
	var typeStr string
	if itemType == TypeFolder {
		typeStr = "folder"
	} else {
		typeStr = "note"
	}
	dialog.ShowConfirm("Delete", "Delete this "+typeStr+"?", func(ok bool) {
		if !ok {
			return
		}
		var err error
		if itemType == TypeFolder {
			err = nt.noteFolderRepo.Delete(id)
		} else {
			err = nt.noteRepo.Delete(id)
		}
		if err != nil {
			dialog.ShowError(err, nt.win)
			return
		}
		if itemType == TypeNote && nt.editor.CurrentNoteID() == id {
			nt.editor.Clear()
		}
		nt.refreshData()
	}, nt.win)
}
func (nt *NotesTab) loadNote(noteID string) {
	note, err := nt.noteRepo.GetByID(noteID)
	if err != nil {
		dialog.ShowError(err, nt.win)
		return
	}
	nt.editor.LoadNote(note)
}
func (nt *NotesTab) sortChildren(ids []string) {
	sort.SliceStable(ids, func(i, j int) bool {
		a := nt.items[ids[i]]
		b := nt.items[ids[j]]
		// Folders first
		if a.Type != b.Type {
			return a.Type == TypeFolder
		}
		// Then by creation date (oldest first)
		return a.CreatedAt.Before(b.CreatedAt)
	})
}
