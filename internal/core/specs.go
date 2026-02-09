package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	localSpecDirName  = ".automate-me/specs"
	globalSpecDirName = "automate-me/specs"
)

func ImportSpecFile(srcPath, repoRoot string, scope PluginScope) (string, error) {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return "", fmt.Errorf("read spec: %w", err)
	}
	manifest, err := ParseManifest(data)
	if err != nil {
		return "", err
	}
	if manifest.Plugin.Exec == "" {
		return "", fmt.Errorf("imported spec requires plugin.exec")
	}

	destDir, err := specDir(repoRoot, scope)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", fmt.Errorf("create spec dir: %w", err)
	}
	fileName := sanitizeFilename(manifest.Plugin.ID) + ".json"
	destPath := filepath.Join(destDir, fileName)
	if err := os.WriteFile(destPath, data, 0o644); err != nil {
		return "", fmt.Errorf("write spec: %w", err)
	}
	return destPath, nil
}

func loadSpecs(repoRoot string) ([]PluginRecord, error) {
	var records []PluginRecord

	if repoRoot != "" {
		localDir, err := specDir(repoRoot, ScopeLocal)
		if err != nil {
			return nil, err
		}
		localSpecs, err := readSpecDir(localDir, ScopeLocal)
		if err != nil {
			return nil, err
		}
		records = append(records, localSpecs...)
	}

	globalDir, err := specDir(repoRoot, ScopeGlobal)
	if err != nil {
		return nil, err
	}
	globalSpecs, err := readSpecDir(globalDir, ScopeGlobal)
	if err != nil {
		return nil, err
	}
	records = append(records, globalSpecs...)

	return records, nil
}

func readSpecDir(dir string, scope PluginScope) ([]PluginRecord, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read spec dir %s: %w", dir, err)
	}
	var records []PluginRecord
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read spec %s: %w", path, err)
		}
		manifest, err := ParseManifest(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: invalid spec %s: %v\n", path, err)
			continue
		}
		if manifest.Plugin.Exec == "" {
			fmt.Fprintf(os.Stderr, "warning: spec %s missing plugin.exec\n", path)
			continue
		}
		directExec := true
		if strings.EqualFold(manifest.Plugin.ExecMode, "protocol") {
			directExec = false
		}
		records = append(records, PluginRecord{
			Path:       manifest.Plugin.Exec,
			Scope:      scope,
			Manifest:   manifest,
			DirectExec: directExec,
		})
	}
	return records, nil
}

func specDir(repoRoot string, scope PluginScope) (string, error) {
	switch scope {
	case ScopeLocal:
		if repoRoot == "" {
			return "", fmt.Errorf("no repo root for local spec")
		}
		return filepath.Join(repoRoot, localSpecDirName), nil
	case ScopeGlobal:
		configDir, err := configBaseDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(configDir, globalSpecDirName), nil
	default:
		return "", fmt.Errorf("unknown scope %q", scope)
	}
}

func sanitizeFilename(value string) string {
	if value == "" {
		return "spec"
	}
	var out strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			out.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			out.WriteRune(r)
		case r >= '0' && r <= '9':
			out.WriteRune(r)
		case r == '-' || r == '_' || r == '.':
			out.WriteRune(r)
		default:
			out.WriteRune('_')
		}
	}
	return out.String()
}
