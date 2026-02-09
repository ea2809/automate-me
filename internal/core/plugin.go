package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type PluginRecord struct {
	Path       string
	Scope      PluginScope
	Manifest   Manifest
	DirectExec bool
}

type TaskRecord struct {
	PluginID    string
	PluginTitle string
	Task        TaskSpec
	Scope       PluginScope
	PluginPath  string
	DirectExec  bool
}

func LoadPlugins(repoRoot string) ([]PluginRecord, error) {
	candidates, err := discoverPluginCandidates(repoRoot)
	if err != nil {
		return nil, err
	}
	specs, err := loadSpecs(repoRoot)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]PluginRecord)
	for _, candidate := range candidates {
		manifest, err := describePlugin(candidate.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %s describe failed: %v\n", candidate.Path, err)
			continue
		}
		record := PluginRecord{Path: candidate.Path, Scope: candidate.Scope, Manifest: manifest, DirectExec: false}
		if existing, ok := byID[manifest.Plugin.ID]; ok {
			if existing.Scope == ScopeLocal && candidate.Scope == ScopeGlobal {
				continue
			}
			if existing.Scope == ScopeGlobal && candidate.Scope == ScopeLocal {
				fmt.Fprintf(os.Stderr, "warning: plugin id %s overridden by local %s (was %s)\n", manifest.Plugin.ID, candidate.Path, existing.Path)
			}
		}
		byID[manifest.Plugin.ID] = record
	}
	for _, spec := range specs {
		if existing, ok := byID[spec.Manifest.Plugin.ID]; ok {
			if existing.Scope == ScopeLocal && spec.Scope == ScopeGlobal {
				continue
			}
			if existing.Scope == ScopeGlobal && spec.Scope == ScopeLocal {
				fmt.Fprintf(os.Stderr, "warning: plugin id %s overridden by local spec %s (was %s)\n", spec.Manifest.Plugin.ID, spec.Path, existing.Path)
			}
		}
		byID[spec.Manifest.Plugin.ID] = spec
	}
	var plugins []PluginRecord
	for _, plugin := range byID {
		plugins = append(plugins, plugin)
	}
	return plugins, nil
}

func describePlugin(path string) (Manifest, error) {
	cmd := exec.Command(path, "describe")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return Manifest{}, err
	}
	return ParseManifest(stdout.Bytes())
}

func BuildTasks(plugins []PluginRecord) []TaskRecord {
	var tasks []TaskRecord
	for _, plugin := range plugins {
		for _, task := range plugin.Manifest.Tasks {
			tasks = append(tasks, TaskRecord{
				PluginID:    plugin.Manifest.Plugin.ID,
				PluginTitle: plugin.Manifest.Plugin.Title,
				Task:        task,
				Scope:       plugin.Scope,
				PluginPath:  plugin.Path,
				DirectExec:  plugin.DirectExec,
			})
		}
	}
	return tasks
}

func RunPluginTask(task TaskRecord, repoRoot, cwd string, args map[string]any) error {
	input := map[string]any{
		"args": args,
		"ctx": map[string]any{
			"repoRoot":       repoRoot,
			"cwd":            cwd,
			"selectedTaskId": TaskID(task.PluginID, task.Task.Name),
		},
	}
	payload, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("encode input JSON: %w", err)
	}
	var cmd *exec.Cmd
	if task.DirectExec {
		cmd = exec.Command(task.PluginPath)
		if repoRoot != "" {
			cmd.Dir = repoRoot
		}
	} else {
		cmd = exec.Command(task.PluginPath, "run", task.Task.Name)
	}
	cmd.Stdin = bytes.NewReader(payload)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), pluginEnv(task, repoRoot, cwd)...)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func pluginEnv(task TaskRecord, repoRoot, cwd string) []string {
	taskID := TaskID(task.PluginID, task.Task.Name)
	return []string{
		"AUTOMATE_ME_REPO_ROOT=" + repoRoot,
		"AUTOMATE_ME_CWD=" + cwd,
		"AUTOMATE_ME_TASK_ID=" + taskID,
		"AUTOMATE_ME_PLUGIN_ID=" + task.PluginID,
		"AUTOMATE_ME_TASK_NAME=" + task.Task.Name,
		"AUTOMATE_ME_SCOPE=" + string(task.Scope),
	}
}

func TaskID(pluginID, taskName string) string {
	return fmt.Sprintf("%s:%s", pluginID, taskName)
}
