# automate-me

`automate-me` is a Go TUI/CLI for running automation tasks from local and global plugins. It discovers plugins, lists their tasks, lets you pick one in a consistent TUI, and executes it with a small JSON protocol and well-defined environment.

Disclaimer: this is an alpha project. Expect breaking changes and rough edges. More features and tooling will be added over time.

## Features
- Interactive TUI task picker with filtering.
- Run tasks by id (`plugin:task`) from the CLI.
- Local and global plugin discovery.
- JSON spec import for simple, direct-exec tasks.
- Protocol-based plugins for richer behavior.

## Install

```bash
go install github.com/ea2809/automate-me/cmd/automate-me@latest
```

Or build from source:

```bash
go build ./cmd/automate-me
```

## Usage

```bash
automate-me            # interactive TUI
automate-me list       # list tasks
automate-me plugins    # list discovered plugins
automate-me run repo:test

automate-me import path/to/spec.json         # import to local spec dir if in a repo
automate-me import path/to/spec.json --local # force local
automate-me import path/to/spec.json --global
```

## Plugin Discovery

`automate-me` searches for executable plugins in these locations (local first):
- Local repo: `.automate-me/bin`
- Global config: `$XDG_CONFIG_HOME/automate-me/bin` (or your OS config dir)

If you run `automate-me` inside a repo, the repo root is the nearest parent containing `.automate-me/`, otherwise it falls back to the nearest `.git/`.

## Spec Import (Direct Exec)

Specs are JSON manifests that define tasks. When `execMode` is omitted or set to `direct`, the command in `plugin.exec` is run directly (no `describe`/`run` subcommands).

Example: `sample-specs/ls.json`

```json
{
  "schemaVersion": 1,
  "plugin": {
    "id": "sample",
    "title": "Sample",
    "version": "0.1.0",
    "exec": "/bin/ls",
    "execMode": "direct"
  },
  "tasks": [
    {
      "name": "repo",
      "title": "List repo files",
      "group": "Sample",
      "description": "Runs ls in the repo root",
      "inputs": []
    }
  ]
}
```

Import it:

```bash
automate-me import sample-specs/ls.json --local
```

Specs are stored at:
- Local: `.automate-me/specs`
- Global: `$XDG_CONFIG_HOME/automate-me/specs` (or your OS config dir)

## Protocol Plugins

Executable plugins can provide tasks dynamically via a simple protocol:

- Describe
  - Command: `<plugin> describe`
  - Output: manifest JSON to stdout
  - Logs: stderr

- Run
  - Command: `<plugin> run <taskName>`
  - Input: JSON on stdin

Input JSON shape:

```json
{
  "args": {"key": "value"},
  "ctx": {
    "repoRoot": "/path/to/repo",
    "cwd": "/path/to/repo/subdir",
    "selectedTaskId": "plugin:task"
  }
}
```

Environment variables provided to all tasks:
- `AUTOMATE_ME_REPO_ROOT`
- `AUTOMATE_ME_CWD`
- `AUTOMATE_ME_TASK_ID`
- `AUTOMATE_ME_PLUGIN_ID`
- `AUTOMATE_ME_TASK_NAME`
- `AUTOMATE_ME_SCOPE`

If a spec sets `plugin.execMode` to `protocol`, `automate-me` will run the plugin with the `run` subcommand.

## Examples

Two minimal protocol plugin examples (sanitized):
- Bash + Python: `examples/automate-me-projects`
- Clojure (Babashka): `examples/gitstatus.clj`
- Python: `examples/automate-me-python`

## Helpers

Reusable helper plugins live in `helpers/`.

- `helpers/package-json-scripts`: Python protocol plugin that reads a real `package.json`, publishes each entry in `scripts` as a task, and runs it with `npm`, `yarn`, or `pnpm` (auto-detected).
- `helpers/package-json-scripts.js`: Node.js protocol plugin with the same behavior (no Python required).

Quick setup (Node version):

```bash
mkdir -p .automate-me/bin
cp helpers/package-json-scripts.js .automate-me/bin/package-json-scripts
chmod +x .automate-me/bin/package-json-scripts
automate-me list
```

## Development

```bash
go test ./...
```

## License

MIT. See `LICENSE`.
