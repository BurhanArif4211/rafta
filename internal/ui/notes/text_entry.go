package notes

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/widget"
)

// touchFriendlyEntry forwards large vertical drags to the surrounding Scroll.
// Put this entry inside a container.NewScroll(entry) and call newTouchFriendlyEntry(scroll).
type touchFriendlyEntry struct {
	widget.Entry
	scroll      *container.Scroll
	start       fyne.Position
	handedOff   bool // true when we've forwarded drag events to the scroll
	thresholdPx float32
}

func newTouchFriendlyEntry(scroll *container.Scroll) *touchFriendlyEntry {
	e := &touchFriendlyEntry{
		scroll:      scroll,
		thresholdPx: 8, // adjust sensitivity if needed
	}
	e.ExtendBaseWidget(e)
	// initialize as a multiline entry
	e.Entry = *widget.NewMultiLineEntry()
	return e
}

// TouchDown: record initial position and keep Entry behaviour (e.g. give focus)
func (e *touchFriendlyEntry) TouchDown(ev *mobile.TouchEvent) {
	e.start = ev.Position
	// ensure normal Entry touch behaviour (focus/cursor)
	e.Entry.TouchDown(ev)
	e.handedOff = false
}

// Dragged: if vertical drag dominates and exceeds threshold, forward as ScrollEvent.
// otherwise keep normal text selection dragging.
func (e *touchFriendlyEntry) Dragged(d *fyne.DragEvent) {
	// if already handed off, continue forwarding as scroll deltas
	if e.handedOff {
		if e.scroll != nil {
			ev := &fyne.ScrollEvent{
				PointEvent: d.PointEvent,
				Scrolled:   d.Dragged, // DX/DY movement since last event
			}
			e.scroll.Scrolled(ev)
		}
		return
	}

	// compute absolute movement from start
	dx := float32(d.PointEvent.Position.X - e.start.X)
	dy := float32(d.PointEvent.Position.Y - e.start.Y)
	adx := float32(math.Abs(float64(dx)))
	ady := float32(math.Abs(float64(dy)))

	// if vertical movement dominates and passes threshold, hand off to scroll
	if ady > adx && ady > e.thresholdPx && e.scroll != nil {
		e.handedOff = true
		// finish entry's drag handling (stop selection)
		e.Entry.DragEnd()
		// forward this drag event as a scroll
		ev := &fyne.ScrollEvent{
			PointEvent: d.PointEvent,
			Scrolled:   d.Dragged,
		}
		e.scroll.Scrolled(ev)
		return
	}

	// otherwise let Entry handle selection drag
	e.Entry.Dragged(d)
}

// DragEnd: finish whichever side we were using
func (e *touchFriendlyEntry) DragEnd() {
	if e.handedOff && e.scroll != nil {
		// emulate end-of-scroll by sending a zero or final small event if desired.
		// container.Scroll does not have DragEnd, so nothing else required.
	} else {
		e.Entry.DragEnd()
	}
	e.handedOff = false
}
