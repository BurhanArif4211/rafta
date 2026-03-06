package notes

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type treeRow struct {
	widget.BaseWidget
	id        string
	itemType  ItemType
	label     *widget.Label
	moreBtn   *widget.Button
	container *fyne.Container
	onAddNote func(string)
	onRename  func(string, string)
	onDelete  func(string, ItemType)
	win       fyne.Window
}

func newTreeRow(branch bool, onAddNote func(string), onRename func(string, string), onDelete func(string, ItemType), win fyne.Window) *treeRow {
	tr := &treeRow{
		label:     widget.NewLabel(""),
		onAddNote: onAddNote,
		onRename:  onRename,
		onDelete:  onDelete,
		win:       win,
	}
	tr.ExtendBaseWidget(tr)

	// Icon based on type
	var icon fyne.Resource
	if branch {
		icon = theme.FolderIcon()
	} else {
		icon = theme.DocumentIcon()
	}
	iconWidget := widget.NewIcon(icon)

	// More options button
	tr.moreBtn = widget.NewButtonWithIcon("", theme.MoreHorizontalIcon(), tr.showMenu)

	// Layout: icon | label (expands) | moreBtn
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
	// Build menu items based on item type
	items := []*fyne.MenuItem{}

	if tr.itemType == TypeFolder {
		// Add note option for folders
		addNoteItem := fyne.NewMenuItem("Add Note", func() {
			if tr.onAddNote != nil {
				tr.onAddNote(tr.id)
			}
		})
		addNoteItem.Icon = theme.ContentAddIcon()
		items = append(items, addNoteItem)
	}

	// Rename option (always present)
	renameItem := fyne.NewMenuItem("Rename", func() {
		if tr.onRename != nil {
			tr.onRename(tr.id, tr.label.Text)
		}
	})
	renameItem.Icon = theme.DocumentCreateIcon()
	items = append(items, renameItem)

	// Delete option (always present)
	deleteItem := fyne.NewMenuItem("Delete", func() {
		if tr.onDelete != nil {
			tr.onDelete(tr.id, tr.itemType)
		}
	})
	deleteItem.Icon = theme.DeleteIcon()
	items = append(items, deleteItem)

	menu := fyne.NewMenu("", items...)

	// Show popup menu anchored to the more button
	popup := widget.NewPopUpMenu(menu, tr.win.Canvas())
	popup.ShowAtPosition(tr.moreBtn.Position().Add(fyne.NewPos(0, tr.moreBtn.Size().Height)))
}
