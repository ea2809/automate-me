package app

import (
	"os"

	"github.com/ea2809/automate-me/internal/core"
)

func resolveRepoRoot(cwd string) (string, error) {
	repoRoot, _, err := core.FindRepoRoot(cwd)
	if err != nil {
		return "", err
	}
	return repoRoot, nil
}

func resolveRepoAndTasks(cwd string) (string, []core.TaskRecord, error) {
	repoRoot, err := resolveRepoRoot(cwd)
	if err != nil {
		return "", nil, err
	}
	tasks, err := loadTasks(repoRoot)
	if err != nil {
		return "", nil, err
	}
	return repoRoot, tasks, nil
}

func currentRepoRoot() (string, error) {
	cwd, err := getwd()
	if err != nil {
		return "", err
	}
	return resolveRepoRoot(cwd)
}

func currentRepoAndTasks() (string, []core.TaskRecord, error) {
	cwd, err := getwd()
	if err != nil {
		return "", nil, err
	}
	return resolveRepoAndTasks(cwd)
}

func getwd() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return wd, nil
}
