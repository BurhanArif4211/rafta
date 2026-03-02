package todos

import (
	"github.com/burhanarif4211/rafta/internal/models"
	"github.com/burhanarif4211/rafta/internal/repository"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	advancedlist "github.com/dweymouth/fyne-advanced-list"
)

// TodoEditor displays and manages steps for the selected todo.
//
// Implementation notes:
//   - Each step description has a binding.String which is stored in te.bindings[step.ID].
//   - The advancedlist template widgets are reused by that package; the update function
//     calls entry.Bind(...) with the correct binding for the step currently occupying
//     that row, so visual widgets may be reused safely without cross-contaminating text.
type TodoEditor struct {
	todoRepo repository.TodoRepository
	stepRepo repository.TodoStepRepository
	win      fyne.Window

	currentTodoID string
	steps         []*models.TodoStep
	bindings      map[string]binding.String // stepID -> binding for description

	stepList     *advancedlist.List
	addStepEntry *widget.Entry
	content      fyne.CanvasObject
}

func NewTodoEditor(todoRepo repository.TodoRepository, stepRepo repository.TodoStepRepository, win fyne.Window) *TodoEditor {
	te := &TodoEditor{
		todoRepo: todoRepo,
		stepRepo: stepRepo,
		win:      win,
		steps:    []*models.TodoStep{},
		bindings: make(map[string]binding.String),
	}

	// advancedlist: create
	te.stepList = advancedlist.NewList(
		// length function
		func() int { return len(te.steps) },
		// create template row (this will be reused)
		func() fyne.CanvasObject {
			// drag handle
			handle := widget.NewLabel("::")

			// checkbox
			check := widget.NewCheck("", nil)

			// multi-line entry (expands)
			entry := widget.NewEntry()
			// entry.Wrapping = fyne.TextWrapWord
			entry.Resize(fyne.NewSize(200, 40)) // ensure a reasonable minimum width/height

			// delete button
			del := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)

			// Layout: left (handle + check) | center (entry expands) | right (delete)
			left := container.NewHBox(handle, check)
			// put entry in border so it expands naturally
			center := container.NewHBox(entry)
			row := container.NewHBox(left, del, center)

			return row
		},
		// update function for a visible row
		func(id advancedlist.ListItemID, obj fyne.CanvasObject) {
			// safety checks
			if id < 0 || int(id) >= len(te.steps) {
				return
			}
			step := te.steps[id]

			// our template row is an HBox: [leftContainer, centerBorder, delBtn]
			row := obj.(*fyne.Container)
			if len(row.Objects) < 3 {
				return
			}

			// extract left and center
			left := row.Objects[0].(*fyne.Container)   // handle + check
			center := row.Objects[2].(*fyne.Container) // Border with entry inside
			delObj := row.Objects[1]                   // delete button

			// extract widgets defensively
			var check *widget.Check
			if len(left.Objects) >= 2 {
				if c, ok := left.Objects[1].(*widget.Check); ok {
					check = c
				}
			}
			var entry *widget.Entry
			// center is a Border; the entry is its first (and only) child
			if len(center.Objects) >= 1 {
				if e, ok := center.Objects[0].(*widget.Entry); ok {
					entry = e
				}
			}
			var delBtn *widget.Button
			if b, ok := delObj.(*widget.Button); ok {
				delBtn = b
			}
			if check == nil || entry == nil || delBtn == nil {
				return
			}

			// Bind checkbox and entry to the step.
			// Use SetChecked / Bind (not direct Text assignment) to keep widgets consistent.
			check.SetChecked(step.Completed)

			// Ensure there is a binding for this step description.
			b, ok := te.bindings[step.ID]
			if !ok {
				b = binding.NewString()
				_ = b.Set(step.Description) // ignore error; Set returns nil normally
				te.bindings[step.ID] = b
			} else {
				// keep binding value in sync with model (in case model changed externally)
				_ = b.Set(step.Description)
			}

			// Bind the entry to the step binding. Bind overrides entry text display.
			// If previously bound to another binding, calling Bind with the new binding is fine.
			entry.Bind(b)

			// capture local copy of step for closures (avoids reuse/closure bugs)
			s := step
			myBinding := b // capture binding too

			// Checkbox change: update model & DB
			check.OnChanged = func(checked bool) {
				// update model
				s.Completed = checked
				// persist
				if err := te.stepRepo.Update(s); err != nil {
					dialog.ShowError(err, te.win)
					// reload to keep consistent state
					te.loadSteps()
					return
				}
				// reflect UI
				te.stepList.Refresh()
			}

			// Entry changed -> update binding and DB.
			// We use OnChanged rather than OnSubmitted because this is a multi-line editor.
			entry.OnChanged = func(text string) {
				// avoid unnecessary DB writes when value hasn't changed
				cur, _ := myBinding.Get()
				if cur == text {
					return
				}
				// update binding first (keeps UI consistent)
				_ = myBinding.Set(text)
				// update model
				s.Description = text
				if err := te.stepRepo.Update(s); err != nil {
					dialog.ShowError(err, te.win)
					// reload from DB to restore consistent state
					te.loadSteps()
					return
				}
			}

			// delete tapped
			delBtn.OnTapped = func() {
				te.deleteStep(s.ID)
			}
		},
	)

	// enable drag and handle reordering
	te.stepList.EnableDragging = true
	te.stepList.OnDragEnd = func(draggedFrom, droppedAt advancedlist.ListItemID) {
		te.reorderSteps(draggedFrom, droppedAt)
	}

	// Add step entry and controls
	te.addStepEntry = widget.NewEntry()
	te.addStepEntry.SetPlaceHolder("New step...")
	addBtn := widget.NewButtonWithIcon("Add", theme.ContentAddIcon(), te.addStep)

	topBar := container.NewHBox(widget.NewLabel("Steps"))
	addBar := container.NewBorder(nil, nil, nil, addBtn, te.addStepEntry)

	te.content = container.NewBorder(topBar, addBar, nil, nil, container.NewScroll(te.stepList))
	return te
}

// Content returns the editor UI
func (te *TodoEditor) Content() fyne.CanvasObject {
	return te.content
}

// LoadTodo loads step data for the provided todo id.
func (te *TodoEditor) LoadTodo(todoID string) {
	te.currentTodoID = todoID
	te.loadSteps()
}

// Clear resets the editor
func (te *TodoEditor) Clear() {
	te.currentTodoID = ""
	te.steps = nil
	// clear bindings
	te.bindings = make(map[string]binding.String)
	te.stepList.Refresh()
}

// loadSteps fetches steps and rebuilds bindings for them
func (te *TodoEditor) loadSteps() {
	if te.currentTodoID == "" {
		te.Clear()
		return
	}
	steps, err := te.stepRepo.GetByTodo(te.currentTodoID)
	if err != nil {
		dialog.ShowError(err, te.win)
		return
	}
	te.steps = steps

	// Rebuild bindings map conservatively:
	// - Keep existing binding objects for step IDs that still exist (so any focused edits keep going)
	// - Create new binding objects for new steps
	// - Remove bindings for steps that no longer exist
	newBindings := make(map[string]binding.String, len(steps))
	for _, s := range steps {
		if b, ok := te.bindings[s.ID]; ok {
			// ensure binding holds current description
			_ = b.Set(s.Description)
			newBindings[s.ID] = b
		} else {
			b := binding.NewString()
			_ = b.Set(s.Description)
			newBindings[s.ID] = b
		}
	}
	te.bindings = newBindings

	te.stepList.Refresh()
}

// addStep creates a new step using the addStepEntry content
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
	if err := te.stepRepo.Create(step); err != nil {
		dialog.ShowError(err, te.win)
		return
	}
	te.addStepEntry.SetText("")
	// reload (this rebuilds bindings and refreshes UI)
	te.loadSteps()
}

// deleteStep asks for confirmation and deletes the step
func (te *TodoEditor) deleteStep(stepID string) {
	dialog.ShowConfirm("Delete Step", "Delete this step?", func(ok bool) {
		if !ok {
			return
		}
		if err := te.stepRepo.Delete(stepID); err != nil {
			dialog.ShowError(err, te.win)
			return
		}
		// remove binding if exists
		delete(te.bindings, stepID)
		te.loadSteps()
	}, te.win)
}

// reorderSteps updates the slice and persists new order.
func (te *TodoEditor) reorderSteps(draggedFrom, droppedAt int) {
	// validate
	if draggedFrom < 0 || draggedFrom >= len(te.steps) || droppedAt < 0 || droppedAt > len(te.steps) {
		return
	}
	// if same, nothing to do
	if draggedFrom == droppedAt || draggedFrom == droppedAt-1 {
		// special-case: when dragging down one position advancedlist may return droppedAt=from+1
		return
	}

	// advancedlist's droppedAt is the insertion index; adjust when moving forward
	if draggedFrom < droppedAt {
		droppedAt--
	}

	// move element
	step := te.steps[draggedFrom]
	// remove element at draggedFrom
	te.steps = append(te.steps[:draggedFrom], te.steps[draggedFrom+1:]...)
	// insert at droppedAt
	if droppedAt >= len(te.steps) {
		te.steps = append(te.steps, step)
	} else {
		te.steps = append(te.steps[:droppedAt], append([]*models.TodoStep{step}, te.steps[droppedAt:]...)...)
	}

	// update display order
	for i, s := range te.steps {
		s.DisplayOrder = i + 1
	}

	// persist order
	if err := te.stepRepo.Reorder(te.currentTodoID, te.getStepIDs()); err != nil {
		dialog.ShowError(err, te.win)
		// reload from DB to be safe
		te.loadSteps()
		return
	}

	// Finally, Refresh the list
	te.stepList.Refresh()
}

func (te *TodoEditor) getStepIDs() []string {
	ids := make([]string, len(te.steps))
	for i, s := range te.steps {
		ids[i] = s.ID
	}
	return ids
}
