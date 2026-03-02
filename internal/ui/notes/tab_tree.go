package notes

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type treeRow struct {
	widget.BaseWidget
	id        string
	itemType  ItemType
	label     *widget.Label
	addBtn    *widget.Button
	renameBtn *widget.Button
	deleteBtn *widget.Button
	container *fyne.Container
	onAddNote func(string)
	onRename  func(string, string)
	onDelete  func(string, ItemType)
}

func newTreeRow(branch bool, onAddNote func(string), onRename func(string, string), onDelete func(string, ItemType)) *treeRow {
	tr := &treeRow{
		label:     widget.NewLabel(""),
		onAddNote: onAddNote,
		onRename:  onRename,
		onDelete:  onDelete,
	}
	tr.ExtendBaseWidget(tr)

	tr.renameBtn = widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() {
		if tr.id != "" {
			tr.onRename(tr.id, tr.label.Text)
		}
	})
	tr.deleteBtn = widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		if tr.id != "" {
			tr.onDelete(tr.id, tr.itemType)
		}
	})

	if branch { // folder
		tr.addBtn = widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
			if tr.id != "" && tr.itemType == TypeFolder {
				tr.onAddNote(tr.id)
			}
		})

		tr.container = container.NewHBox(
			widget.NewIcon(theme.FolderIcon()),
			tr.label,
			tr.addBtn,
			tr.renameBtn,
			tr.deleteBtn,
		)
	} else { // note
		tr.container = container.NewHBox(
			widget.NewIcon(theme.DocumentIcon()),
			tr.label,
			tr.renameBtn,
			tr.deleteBtn,
		)
	}
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
