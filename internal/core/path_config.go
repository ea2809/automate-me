package core

import (
	"fmt"
	"path/filepath"
)

const (
	localConfigDirName  = ".automate-me"
	globalConfigDirName = "automate-me"
	gitDirName          = ".git"
	binDirName          = "bin"
	specsDirName        = "specs"
)

type pathConfig struct {
	repoRoot string
}

func newPathConfig(repoRoot string) pathConfig {
	return pathConfig{repoRoot: repoRoot}
}

func (p pathConfig) localRoot() (string, error) {
	if p.repoRoot == "" {
		return "", fmt.Errorf("no repo root for local path")
	}
	return filepath.Join(p.repoRoot, localConfigDirName), nil
}

func (p pathConfig) localBin() (string, error) {
	root, err := p.localRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, binDirName), nil
}

func (p pathConfig) localSpecs() (string, error) {
	root, err := p.localRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, specsDirName), nil
}

func (p pathConfig) globalRoot() (string, error) {
	configDir, err := configBaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, globalConfigDirName), nil
}

func (p pathConfig) globalBin() (string, error) {
	root, err := p.globalRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, binDirName), nil
}

func (p pathConfig) globalSpecs() (string, error) {
	root, err := p.globalRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, specsDirName), nil
}
