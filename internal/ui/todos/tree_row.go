package todos

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ItemType int

const (
	TypeFolder ItemType = iota
	TypeTodo
)

type treeRow struct {
	widget.BaseWidget
	id        string
	itemType  ItemType
	label     *widget.Label
	moreBtn   *widget.Button
	container *fyne.Container
	onAdd     func(string)
	onRename  func(string, string)
	onDelete  func(string, ItemType)
	win       fyne.Window
}

func newTreeRow(branch bool, onAdd func(string), onRename func(string, string), onDelete func(string, ItemType), win fyne.Window) *treeRow {
	tr := &treeRow{
		label:    widget.NewLabel(""),
		onAdd:    onAdd,
		onRename: onRename,
		onDelete: onDelete,
		win:      win,
	}
	tr.ExtendBaseWidget(tr)

	var icon fyne.Resource
	if branch {
		icon = theme.FolderIcon()
	} else {
		icon = theme.ListIcon() // or DocumentIcon?
	}
	iconWidget := widget.NewIcon(icon)

	tr.moreBtn = widget.NewButtonWithIcon("", theme.MoreHorizontalIcon(), tr.showMenu)

	tr.container = container.NewHBox(
		iconWidget,
		tr.label,
		layout.NewSpacer(),
		tr.moreBtn,
	)
	return tr
}

func (tr *treeRow) SetItem(id string, name string, itemType ItemType) {
	tr.id = id
	tr.itemType = itemType
	tr.label.SetText(name)
}

func (tr *treeRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(tr.container)
}

func (tr *treeRow) showMenu() {
	items := []*fyne.MenuItem{}

	if tr.itemType == TypeFolder {
		addItem := fyne.NewMenuItem("Add Todo", func() {
			if tr.onAdd != nil {
				tr.onAdd(tr.id)
			}
		})
		addItem.Icon = theme.ContentAddIcon()
		items = append(items, addItem)
	} else if tr.itemType == TypeTodo {
		addStepItem := fyne.NewMenuItem("Add Step", func() {
			if tr.onAdd != nil {
				tr.onAdd(tr.id)
			}
		})
		addStepItem.Icon = theme.ContentAddIcon()
		items = append(items, addStepItem)
	}

	renameItem := fyne.NewMenuItem("Rename", func() {
		if tr.onRename != nil {
			tr.onRename(tr.id, tr.label.Text)
		}
	})
	renameItem.Icon = theme.DocumentCreateIcon()
	items = append(items, renameItem)

	deleteItem := fyne.NewMenuItem("Delete", func() {
		if tr.onDelete != nil {
			tr.onDelete(tr.id, tr.itemType)
		}
	})
	deleteItem.Icon = theme.DeleteIcon()
	items = append(items, deleteItem)

	menu := fyne.NewMenu("", items...)

	popup := widget.NewPopUpMenu(menu, tr.win.Canvas())
	popup.ShowAtPosition(tr.moreBtn.Position().Add(fyne.NewPos(0, tr.moreBtn.Size().Height)))
}
