# Ralph

![Ralph](ralph.webp)

Ralph is a Go CLI for running an autonomous Codex-driven PRD-to-implementation loop.

Current top-level focus:
- Go implementation under `cmd/` and `internal/`
- Codex-first workflow
- `.ralph/` project state directory
- interactive `ralph prd` flow
- iterative `ralph run` execution

Legacy shell, Amp, Claude Code, marketplace, and flowchart assets were moved to:
- `backup/legacy-shell/`

## Prerequisites

- [Codex CLI](https://developers.openai.com/codex/) installed and authenticated
- a git repository for the target project

## Build

```bash
./build.sh
```

Build output:

```bash
output/ralph
```

Or run directly with Go:

```bash
go run ./cmd/ralph help
```

## Commands

```bash
ralph init [path] [--force]
ralph doctor
ralph prd [--convert-only .ralph/tasks/prd-example.md]
ralph run [max-iterations] [--tool codex] [--until-done]
```

### `ralph init`

Initializes the target git repository with:

- `.ralph/prd.json`
- `.ralph/progress.txt`
- `.ralph/CODEX.md`
- `.ralph/tasks/prd-template.md`
- `.codex/skills/ralph-prd/SKILL.md`
- `.codex/skills/ralph-prd-converter/SKILL.md`

If no path is provided, initialization uses the current directory.

Examples:

```bash
ralph init
ralph init /path/to/repo
ralph init --force
```

### `ralph doctor`

Validates the current repository and Ralph setup:

- git repository detected
- Codex installed
- `.ralph/` files present
- repo-local Codex skills present
- `prd.json` parses and contains work items

### `ralph prd`

Starts an interactive Codex session to define a feature PRD.

Behavior:
- prints a high-visibility banner before starting
- tells the user how to exit (`Ctrl-D` or `exit`)
- if Codex shows a trust prompt, the user should choose `Yes, continue`
- expects Codex to write `.ralph/tasks/prd-[feature-name].md`
- after Codex exits, automatically converts the latest changed PRD markdown into `.ralph/prd.json`

Direct conversion is also supported:

```bash
ralph prd --convert-only .ralph/tasks/prd-my-feature.md
```

### `ralph run`

Runs the implementation loop.

Default behavior:
- tool: `codex`
- max iterations: `10`
- optional unlimited mode: `--until-done`

Flow:
1. validate the repository
2. load `.ralph/prd.json`
3. pick the highest-priority unfinished story
4. run Codex non-interactively
5. stream output live
6. re-check `.ralph/prd.json`
7. stop when all stories are complete, max iterations is reached, or the no-progress circuit breaker trips

Example:

```bash
ralph run
ralph run 1
ralph run 20
ralph run --until-done
```

## Repository layout

Top-level active code:

- `cmd/ralph/` - CLI entrypoint
- `internal/app/` - command dispatch
- `internal/doctor/` - repo validation
- `internal/project/` - `.ralph/` state and file management
- `internal/runner/` - Codex execution and PRD flow
- `internal/templates/` - embedded prompt and skill templates
- `build.sh` - local build script

Legacy backup:

- `backup/legacy-shell/` - previous shell runner, prompts, skills, marketplace files, templates, and flowchart assets

## Notes

- commit messages for autonomous story work should use: `feat: US-001 Story title`
- repo-local Codex skills are intentionally installed under `.codex/skills/`
- the interactive `ralph prd` flow converts the latest changed `.ralph/tasks/prd-*.md` after Codex exits
