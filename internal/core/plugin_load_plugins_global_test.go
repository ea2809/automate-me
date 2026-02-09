package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPluginsNoRepoRootLoadsGlobalSpecs(t *testing.T) {
	base := t.TempDir()
	configDir := filepath.Join(base, "config")
	globalSpecDir := filepath.Join(configDir, "automate-me", "specs")
	if err := os.MkdirAll(globalSpecDir, 0o755); err != nil {
		t.Fatal(err)
	}

	spec := `{"schemaVersion":1,"plugin":{"id":"p","title":"global","exec":"/bin/echo"},"tasks":[{"name":"t","title":"t"}]}`
	if err := os.WriteFile(filepath.Join(globalSpecDir, "p.json"), []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}

	os.Setenv("XDG_CONFIG_HOME", configDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")

	plugins, err := LoadPlugins("")
	if err != nil {
		t.Fatal(err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0].Manifest.Plugin.Title != "global" {
		t.Fatalf("expected global plugin, got %s", plugins[0].Manifest.Plugin.Title)
	}
}
