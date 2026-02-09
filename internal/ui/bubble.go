package ui

import (
	"bufio"
	"fmt"
	"os"

	"github.com/ea2809/automate-me/internal/app"
	"github.com/ea2809/automate-me/internal/core"
)

type BubbleUI struct {
	theme Theme
}

func NewBubbleUI() *BubbleUI {
	return &BubbleUI{theme: DefaultTheme()}
}

func (b *BubbleUI) ClearScreen() {
	fmt.Print("\x1b[2J\x1b[H")
}

func (b *BubbleUI) SelectTask(tasks []core.TaskRecord, state app.SelectionState) (core.TaskRecord, app.SelectionState, error) {
	return SelectTask(tasks, state)
}

func (b *BubbleUI) PromptInputs(inputs []core.InputSpec, defaults map[string]any) (map[string]any, error) {
	return PromptInputsWithDefaults(inputs, defaults)
}

func (b *BubbleUI) RenderRunning(taskID, pluginTitle string) {
	fmt.Printf("%s %s\n", b.theme.Running.Render("Running"), b.theme.Dim.Render(taskID))
	fmt.Printf("%s %s\n\n", b.theme.Dim.Render("Plugin"), pluginTitle)
}

func (b *BubbleUI) RenderLoading(message string) {
	fmt.Printf("%s %s\n", b.theme.Loading.Render("Loading"), b.theme.Dim.Render(message))
}

func (b *BubbleUI) WaitForEnter() error {
	fmt.Print("\nPress Enter to return to the menu...")
	reader := bufio.NewReader(os.Stdin)
	_, err := reader.ReadString('\n')
	return err
}
