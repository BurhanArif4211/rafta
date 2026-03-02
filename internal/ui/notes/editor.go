package notes

import (
	"github.com/burhanarif4211/rafta/internal/models"
	"github.com/burhanarif4211/rafta/internal/repository"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type NoteEditor struct {
	noteRepo      repository.NoteRepository
	win           fyne.Window
	entry         *widget.Entry
	preview       *widget.RichText
	stack         *fyne.Container
	saveBtn       *widget.Button
	previewBtn    *widget.Button
	currentNoteID string
	content       fyne.CanvasObject
}

func NewNoteEditor(noteRepo repository.NoteRepository, win fyne.Window) *NoteEditor {
	ne := &NoteEditor{
		noteRepo: noteRepo,
		win:      win,
	}
	ne.entry = widget.NewMultiLineEntry()
	ne.entry.Wrapping = fyne.TextWrapWord
	ne.preview = widget.NewRichText()
	ne.preview.Wrapping = fyne.TextWrapWord
	ne.stack = container.NewStack(ne.entry, ne.preview)
	ne.preview.Hide()

	ne.saveBtn = widget.NewButtonWithIcon("", theme.DocumentSaveIcon(), ne.save)
	ne.previewBtn = widget.NewButtonWithIcon("", theme.MenuIcon(), ne.togglePreview)
	toolbar := container.NewHBox(ne.saveBtn, ne.previewBtn)
	ne.content = container.NewBorder(toolbar, nil, nil, nil, ne.stack)
	return ne
}

func (ne *NoteEditor) Content() fyne.CanvasObject {
	return ne.content
}

func (ne *NoteEditor) LoadNote(note *models.Note) {
	ne.currentNoteID = note.ID
	ne.entry.SetText(note.Content)
}

func (ne *NoteEditor) Clear() {
	ne.currentNoteID = ""
	ne.entry.SetText("")
}

func (ne *NoteEditor) CurrentNoteID() string {
	return ne.currentNoteID
}

func (ne *NoteEditor) save() {
	if ne.currentNoteID == "" {
		dialog.ShowInformation("No note", "No note is currently open.", ne.win)
		return
	}
	note, err := ne.noteRepo.GetByID(ne.currentNoteID)
	if err != nil {
		dialog.ShowError(err, ne.win)
		return
	}
	note.Content = ne.entry.Text
	err = ne.noteRepo.Update(note)
	if err != nil {
		dialog.ShowError(err, ne.win)
	} else {
		dialog.ShowInformation("Saved", "Note saved.", ne.win)
	}
}

func (ne *NoteEditor) togglePreview() {
	if ne.preview.Visible() {
		ne.preview.Hide()
		ne.entry.Show()
		ne.previewBtn.SetText("Preview")
	} else {
		ne.preview.ParseMarkdown(ne.entry.Text) // Fyne's RichText can parse markdown
		ne.entry.Hide()
		ne.preview.Show()
		ne.previewBtn.SetText("Edit")
	}
	ne.stack.Refresh()
}
