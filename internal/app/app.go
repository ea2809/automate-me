package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/ea2809/automate-me/internal/core"
)

const AppName = "automate-me"

// ErrUserCanceled bubbles up from UI input.
var ErrUserCanceled = errors.New("user canceled")

// ErrRefresh indicates the UI requested a refresh.
var ErrRefresh = errors.New("refresh requested")

func RunInteractive(uiDriver UI) error {
	uiDriver.ClearScreen()
	cwd, err := getwd()
	if err != nil {
		return err
	}
	repoRoot, err := resolveRepoRoot(cwd)
	if err != nil {
		return err
	}
	state := SelectionState{}
	lastArgs := make(map[string]map[string]any)
	tasks, err := refreshTasks(uiDriver, repoRoot)
	if err != nil {
		return err
	}
	return runInteractiveLoop(uiDriver, repoRoot, cwd, tasks, state, lastArgs)
}

func runInteractiveLoop(uiDriver UI, repoRoot, cwd string, tasks []core.TaskRecord, state SelectionState, lastArgs map[string]map[string]any) error {
	for {
		selected, updatedTasks, nextState, err := selectTaskWithRefresh(uiDriver, repoRoot, tasks, state)
		if err != nil {
			return err
		}
		tasks = updatedTasks
		state = nextState
		taskID, args, err := runSelectedTask(uiDriver, selected, repoRoot, cwd, lastArgs)
		if errors.Is(err, ErrUserCanceled) {
			uiDriver.ClearScreen()
			continue
		}
		if err != nil {
			return err
		}
		uiDriver.ClearScreen()
		lastArgs[taskID] = args
	}
}

func selectTaskWithRefresh(uiDriver UI, repoRoot string, tasks []core.TaskRecord, state SelectionState) (core.TaskRecord, []core.TaskRecord, SelectionState, error) {
	for {
		selected, nextState, err := uiDriver.SelectTask(tasks, state)
		if errors.Is(err, ErrRefresh) {
			updatedTasks, loadErr := refreshTasks(uiDriver, repoRoot)
			if loadErr != nil {
				return core.TaskRecord{}, tasks, nextState, loadErr
			}
			tasks = updatedTasks
			state = nextState
			continue
		}
		if err != nil {
			return core.TaskRecord{}, tasks, nextState, err
		}
		return selected, tasks, nextState, nil
	}
}

func runSelectedTask(uiDriver UI, selected core.TaskRecord, repoRoot, cwd string, lastArgs map[string]map[string]any) (string, map[string]any, error) {
	taskID := core.TaskID(selected.PluginID, selected.Task.Name)
	args, err := uiDriver.PromptInputs(selected.Task.Inputs, lastArgs[taskID])
	if err != nil {
		return "", nil, err
	}
	uiDriver.ClearScreen()
	uiDriver.RenderRunning(taskID, selected.PluginTitle)
	if err := core.RunPluginTask(selected, repoRoot, cwd, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if err := uiDriver.WaitForEnter(); err != nil {
		return "", nil, err
	}
	return taskID, args, nil
}

func RunTaskByID(uiDriver UI, id string) error {
	cwd, err := getwd()
	if err != nil {
		return err
	}
	repoRoot, tasks, err := resolveRepoAndTasks(cwd)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if core.TaskID(task.PluginID, task.Task.Name) == id {
			args, err := uiDriver.PromptInputs(task.Task.Inputs, nil)
			if err != nil {
				return err
			}
			return core.RunPluginTask(task, repoRoot, cwd, args)
		}
	}
	return fmt.Errorf("task not found: %s", id)
}

func ListTasksWithWriter(writer io.Writer) error {
	_, tasks, err := currentRepoAndTasks()
	if err != nil {
		return err
	}
	for _, task := range tasks {
		fmt.Fprintf(writer, "%s\t%s\t%s\n", core.TaskID(task.PluginID, task.Task.Name), task.Task.Title, task.Task.Description)
	}
	return nil
}

func ListPluginsWithWriter(writer io.Writer) error {
	repoRoot, err := currentRepoRoot()
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

func refreshTasks(uiDriver UI, repoRoot string) ([]core.TaskRecord, error) {
	uiDriver.ClearScreen()
	uiDriver.RenderLoading("Loading tasks...")
	tasks, err := loadTasks(repoRoot)
	uiDriver.ClearScreen()
	return tasks, err
}

func loadTasks(repoRoot string) ([]core.TaskRecord, error) {
	plugins, err := core.LoadPlugins(repoRoot)
	if err != nil {
		return nil, err
	}
	tasks := core.BuildTasks(plugins)
	if len(tasks) == 0 {
		return nil, fmt.Errorf("no tasks found")
	}
	sortTasks(tasks)
	return tasks, nil
}

func sortTasks(tasks []core.TaskRecord) {
	sort.Slice(tasks, func(i, j int) bool {
		ai := core.TaskID(tasks[i].PluginID, tasks[i].Task.Name)
		aj := core.TaskID(tasks[j].PluginID, tasks[j].Task.Name)
		return ai < aj
	})
}
