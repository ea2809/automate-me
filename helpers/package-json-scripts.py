#!/usr/bin/env python3
import json
import os
import shutil
import subprocess
import sys
from pathlib import Path

PLUGIN_ID = "package"
PLUGIN_TITLE = "Package Scripts"


def find_package_json(base: Path) -> Path | None:
    current = base
    for candidate in [current, *current.parents]:
        pkg = candidate / "package.json"
        if pkg.is_file():
            return pkg
    return None


def load_scripts(pkg_path: Path) -> dict[str, str]:
    data = json.loads(pkg_path.read_text(encoding="utf-8"))
    scripts = data.get("scripts", {})
    if not isinstance(scripts, dict):
        return {}
    out: dict[str, str] = {}
    for name, command in scripts.items():
        if isinstance(name, str) and name and isinstance(command, str):
            out[name] = command
    return out


def describe() -> int:
    cwd = Path.cwd()
    pkg_path = find_package_json(cwd)
    tasks = []

    if pkg_path:
        scripts = load_scripts(pkg_path)
        for name, command in sorted(scripts.items()):
            tasks.append(
                {
                    "name": name,
                    "title": f"Run {name}",
                    "group": "Node",
                    "description": command,
                    "inputs": [],
                }
            )

    manifest = {
        "schemaVersion": 1,
        "plugin": {
            "id": PLUGIN_ID,
            "title": PLUGIN_TITLE,
            "version": "0.1.0",
        },
        "tasks": tasks,
    }
    print(json.dumps(manifest))
    return 0


def pick_runner(workdir: Path) -> list[str]:
    if (workdir / "pnpm-lock.yaml").is_file() and shutil.which("pnpm"):
        return ["pnpm", "run"]
    if (workdir / "yarn.lock").is_file() and shutil.which("yarn"):
        return ["yarn", "run"]
    if shutil.which("npm"):
        return ["npm", "run"]
    if shutil.which("yarn"):
        return ["yarn", "run"]
    if shutil.which("pnpm"):
        return ["pnpm", "run"]
    print("no package manager found (npm/yarn/pnpm)", file=sys.stderr)
    sys.exit(1)


def read_payload() -> dict:
    raw = sys.stdin.read().strip()
    if not raw:
        return {}
    try:
        data = json.loads(raw)
    except json.JSONDecodeError as exc:
        print(f"invalid input JSON: {exc}", file=sys.stderr)
        sys.exit(1)
    if not isinstance(data, dict):
        print("input JSON must be an object", file=sys.stderr)
        sys.exit(1)
    return data


def run(task_name: str) -> int:
    payload = read_payload()
    ctx = payload.get("ctx", {})
    if not isinstance(ctx, dict):
        ctx = {}

    cwd = ctx.get("cwd") or os.getenv("AUTOMATE_ME_CWD") or os.getcwd()
    workdir = Path(cwd)
    pkg_path = find_package_json(workdir)
    if not pkg_path:
        print("package.json not found from current context", file=sys.stderr)
        return 1

    scripts = load_scripts(pkg_path)
    if task_name not in scripts:
        print(f"unknown script: {task_name}", file=sys.stderr)
        return 1

    runner = pick_runner(pkg_path.parent)
    cmd = [*runner, task_name]
    proc = subprocess.run(cmd, cwd=str(pkg_path.parent))
    return proc.returncode


def usage() -> int:
    print("usage: package-json-scripts describe|run <task>", file=sys.stderr)
    return 1


def main(argv: list[str]) -> int:
    if len(argv) < 2:
        return usage()
    cmd = argv[1]
    if cmd == "describe":
        return describe()
    if cmd == "run":
        if len(argv) < 3:
            return usage()
        return run(argv[2])
    return usage()


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
