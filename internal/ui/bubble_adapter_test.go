package ui

import (
	"testing"

	"github.com/ea2809/automate-me/internal/app"
	"github.com/ea2809/automate-me/internal/core"
)

type fakeSelect struct {
	state app.SelectionState
}

func (f *fakeSelect) SelectTask(tasks []core.TaskRecord, state app.SelectionState) (core.TaskRecord, app.SelectionState, error) {
	f.state = state
	return core.TaskRecord{}, app.SelectionState{Filter: "keep", Cursor: 2}, nil
}

func TestBubbleUISelectTaskPassesState(t *testing.T) {
	ui := NewBubbleUI()
	original := SelectTask
	defer func() { SelectTask = original }()

	recorder := &fakeSelect{}
	SelectTask = recorder.SelectTask

	state := app.SelectionState{Filter: "q", Cursor: 1}
	_, next, err := ui.SelectTask([]core.TaskRecord{}, state)
	if err != nil {
		t.Fatal(err)
	}
	if recorder.state != state {
		t.Fatalf("expected state to be passed through")
	}
	if next.Filter != "keep" || next.Cursor != 2 {
		t.Fatalf("unexpected next state: %+v", next)
	}
}
