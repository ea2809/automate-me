package core

import "testing"

func TestParseManifestErrors(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{
			name: "invalid json",
			json: "{",
		},
		{
			name: "unsupported schema",
			json: `{"schemaVersion":2,"plugin":{"id":"p"},"tasks":[]}`,
		},
		{
			name: "missing plugin id",
			json: `{"schemaVersion":1,"plugin":{},"tasks":[]}`,
		},
		{
			name: "missing task name",
			json: `{"schemaVersion":1,"plugin":{"id":"p"},"tasks":[{"title":"t"}]}`,
		},
	}

	for _, tt := range tests {
		if _, err := ParseManifest([]byte(tt.json)); err == nil {
			t.Fatalf("%s: expected error", tt.name)
		}
	}
}

func TestParseManifestDefaultsTitle(t *testing.T) {
	manifest, err := ParseManifest([]byte(`{
  "schemaVersion": 1,
  "plugin": {"id": "p"},
  "tasks": [{"name": "t"}]
}`))
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Tasks[0].Title != "t" {
		t.Fatalf("expected title to default to name")
	}
}

func TestParseManifest(t *testing.T) {
	data := []byte(`{
  "schemaVersion": 1,
  "plugin": {"id": "repo", "title": "Repo"},
  "tasks": [{"name": "test", "title": "Run tests"}]
}`)
	manifest, err := ParseManifest(data)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Plugin.ID != "repo" {
		t.Fatalf("expected plugin id repo, got %s", manifest.Plugin.ID)
	}
}
