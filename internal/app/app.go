package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/ea2809/automate-me/internal/core"
)

const appName = "automate-me"

// ErrUserCanceled bubbles up from UI input.
var ErrUserCanceled = errors.New("user canceled")

// ErrRefresh indicates the UI requested a refresh.
var ErrRefresh = errors.New("refresh requested")

func Run(argv []string, uiDriver UI) error {
	args := argv[1:]
	if len(args) == 0 {
		return runInteractive(uiDriver)
	}

	switch args[0] {
	case "run":
		if len(args) < 2 {
			return fmt.Errorf("usage: %s run <taskId>", appName)
		}
		return runTaskByID(uiDriver, args[1])
	case "list":
		return listTasksWithWriter(os.Stdout)
	case "plugins":
		return listPluginsWithWriter(os.Stdout)
	case "import":
		return importSpec(args[1:])
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func runInteractive(uiDriver UI) error {
	uiDriver.ClearScreen()
	repoRoot, _, err := core.FindRepoRoot(mustGetwd())
	if err != nil {
		return err
	}
	state := SelectionState{}
	lastArgs := make(map[string]map[string]any)
	tasks, err := loadTasks(uiDriver, repoRoot)
	if err != nil {
		return err
	}
	for {
		selected, nextState, err := uiDriver.SelectTask(tasks, state)
		if err != nil {
			if errors.Is(err, ErrRefresh) {
				tasks, err = loadTasks(uiDriver, repoRoot)
				if err != nil {
					return err
				}
				state = nextState
				continue
			}
			return err
		}
		state = nextState
		taskID := core.TaskID(selected.PluginID, selected.Task.Name)
		args, err := uiDriver.PromptInputs(selected.Task.Inputs, lastArgs[taskID])
		if err != nil {
			if errors.Is(err, ErrUserCanceled) {
				uiDriver.ClearScreen()
				continue
			}
			return err
		}
		uiDriver.ClearScreen()
		uiDriver.RenderRunning(taskID, selected.PluginTitle)
		if err := core.RunPluginTask(selected, repoRoot, mustGetwd(), args); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		if err := uiDriver.WaitForEnter(); err != nil {
			return err
		}
		uiDriver.ClearScreen()
		lastArgs[taskID] = args
	}
}

func runTaskByID(uiDriver UI, id string) error {
	repoRoot, _, err := core.FindRepoRoot(mustGetwd())
	if err != nil {
		return err
	}
	plugins, err := core.LoadPlugins(repoRoot)
	if err != nil {
		return err
	}
	tasks := core.BuildTasks(plugins)
	for _, task := range tasks {
		if core.TaskID(task.PluginID, task.Task.Name) == id {
			args, err := uiDriver.PromptInputs(task.Task.Inputs, nil)
			if err != nil {
				return err
			}
			return core.RunPluginTask(task, repoRoot, mustGetwd(), args)
		}
	}
	return fmt.Errorf("task not found: %s", id)
}

func listTasksWithWriter(writer io.Writer) error {
	repoRoot, _, err := core.FindRepoRoot(mustGetwd())
	if err != nil {
		return err
	}
	plugins, err := core.LoadPlugins(repoRoot)
	if err != nil {
		return err
	}
	tasks := core.BuildTasks(plugins)
	sortTasks(tasks)
	for _, task := range tasks {
		fmt.Fprintf(writer, "%s\t%s\t%s\n", core.TaskID(task.PluginID, task.Task.Name), task.Task.Title, task.Task.Description)
	}
	return nil
}

func listPluginsWithWriter(writer io.Writer) error {
	repoRoot, _, err := core.FindRepoRoot(mustGetwd())
	if err != nil {
		return err
	}
	plugins, err := core.LoadPlugins(repoRoot)
	if err != nil {
		return err
	}
	if len(plugins) == 0 {
		fmt.Fprintln(writer, "no plugins found")
		return nil
	}
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Manifest.Plugin.ID < plugins[j].Manifest.Plugin.ID
	})
	for _, plugin := range plugins {
		fmt.Fprintf(writer, "%s\t%s\t%s\n", plugin.Manifest.Plugin.ID, plugin.Scope, plugin.Path)
	}
	return nil
}

func loadTasks(uiDriver UI, repoRoot string) ([]core.TaskRecord, error) {
	uiDriver.ClearScreen()
	uiDriver.RenderLoading("Loading tasks...")
	plugins, err := core.LoadPlugins(repoRoot)
	if err != nil {
		return nil, err
	}
	tasks := core.BuildTasks(plugins)
	if len(tasks) == 0 {
		return nil, fmt.Errorf("no tasks found")
	}
	sortTasks(tasks)
	uiDriver.ClearScreen()
	return tasks, nil
}

func sortTasks(tasks []core.TaskRecord) {
	sort.Slice(tasks, func(i, j int) bool {
		ai := core.TaskID(tasks[i].PluginID, tasks[i].Task.Name)
		aj := core.TaskID(tasks[j].PluginID, tasks[j].Task.Name)
		return ai < aj
	})
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd
}

func printUsage() {
	fmt.Printf(`%s

Usage:
  %s            Start interactive TUI
  %s run <id>   Run task by id (plugin:task)
  %s list       List tasks
  %s plugins    List discovered plugins
  %s import     Import a JSON spec
`, appName, appName, appName, appName, appName, appName)
}
