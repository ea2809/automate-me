#!/usr/bin/env node

const fs = require('node:fs');
const path = require('node:path');
const { spawnSync } = require('node:child_process');

const PLUGIN_ID = 'package';
const PLUGIN_TITLE = 'Package Scripts';

function findPackageJson(startDir) {
  let current = path.resolve(startDir);
  while (true) {
    const pkg = path.join(current, 'package.json');
    if (fs.existsSync(pkg) && fs.statSync(pkg).isFile()) {
      return pkg;
    }
    const parent = path.dirname(current);
    if (parent === current) {
      return null;
    }
    current = parent;
  }
}

function loadScripts(pkgPath) {
  let data;
  try {
    data = JSON.parse(fs.readFileSync(pkgPath, 'utf8'));
  } catch (err) {
    console.error(`invalid package.json: ${err.message}`);
    process.exit(1);
  }

  if (!data || typeof data !== 'object' || !data.scripts || typeof data.scripts !== 'object') {
    return {};
  }

  const scripts = {};
  for (const [name, cmd] of Object.entries(data.scripts)) {
    if (typeof name === 'string' && name && typeof cmd === 'string') {
      scripts[name] = cmd;
    }
  }
  return scripts;
}

function describe() {
  const pkgPath = findPackageJson(process.cwd());
  const tasks = [];

  if (pkgPath) {
    const scripts = loadScripts(pkgPath);
    for (const name of Object.keys(scripts).sort()) {
      tasks.push({
        name,
        title: `Run ${name}`,
        group: 'Node',
        description: scripts[name],
        inputs: [],
      });
    }
  }

  const manifest = {
    schemaVersion: 1,
    plugin: {
      id: PLUGIN_ID,
      title: PLUGIN_TITLE,
      version: '0.1.0',
    },
    tasks,
  };

  process.stdout.write(`${JSON.stringify(manifest)}\n`);
}

function hasCommand(cmd) {
  const checker = process.platform === 'win32' ? 'where' : 'command';
  const args = process.platform === 'win32' ? [cmd] : ['-v', cmd];
  const res = spawnSync(checker, args, { stdio: 'ignore', shell: process.platform !== 'win32' });
  return res.status === 0;
}

function pickRunner(workdir) {
  if (fs.existsSync(path.join(workdir, 'pnpm-lock.yaml')) && hasCommand('pnpm')) {
    return ['pnpm', 'run'];
  }
  if (fs.existsSync(path.join(workdir, 'yarn.lock')) && hasCommand('yarn')) {
    return ['yarn', 'run'];
  }
  if (hasCommand('npm')) {
    return ['npm', 'run'];
  }
  if (hasCommand('yarn')) {
    return ['yarn', 'run'];
  }
  if (hasCommand('pnpm')) {
    return ['pnpm', 'run'];
  }
  console.error('no package manager found (npm/yarn/pnpm)');
  process.exit(1);
}

function readPayload() {
  const raw = fs.readFileSync(0, 'utf8').trim();
  if (!raw) {
    return {};
  }
  try {
    const data = JSON.parse(raw);
    if (data && typeof data === 'object') {
      return data;
    }
    console.error('input JSON must be an object');
    process.exit(1);
  } catch (err) {
    console.error(`invalid input JSON: ${err.message}`);
    process.exit(1);
  }
}

function run(taskName) {
  const payload = readPayload();
  const ctx = payload && typeof payload.ctx === 'object' ? payload.ctx : {};
  const cwd = ctx.cwd || process.env.AUTOMATE_ME_CWD || process.cwd();

  const pkgPath = findPackageJson(cwd);
  if (!pkgPath) {
    console.error('package.json not found from current context');
    process.exit(1);
  }

  const scripts = loadScripts(pkgPath);
  if (!Object.prototype.hasOwnProperty.call(scripts, taskName)) {
    console.error(`unknown script: ${taskName}`);
    process.exit(1);
  }

  const workdir = path.dirname(pkgPath);
  const runner = pickRunner(workdir);
  const result = spawnSync(runner[0], [...runner.slice(1), taskName], {
    cwd: workdir,
    stdio: 'inherit',
  });

  if (result.error) {
    console.error(result.error.message);
    process.exit(1);
  }
  process.exit(result.status || 0);
}

function usage() {
  console.error('usage: package-json-scripts.js describe|run <task>');
  process.exit(1);
}

function main(argv) {
  if (argv.length < 3) {
    usage();
  }
  const cmd = argv[2];
  if (cmd === 'describe') {
    describe();
    return;
  }
  if (cmd === 'run') {
    if (argv.length < 4) {
      usage();
    }
    run(argv[3]);
    return;
  }
  usage();
}

main(process.argv);
