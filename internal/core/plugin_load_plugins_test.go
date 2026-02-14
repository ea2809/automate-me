package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPluginsSpecPrecedence(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	paths := newPathConfig(repo)
	localSpecDir, err := paths.localSpecs()
	if err != nil {
		t.Fatal(err)
	}
	globalConfig := filepath.Join(base, "config")
	globalSpecDir := filepath.Join(globalConfig, globalConfigDirName, specsDirName)
	if err := os.MkdirAll(localSpecDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(globalSpecDir, 0o755); err != nil {
		t.Fatal(err)
	}

	localSpec := `{"schemaVersion":1,"plugin":{"id":"p","title":"local","exec":"/bin/echo"},"tasks":[{"name":"t","title":"t"}]}`
	globalSpec := `{"schemaVersion":1,"plugin":{"id":"p","title":"global","exec":"/bin/echo"},"tasks":[{"name":"t","title":"t"}]}`
	if err := os.WriteFile(filepath.Join(localSpecDir, "p.json"), []byte(localSpec), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(globalSpecDir, "p.json"), []byte(globalSpec), 0o644); err != nil {
		t.Fatal(err)
	}

	os.Setenv("XDG_CONFIG_HOME", globalConfig)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	plugins, err := LoadPlugins(repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0].Manifest.Plugin.Title != "local" {
		t.Fatalf("expected local spec to win, got %s", plugins[0].Manifest.Plugin.Title)
	}
}
