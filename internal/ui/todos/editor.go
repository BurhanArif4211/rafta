package todos

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/burhanarif4211/rafta/internal/models"
	"github.com/burhanarif4211/rafta/internal/repository"
)

type TodoEditor struct {
	todoRepo repository.TodoRepository
	stepRepo repository.TodoStepRepository
	win      fyne.Window

	currentTodoID string
	steps         []*models.TodoStep
	stepList      *widget.List
	addStepEntry  *widget.Entry
	content       fyne.CanvasObject

	selectedStepIndex int // -1 means none
}

func NewTodoEditor(todoRepo repository.TodoRepository, stepRepo repository.TodoStepRepository, win fyne.Window) *TodoEditor {
	te := &TodoEditor{
		todoRepo:          todoRepo,
		stepRepo:          stepRepo,
		win:               win,
		steps:             []*models.TodoStep{},
		selectedStepIndex: -1,
	}

	// Step list with custom row layout
	te.stepList = widget.NewList(
		func() int { return len(te.steps) },
		func() fyne.CanvasObject {
			check := widget.NewCheck("", nil)
			label := widget.NewLabel("")
			editBtn := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
			deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
			// Left side: check + label
			left := container.NewHBox(check, label)
			// Right side: edit + delete
			right := container.NewHBox(editBtn, deleteBtn)
			// Border layout pushes left to left, right to right
			return container.NewBorder(nil, nil, left, right, nil)
		},
		func(i int, obj fyne.CanvasObject) {
			if i >= len(te.steps) {
				return
			}
			step := te.steps[i]
			border := obj.(*fyne.Container)
			// Border layout: objects[2]=left, objects[3]=right (top,bottom,center are nil)
			leftBox := border.Objects[0].(*fyne.Container)
			rightBox := border.Objects[1].(*fyne.Container)
			check := leftBox.Objects[0].(*widget.Check)
			label := leftBox.Objects[1].(*widget.Label)
			editBtn := rightBox.Objects[0].(*widget.Button)
			deleteBtn := rightBox.Objects[1].(*widget.Button)

			check.SetChecked(step.Completed)
			check.OnChanged = func(checked bool) {
				step.Completed = checked
				te.stepRepo.Update(step)
			}
			label.SetText(step.Description)

			editBtn.OnTapped = func() {
				te.editStep(step)
			}
			deleteBtn.OnTapped = func() {
				te.deleteStep(step.ID)
			}
		},
	)

	te.stepList.OnSelected = func(id widget.ListItemID) {
		te.selectedStepIndex = id
	}
	te.stepList.OnUnselected = func(id widget.ListItemID) {
		te.selectedStepIndex = -1
	}

	te.addStepEntry = widget.NewEntry()
	te.addStepEntry.SetPlaceHolder("New step...")
	addBtn := widget.NewButtonWithIcon("Add", theme.ContentAddIcon(), te.addStep)

	moveUpBtn := widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() {
		if te.selectedStepIndex <= 0 || te.selectedStepIndex >= len(te.steps) {
			return
		}
		te.moveStep(te.selectedStepIndex, te.selectedStepIndex-1)
	})
	moveDownBtn := widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() {
		if te.selectedStepIndex < 0 || te.selectedStepIndex >= len(te.steps)-1 {
			return
		}
		te.moveStep(te.selectedStepIndex, te.selectedStepIndex+1)
	})
	reorderBox := container.NewHBox(moveUpBtn, moveDownBtn)

	topBar := container.NewHBox(widget.NewLabel("Steps"), reorderBox)
	addBar := container.NewBorder(nil, nil, nil, addBtn, te.addStepEntry)

	te.content = container.NewBorder(topBar, addBar, nil, nil, container.NewScroll(te.stepList))
	return te
}

func (te *TodoEditor) Content() fyne.CanvasObject {
	return te.content
}

func (te *TodoEditor) LoadTodo(todoID string) {
	te.currentTodoID = todoID
	te.loadSteps()
}

func (te *TodoEditor) Clear() {
	te.currentTodoID = ""
	te.steps = nil
	te.stepList.Refresh()
	te.selectedStepIndex = -1
}

func (te *TodoEditor) loadSteps() {
	if te.currentTodoID == "" {
		te.steps = nil
		te.stepList.Refresh()
		return
	}
	steps, err := te.stepRepo.GetByTodo(te.currentTodoID)
	if err != nil {
		dialog.ShowError(err, te.win)
		return
	}
	te.steps = steps
	te.stepList.Refresh()
}

func (te *TodoEditor) addStep() {
	if te.currentTodoID == "" {
		dialog.ShowInformation("No todo", "Select a todo first.", te.win)
		return
	}
	desc := te.addStepEntry.Text
	if desc == "" {
		return
	}
	nextOrder := len(te.steps) + 1
	step := models.NewTodoStep(te.currentTodoID, desc, nextOrder)
	err := te.stepRepo.Create(step)
	if err != nil {
		dialog.ShowError(err, te.win)
		return
	}
	te.addStepEntry.SetText("")
	te.loadSteps()
}

func (te *TodoEditor) editStep(step *models.TodoStep) {
	entry := widget.NewEntry()
	entry.SetText(step.Description)
	dialog.ShowForm("Edit Step", "Save", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Description", entry),
	}, func(ok bool) {
		if !ok || entry.Text == "" || entry.Text == step.Description {
			return
		}
		step.Description = entry.Text
		err := te.stepRepo.Update(step)
		if err != nil {
			dialog.ShowError(err, te.win)
			return
		}
		te.loadSteps()
	}, te.win)
}

func (te *TodoEditor) deleteStep(stepID string) {
	dialog.ShowConfirm("Delete Step", "Delete this step?", func(ok bool) {
		if !ok {
			return
		}
		err := te.stepRepo.Delete(stepID)
		if err != nil {
			dialog.ShowError(err, te.win)
			return
		}
		te.loadSteps()
	}, te.win)
}

// moveStep moves a step from index 'from' to index 'to' (adjacent only, but general works)
func (te *TodoEditor) moveStep(from, to int) {
	if from == to || from < 0 || to < 0 || from >= len(te.steps) || to >= len(te.steps) {
		return
	}
	// Store the step we're moving
	movedStep := te.steps[from]
	stepID := movedStep.ID

	// Swap in slice (for adjacent moves, but works for any)
	te.steps = append(append(te.steps[:from], te.steps[from+1:]...), nil) // remove
	copy(te.steps[to+1:], te.steps[to:])                                  // make room
	te.steps[to] = movedStep                                              // insert

	// Update display_order
	for i, s := range te.steps {
		s.DisplayOrder = i + 1
	}

	// Persist new order
	err := te.stepRepo.Reorder(te.currentTodoID, te.getStepIDs())
	if err != nil {
		dialog.ShowError(err, te.win)
		te.loadSteps() // revert
		return
	}

	// Reload steps to get fresh objects and avoid closure issues
	te.loadSteps()

	// Find new index of the moved step and select it
	for i, s := range te.steps {
		if s.ID == stepID {
			te.stepList.Select(i)
			break
		}
	}
}

func (te *TodoEditor) getStepIDs() []string {
	ids := make([]string, len(te.steps))
	for i, s := range te.steps {
		ids[i] = s.ID
	}
	return ids
}
