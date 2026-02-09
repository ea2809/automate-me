package app

import "github.com/ea2809/automate-me/internal/core"

type UI interface {
	ClearScreen()
	SelectTask(tasks []core.TaskRecord, state SelectionState) (core.TaskRecord, SelectionState, error)
	PromptInputs(inputs []core.InputSpec, defaults map[string]any) (map[string]any, error)
	RenderRunning(taskID, pluginTitle string)
	RenderLoading(message string)
	WaitForEnter() error
}

// SelectionState keeps the UI cursor and filter between runs.
// This is owned by app to keep UIs decoupled.
type SelectionState struct {
	Filter string
	Cursor int
}
