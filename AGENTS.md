# Ralph Agent Instructions

## Overview

Ralph is an autonomous AI agent loop that runs AI coding tools (Amp, Claude Code, or Codex) repeatedly until all PRD items are complete. Each iteration is a fresh instance with clean context.

## Commands

```bash
# One-command init into another repository
./init-ralph.sh /path/to/target-repo

# Run the flowchart dev server
cd flowchart && npm run dev

# Build the flowchart
cd flowchart && npm run build

# Run Ralph with Codex (default)
./.ralph/ralph.sh [max_iterations]

# Run Ralph with Amp
./.ralph/ralph.sh --tool amp [max_iterations]

# Run Ralph with Claude Code
./.ralph/ralph.sh --tool claude [max_iterations]

# Run Ralph with Codex explicitly
./.ralph/ralph.sh --tool codex [max_iterations]
```

## Key Files

- `ralph.sh` - The bash loop that spawns fresh AI instances (supports `--tool amp`, `--tool claude`, or `--tool codex`)
- `init-ralph.sh` - One-command bootstrap script that creates `.ralph/` in a target repository
- `prompt.md` - Instructions given to each AMP instance
-  `CLAUDE.md` - Instructions given to each Claude Code instance
- `CODEX.md` - Instructions given to each Codex instance
- `prd.json.example` - Example PRD format
- `flowchart/` - Interactive React Flow diagram explaining how Ralph works

## Flowchart

The `flowchart/` directory contains an interactive visualization built with React Flow. It's designed for presentations - click through to reveal each step with animations.

To run locally:
```bash
cd flowchart
npm install
npm run dev
```

## Patterns

- Each iteration spawns a fresh AI instance (Amp, Claude Code, or Codex) with clean context
- Memory persists via git history, `.ralph/progress.txt`, and `.ralph/prd.json`
- Stories should be small enough to complete in one context window
- Commit messages must use exactly this format: `feat: US-001 Story title`
- Always update AGENTS.md with discovered patterns for future iterations
