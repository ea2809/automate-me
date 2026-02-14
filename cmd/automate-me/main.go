package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/ea2809/automate-me/internal/app"
	"github.com/ea2809/automate-me/internal/ui"
)

func main() {
	if err := internalRun(os.Args, ui.NewBubbleUI()); err != nil {
		if errors.Is(err, app.ErrUserCanceled) {
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func internalRun(argv []string, uiDriver app.UI) error {
	args := argv[1:]
	if len(args) == 0 {
		return app.RunInteractive(uiDriver)
	}

	switch args[0] {
	case "run":
		if len(args) < 2 {
			return fmt.Errorf("usage: %s run <taskId>", app.AppName)
		}
		return app.RunTaskByID(uiDriver, args[1])
	case "list":
		return app.ListTasksWithWriter(os.Stdout)
	case "plugins":
		return app.ListPluginsWithWriter(os.Stdout)
	case "import":
		return app.ImportSpec(args[1:])
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func printUsage() {
	fmt.Printf(`%s

Usage:
  %s            Start interactive TUI
  %s run <id>   Run task by id (plugin:task)
  %s list       List tasks
  %s plugins    List discovered plugins
  %s import     Import a JSON spec
`, app.AppName, app.AppName, app.AppName, app.AppName, app.AppName, app.AppName)
}
