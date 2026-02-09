package app

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/ea2809/automate-me/internal/core"
)

func importSpec(args []string) error {
	fs := flag.NewFlagSet("import", flag.ContinueOnError)
	var useGlobal bool
	var useLocal bool
	fs.BoolVar(&useGlobal, "global", false, "store in global spec dir")
	fs.BoolVar(&useLocal, "local", false, "store in repo spec dir")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if useGlobal && useLocal {
		return errors.New("use only one of --global or --local")
	}
	if fs.NArg() < 1 {
		return errors.New("usage: automate-me import <spec.json> [--local|--global]")
	}

	path := fs.Arg(0)
	repoRoot, _, err := core.FindRepoRoot(mustGetwd())
	if err != nil {
		return err
	}
	scope := core.ScopeGlobal
	if useLocal || (!useGlobal && repoRoot != "") {
		scope = core.ScopeLocal
	}
	storedPath, err := core.ImportSpecFile(path, repoRoot, scope)
	if err != nil {
		return err
	}
	fmt.Printf("imported %s -> %s\n", filepath.Base(path), storedPath)
	return nil
}
