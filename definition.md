Genial: Config 1 (Core en Go + plugins ejecutables con protocolo describe/run en JSON) es una base muy limpia para una “super herramienta” personal, extensible, con UX consistente y plugins por repo.

A continuación te dejo una definición completa (tipo mini‑especificación) para que cuando te pongas a implementarlo, no tengas que rediseñar sobre la marcha.

⸻

Devkit Config 1: Especificación (v1)

1) Objetivo y reglas

Objetivo

Tener un binario global (Go) llamado por ejemplo devkit que:
	1.	Descubre plugins (por repo + globales).
	2.	Pregunta a cada plugin qué tareas ofrece (describe).
	3.	Te muestra un selector único (misma UX siempre).
	4.	Ejecuta la tarea elegida delegando en el plugin (run).

Regla de oro para mantener UX consistente
	•	Los plugins no hacen UI (no menús, no prompts interactivos).
	•	Los plugins solo:
	•	describen tareas (metadata)
	•	ejecutan tareas con inputs ya resueltos

⸻

2) Terminología
	•	Core: el binario global (devkit) escrito en Go.
	•	Plugin: un ejecutable (binario Go, script bb, etc.) que implementa describe y run.
	•	Task: una acción ejecutable ofrecida por un plugin.
	•	Task ID: identificador único de una tarea (<pluginId>:<taskName>).

⸻

3) Estructura de carpetas estándar

En cada repo (local)

repo/
  .devkit/
    bin/
      devkit-repo        # plugin ejecutable (bb, go, etc.)
      devkit-docker

Global (usuario)
	•	~/.config/devkit/bin/  (o $XDG_CONFIG_HOME/devkit/bin/)

~/.config/devkit/
  bin/
    devkit-git
    devkit-personal

Nota: si no quieres depender de XDG, puedes fijarlo solo a ~/.config/devkit/bin y listo.

⸻

4) Descubrimiento de repo root

Repo root = la carpeta “raíz” donde tiene sentido cargar plugins locales.

Regla propuesta:
	1.	Si existe .devkit/ en algún padre → ese padre es repo root.
	2.	Si no, si existe .git/ en algún padre → ese padre es repo root.
	3.	Si no hay nada → “no estás en un repo”; solo plugins globales.

Esto te permite invocar devkit desde cualquier subcarpeta y que encuentre el repo.

⸻

5) Descubrimiento de plugins

Directorios a escanear (orden de prioridad)
	1.	${repoRoot}/.devkit/bin (si hay repoRoot)
	2.	~/.config/devkit/bin (global)

Qué cuenta como plugin
	•	Archivos ejecutables dentro de esos directorios.
	•	Recomendación de naming (opcional, para evitar falsos positivos):
	•	nombre empieza por devkit- (p.ej. devkit-docker)

MVP: puedes exigir el prefijo devkit- y ya. Más adelante, si quieres, permites “cualquier ejecutable” dentro de .devkit/bin.

Colisiones y precedencia
	•	Si dos plugins declaran el mismo plugin.id:
	•	por defecto: el local (repo) sobrescribe al global
	•	el core debe avisar (warning) y mostrar origen (local/global)
	•	Si dos tasks colisionan (pluginId:taskName igual) aplica la misma regla.

⸻

6) Protocolo del plugin (CLI)

Todo plugin debe soportar:

A) describe
	•	Comando:
pluginExecutable describe
	•	Salida: JSON por stdout (solo el manifest)
	•	Logs: a stderr
	•	Exit code:
	•	0 ok
	•	!=0 error (core lo reporta)

B) run
	•	Comando:
pluginExecutable run <taskName>
	•	Inputs: JSON por stdin (recomendado siempre; si no hay inputs, {})
	•	Salida:
	•	stdout/stderr: output normal del plugin (se muestra al usuario)
	•	(opcional futuro) JSON final por stdout si declaras “machineOutput”
	•	Exit code:
	•	0 ok
	•	!=0 el core considera fallo (propaga código)

⸻

7) Contrato de datos: Manifest JSON (schema v1)

Estructura mínima

{
  "schemaVersion": 1,
  "plugin": {
    "id": "repo",
    "title": "Repo tools",
    "version": "0.1.0"
  },
  "tasks": [
    {
      "name": "test",
      "title": "Run tests",
      "group": "QA",
      "description": "Ejecuta tests del repo",
      "inputs": [
        {
          "name": "pattern",
          "type": "string",
          "required": false,
          "prompt": "Regex/patrón (opcional)"
        }
      ]
    }
  ]
}

Campos recomendados

plugin
	•	id: string estable (sin espacios). Ej: repo, docker, git.
	•	title: nombre humano.
	•	version: opcional pero útil.

task
	•	name: string estable (único dentro del plugin) → se convierte a <plugin.id>:<name>
	•	title: texto para el menú
	•	group: para agrupar en UI (Docker/QA/Release…)
	•	description: ayuda corta
	•	inputs: lista de inputs (puede ser vacía)

⸻

8) Tipos de inputs (v1)

Para que el core pueda construir prompts uniformes, define un set pequeño pero útil:

InputSpec

{
  "name": "env",
  "type": "enum",
  "required": true,
  "prompt": "Entorno",
  "choices": ["dev", "staging", "prod"],
  "default": "dev"
}

Tipos propuestos (v1)
	•	string
	•	int
	•	float
	•	bool
	•	enum (con choices)
	•	path (string, pero el core puede validar existencia si quieres)
	•	multienum (lista, con choices)

Campos comunes
	•	name (clave)
	•	type
	•	required
	•	prompt
	•	default (opcional)
	•	choices (solo enum/multienum)
	•	secret (bool; si es true, el core no muestra el input en claro)

Mantén v1 pequeño. Más tipos = más carga para el core (validación/UI).

⸻

9) Contrato de ejecución: Input JSON para run

El core llama a run con JSON por stdin:

{
  "args": {
    "pattern": "Foo"
  },
  "ctx": {
    "repoRoot": "/abs/path/to/repo",
    "cwd": "/abs/path/to/repo/subdir",
    "selectedTaskId": "repo:test"
  }
}

ctx recomendado (v1)
	•	repoRoot: raíz del repo detectada
	•	cwd: donde se invocó devkit
	•	selectedTaskId: útil para logs

⸻

10) Variables de entorno estándar (para plugins)

Además del JSON por stdin, el core debe exportar envs (más fácil para scripts):
	•	DEVKIT_REPO_ROOT
	•	DEVKIT_CWD
	•	DEVKIT_TASK_ID (ej: repo:test)
	•	DEVKIT_PLUGIN_ID (ej: repo)
	•	DEVKIT_TASK_NAME (ej: test)

Y opcional:
	•	DEVKIT_SCOPE = repo|global (para debug)

⸻

11) Flujo del core (cómo se comporta devkit)

devkit (sin args)
	1.	Detecta repoRoot
	2.	Descubre plugins (local + global)
	3.	Para cada plugin:
	•	obtiene manifest vía describe (cacheable)
	4.	Construye lista de tasks (taskId = plugin.id:name)
	5.	Muestra menú único:
	•	búsqueda/filtro
	•	agrupación por group
	•	muestra title y description
	6.	Si la task tiene inputs:
	•	core pregunta inputs según inputs[]
	7.	Ejecuta:
	•	plugin run <taskName>
	•	stdin JSON con args+ctx
	•	reenvía stdout/stderr al usuario
	8.	Exit code del core = exit code del plugin

devkit run <taskId>
	•	salta el menú y ejecuta directamente (útil para CI o alias)

devkit list
	•	imprime lista de tasks disponibles (texto o JSON)
	•	útil para debug/scripting

devkit plugins
	•	muestra plugins detectados y desde dónde vienen (local/global + path)

⸻

12) Caché de manifests (muy recomendable)

Problema: llamar describe en 10 plugins cada vez puede sentirse lento.

Regla de cacheo (simple y robusta)
	•	Cache key = hash de:
	•	ruta del ejecutable
	•	mtime y size (del ejecutable)
	•	versión del core (para invalidar si cambia schema/interpretación)
	•	Cache storage:
	•	~/.cache/devkit/manifest-cache.json o un fichero por plugin

Comandos útiles
	•	devkit cache clear
	•	devkit cache stats (opcional)

Requisito contractual para plugins
	•	describe debe ser:
	•	rápido
	•	sin side-effects
	•	determinista (para que el caché sea fiable)

⸻

13) Buenas prácticas para plugins (bb y Go)

Plugins Babashka (Clojure)
	•	Shebang: #!/usr/bin/env bb
	•	describe imprime un mapa convertido a JSON (idealmente constante)
	•	run:
	•	parsea stdin JSON
	•	usa DEVKIT_REPO_ROOT para cd cuando corresponda
	•	ejecuta comandos con babashka.process

Plugins Go
	•	Binario pequeño
	•	describe devuelve JSON fijo (o ensamblado rápido)
	•	run hace lo que tenga que hacer

⸻

14) Extensiones futuras (sin romper v1)

Diseña desde ya con estas ideas en mente:
	•	schemaVersion en el manifest (ya lo tienes)
	•	Capabilities opcionales:
	•	plugin.capabilities: ["describe", "run", "doctor"]
	•	Hooks (futuro):
	•	preRun, postRun declarativos (o tasks especiales)
	•	Dependencias entre tasks (futuro):
	•	task.dependsOn: ["repo:build"]

Lo importante: v1 no lo necesita, pero dejar hueco evita rediseño.

⸻

Resumen: “cómo debería ser” en una frase

Un core Go que descubre ejecutables en .devkit/bin y ~/.config/devkit/bin, les pide un manifest JSON con describe, muestra un menú único, resuelve inputs y ejecuta con run pasando args+ctx por stdin y contexto por env; con caché de manifests para que sea instantáneo.

⸻

Si cuando vuelvas a esto quieres, te puedo dejar también:
	•	un schema JSON formal (para validar manifests),
	•	la tabla exacta de tipos/validaciones,
	•	y un diseño de UI con Bubble Tea (layout, filtros, groups, y cómo mostrar descripciones/origen).
