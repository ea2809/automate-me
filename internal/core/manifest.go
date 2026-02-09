package core

import (
	"encoding/json"
	"fmt"
)

type Manifest struct {
	SchemaVersion int        `json:"schemaVersion"`
	Plugin        PluginInfo `json:"plugin"`
	Tasks         []TaskSpec `json:"tasks"`
}

type PluginInfo struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Version  string `json:"version,omitempty"`
	Exec     string `json:"exec,omitempty"`
	ExecMode string `json:"execMode,omitempty"`
}

type TaskSpec struct {
	Name        string      `json:"name"`
	Title       string      `json:"title"`
	Group       string      `json:"group,omitempty"`
	Description string      `json:"description,omitempty"`
	Inputs      []InputSpec `json:"inputs,omitempty"`
}

type InputSpec struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	Prompt   string   `json:"prompt"`
	Default  any      `json:"default,omitempty"`
	Choices  []string `json:"choices,omitempty"`
	Secret   bool     `json:"secret,omitempty"`
}

func ParseManifest(data []byte) (Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("invalid manifest JSON: %w", err)
	}
	if m.SchemaVersion != 1 {
		return Manifest{}, fmt.Errorf("unsupported schemaVersion: %d", m.SchemaVersion)
	}
	if m.Plugin.ID == "" {
		return Manifest{}, fmt.Errorf("manifest missing plugin.id")
	}
	for i, task := range m.Tasks {
		if task.Name == "" {
			return Manifest{}, fmt.Errorf("task[%d] missing name", i)
		}
		if task.Title == "" {
			m.Tasks[i].Title = task.Name
		}
	}
	return m, nil
}
