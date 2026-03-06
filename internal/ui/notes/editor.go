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

// minRichText is a RichText with a tiny minimum size to prevent window resizing.
type minRichText struct {
	widget.RichText
}

func newMinRichText() *minRichText {
	base := widget.NewRichText()
	return &minRichText{RichText: *base}
}

type NoteEditor struct {
	noteRepo          repository.NoteRepository
	win               fyne.Window
	entry             *widget.Entry
	preview           *minRichText
	scrollablePreview *container.Scroll
	stack             *fyne.Container
	saveBtn           *widget.Button
	previewBtn        *widget.Button
	currentNoteID     string
	content           fyne.CanvasObject
}

func NewNoteEditor(noteRepo repository.NoteRepository, win fyne.Window) *NoteEditor {
	ne := &NoteEditor{
		noteRepo: noteRepo,
		win:      win,
	}

	// Entry with small minimum size
	ne.entry = widget.NewMultiLineEntry()
	ne.entry.Wrapping = fyne.TextWrapBreak // Break long words to keep min width small
	// ne.entry.MinRowsVisible = 1            // Force small min height

	// Preview with tiny minimum size
	ne.preview = newMinRichText()
	ne.scrollablePreview = container.NewScroll(ne.preview)

	// Stack: both widgets, only one visible at a time
	ne.stack = container.NewStack(ne.scrollablePreview, ne.entry)
	ne.preview.Hide() // start in edit mode

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
		ne.preview.Scroll = container.ScrollNone
		ne.entry.Show()
		ne.entry.Scroll = container.ScrollBoth
		ne.previewBtn.SetText("Preview")
	} else {
		ne.preview.ParseMarkdown(ne.entry.Text) // Update preview content
		ne.entry.Hide()
		ne.entry.Scroll = container.ScrollNone
		ne.preview.Show()
		ne.preview.Scroll = container.ScrollBoth
		ne.previewBtn.SetText("Edit")
	}
	ne.stack.Refresh() // Ensure layout updates
}
