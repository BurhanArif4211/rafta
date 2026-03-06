package todos

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

type Item struct {
	ID        string
	Type      ItemType
	Name      string
	ParentID  string // for todos: folder ID; for folders: parent folder ID (empty if root)
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TodosTab struct {
	folderRepo repository.TodoFolderRepository
	todoRepo   repository.TodoRepository
	stepRepo   repository.TodoStepRepository

	// Data
	items   map[string]*Item
	rootIDs []string

	// UI
	tree           *widget.Tree
	editor         *TodoEditor
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
	onAdd    func(string)
	onRename func(string, string)
	onDelete func(string, ItemType)
	onSelect func(string)
	mu       sync.RWMutex
}

func NewTodosTab(folderRepo repository.TodoFolderRepository, todoRepo repository.TodoRepository, stepRepo repository.TodoStepRepository, win fyne.Window) *TodosTab {
	tt := &TodosTab{
		folderRepo:  folderRepo,
		todoRepo:    todoRepo,
		stepRepo:    stepRepo,
		win:         win,
		items:       make(map[string]*Item),
		showSidebar: true,
	}
	tt.onAdd = tt.handleAdd
	tt.onRename = tt.renameItem
	tt.onDelete = tt.deleteItem
	tt.onSelect = tt.selectTodo
	tt.buildUI()
	tt.refreshData()
	return tt
}

func (tt *TodosTab) buildUI() {
	// ----- Left panel: folder toolbar + tree -----
	tt.tree = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			if id == "" {
				return tt.rootIDs
			}
			var children []string
			for _, item := range tt.items {
				if item.ParentID == id {
					children = append(children, item.ID)
				}
			}
			tt.sortChildren(children)
			return children
		},
		func(id widget.TreeNodeID) bool {
			if id == "" {
				return true
			}
			item, ok := tt.items[id]
			return ok && item.Type == TypeFolder
		},
		func(branch bool) fyne.CanvasObject {
			return newTreeRow(branch, tt.onAdd, tt.onRename, tt.onDelete, tt.win)
		},
		func(id widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
			row := obj.(*treeRow)
			if id == "" {
				return
			}
			item, ok := tt.items[id]
			if !ok {
				return
			}
			row.SetItem(id, item.Name, item.Type)
		},
	)
	tt.tree.OnSelected = func(id widget.TreeNodeID) {
		tt.selectedID = id
		item, ok := tt.items[id]
		if ok && item.Type == TypeTodo {
			tt.onSelect(id)
		}
	}
	tt.tree.OnUnselected = func(id widget.TreeNodeID) {
		tt.selectedID = ""
		tt.editor.Clear()
	}

	// Folder toolbar
	folderToolbar := container.NewHBox(
		widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), tt.refreshData),
		widget.NewButtonWithIcon("", theme.ContentAddIcon(), tt.createFolder),
		widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() {
			if tt.selectedID != "" {
				item := tt.items[tt.selectedID]
				tt.onRename(tt.selectedID, item.Name)
			}
		}),
		widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			if tt.selectedID != "" {
				item := tt.items[tt.selectedID]
				tt.onDelete(tt.selectedID, item.Type)
			}
		}),
	)
	tt.leftPanel = container.NewBorder(folderToolbar, nil, nil, nil, tt.tree)

	// ----- Editor -----
	tt.editor = NewTodoEditor(tt.todoRepo, tt.stepRepo, tt.win)

	// ----- Split container -----
	tt.splitContainer = container.NewHSplit(tt.leftPanel, tt.editor.Content())
	tt.splitContainer.SetOffset(0.3)

	// ----- Editor-only view -----
	tt.editorOnly = container.NewStack(tt.editor.Content())

	// ----- Main view -----
	tt.mainView = container.NewStack(tt.splitContainer)

	tt.content = container.NewBorder(
		container.NewHBox(widget.NewButtonWithIcon("", theme.NavigateBackIcon(), tt.toggleSidebar)),
		nil, nil, nil,
		tt.mainView)
}

func (tt *TodosTab) toggleSidebar() {
	tt.showSidebar = !tt.showSidebar
	tt.mainView.RemoveAll()
	if tt.showSidebar {
		tt.mainView.Add(tt.splitContainer)
	} else {
		tt.mainView.Add(tt.editorOnly)
	}
	tt.mainView.Refresh()
}

func (tt *TodosTab) Content() fyne.CanvasObject {
	return tt.content
}

func (tt *TodosTab) refreshData() {
	// Fetch all folders
	folders, err := tt.folderRepo.GetAll()
	if err != nil {
		dialog.ShowError(err, tt.win)
		return
	}
	// Fetch all todos
	todos, err := tt.todoRepo.GetAll()
	if err != nil {
		dialog.ShowError(err, tt.win)
		return
	}
	// Build items map
	tt.items = make(map[string]*Item)
	for _, f := range folders {
		parentID := ""
		if f.ParentID.Valid {
			parentID = f.ParentID.String
		}
		tt.items[f.ID] = &Item{
			ID:        f.ID,
			Type:      TypeFolder,
			Name:      f.Name,
			ParentID:  parentID,
			CreatedAt: f.CreatedAt,
			UpdatedAt: f.UpdatedAt,
		}
	}
	for _, t := range todos {
		tt.items[t.ID] = &Item{
			ID:        t.ID,
			Type:      TypeTodo,
			Name:      t.Title,
			ParentID:  t.FolderID,
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.UpdatedAt,
		}
	}
	// Compute root IDs
	tt.rootIDs = nil
	for _, item := range tt.items {
		if item.ParentID == "" {
			tt.rootIDs = append(tt.rootIDs, item.ID)
		}
	}
	tt.sortChildren(tt.rootIDs)

	tt.tree.Refresh()
}

// Folder creation (always a folder)
func (tt *TodosTab) createFolder() {
	entry := widget.NewEntry()
	dialog.ShowForm("New Folder", "Create", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Name", entry),
	}, func(ok bool) {
		if !ok || entry.Text == "" {
			return
		}
		var parentID sql.NullString
		if tt.selectedID != "" {
			if item, exists := tt.items[tt.selectedID]; exists && item.Type == TypeFolder {
				parentID = sql.NullString{String: tt.selectedID, Valid: true}
			}
		}
		folder := models.NewTodoFolder(entry.Text, parentID)
		err := tt.folderRepo.Create(folder)
		if err != nil {
			dialog.ShowError(err, tt.win)
			return
		}
		tt.refreshData()
	}, tt.win)
}

// handleAdd: if called from a folder, add a todo; if from a todo, add a step
func (tt *TodosTab) handleAdd(parentID string) {
	item, exists := tt.items[parentID]
	if !exists {
		return
	}
	switch item.Type {
	case TypeFolder:
		tt.createTodo(parentID)
	case TypeTodo:
		// Focus step entry in editor
		tt.editor.addStepEntry.FocusGained() // but we need to ensure the todo is selected
		if tt.selectedID != parentID {
			tt.tree.Select(parentID)
		}
	}
}

func (tt *TodosTab) createTodo(folderID string) {
	entry := widget.NewEntry()
	dialog.ShowForm("New Todo", "Create", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Title", entry),
	}, func(ok bool) {
		if !ok || entry.Text == "" {
			return
		}
		todo := models.NewTodo(entry.Text, folderID)
		err := tt.todoRepo.Create(todo)
		if err != nil {
			dialog.ShowError(err, tt.win)
			return
		}
		tt.refreshData()
		// Optionally select the new todo
		tt.tree.Select(todo.ID)
	}, tt.win)
}

func (tt *TodosTab) renameItem(id string, currentName string) {
	entry := widget.NewEntry()
	entry.SetText(currentName)
	dialog.ShowForm("Rename", "Save", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Name", entry),
	}, func(ok bool) {
		if !ok || entry.Text == "" || entry.Text == currentName {
			return
		}
		item := tt.items[id]
		var err error
		if item.Type == TypeFolder {
			folder, _ := tt.folderRepo.GetByID(id)
			folder.Name = entry.Text
			err = tt.folderRepo.Update(folder)
		} else {
			todo, _ := tt.todoRepo.GetByID(id)
			todo.Title = entry.Text
			err = tt.todoRepo.Update(todo)
		}
		if err != nil {
			dialog.ShowError(err, tt.win)
			return
		}
		tt.refreshData()
	}, tt.win)
}

func (tt *TodosTab) deleteItem(id string, itemType ItemType) {
	var typeStr string
	if itemType == TypeFolder {
		typeStr = "folder"
	} else {
		typeStr = "todo"
	}
	dialog.ShowConfirm("Delete", "Delete this "+typeStr+"?", func(ok bool) {
		if !ok {
			return
		}
		var err error
		if itemType == TypeFolder {
			err = tt.folderRepo.Delete(id)
		} else {
			err = tt.todoRepo.Delete(id)
		}
		if err != nil {
			dialog.ShowError(err, tt.win)
			return
		}
		if itemType == TypeTodo && tt.editor.currentTodoID == id {
			tt.editor.Clear()
		}
		tt.refreshData()
	}, tt.win)
}

func (tt *TodosTab) selectTodo(todoID string) {
	tt.editor.LoadTodo(todoID)
}
func (nt *TodosTab) sortChildren(ids []string) {
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
