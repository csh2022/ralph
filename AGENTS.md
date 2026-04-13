# Ralph Agent Instructions

## Overview

Ralph is now a Go-first Codex CLI project. The active implementation lives under `cmd/` and `internal/`.

Legacy shell/bootstrap/marketplace/flowchart assets were moved to:
- `backup/legacy-shell/`

## Commands

```bash
# Build the CLI
./build.sh

# Run the CLI directly with Go
go run ./cmd/ralph help

# Initialize Ralph in a target repository
./output/ralph init [path] [--force]

# Validate the current repository
./output/ralph doctor

# Start the interactive PRD flow
./output/ralph prd

# Convert a markdown PRD directly
./output/ralph prd --convert-only .ralph/tasks/prd-example.md

# Run the autonomous loop
./output/ralph run [max_iterations]
```

## Active Files

- `cmd/ralph/` - CLI entrypoint
- `internal/app/` - command handling
- `internal/doctor/` - validation checks
- `internal/project/` - `.ralph/` and `.codex/skills/` state management
- `internal/runner/` - Codex execution paths
- `internal/templates/` - embedded prompts and skill templates
- `build.sh` - builds `output/ralph`

## Patterns

- The active product surface is the Go CLI, not the legacy shell runner.
- Repository-local Codex skills live under `.codex/skills/ralph-prd` and `.codex/skills/ralph-prd-converter`.
- The interactive `ralph prd` flow converts the latest changed `.ralph/tasks/prd-*.md` after Codex exits.
- Commit messages for autonomous story work must use exactly this format: `feat: US-001 Story title`.
- Keep legacy compatibility assets under `backup/legacy-shell/` instead of mixing them into the top-level product surface.
- Always update AGENTS.md with genuinely reusable patterns discovered while changing the active Go implementation.
