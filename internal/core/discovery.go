package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type PluginScope string

const (
	ScopeLocal  PluginScope = "local"
	ScopeGlobal PluginScope = "global"
)

type pluginCandidate struct {
	Path  string
	Scope PluginScope
}

func FindRepoRoot(start string) (string, bool, error) {
	dir := start
	var foundGit string
	for {
		if exists(filepath.Join(dir, ".automate-me")) {
			return dir, true, nil
		}
		if foundGit == "" && exists(filepath.Join(dir, ".git")) {
			foundGit = dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if foundGit != "" {
		return foundGit, true, nil
	}
	return "", false, nil
}

func discoverPluginCandidates(repoRoot string) ([]pluginCandidate, error) {
	var candidates []pluginCandidate
	if repoRoot != "" {
		localDir := filepath.Join(repoRoot, ".automate-me", "bin")
		locals, err := findExecutables(localDir)
		if err != nil {
			return nil, err
		}
		for _, path := range locals {
			candidates = append(candidates, pluginCandidate{Path: path, Scope: ScopeLocal})
		}
	}

	configDir, err := configBaseDir()
	if err != nil {
		return nil, err
	}
	globalDir := filepath.Join(configDir, "automate-me", "bin")
	globals, err := findExecutables(globalDir)
	if err != nil {
		return nil, err
	}
	for _, path := range globals {
		candidates = append(candidates, pluginCandidate{Path: path, Scope: ScopeGlobal})
	}

	return candidates, nil
}

func findExecutables(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}
	var out []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fullPath := filepath.Join(dir, entry.Name())
		if !isExecutable(fullPath, entry) {
			continue
		}
		out = append(out, fullPath)
	}
	return out, nil
}

func isExecutable(path string, entry fs.DirEntry) bool {
	info, err := entry.Info()
	if err == nil {
		mode := info.Mode()
		if mode&0111 != 0 {
			return true
		}
	}
	// For windows
	if strings.HasSuffix(strings.ToLower(path), ".exe") {
		return true
	}
	return false
}

func configBaseDir() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir, nil
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}
	return configDir, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
