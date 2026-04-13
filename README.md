# Ralph

![Ralph](ralph.webp)

Ralph is an autonomous AI agent loop that runs AI coding tools ([Amp](https://ampcode.com), [Claude Code](https://docs.anthropic.com/en/docs/claude-code), or [Codex](https://developers.openai.com/codex/)) repeatedly until all PRD items are complete. Each iteration is a fresh instance with clean context. Memory persists via git history, `.ralph/progress.txt`, and `.ralph/prd.json`.

Based on [Geoffrey Huntley's Ralph pattern](https://ghuntley.com/ralph/).

[Read my in-depth article on how I use Ralph](https://x.com/ryancarson/status/2008548371712135632)

## Prerequisites

- One of the following AI coding tools installed and authenticated:
  - [Codex CLI](https://developers.openai.com/codex/) (default)
  - [Amp CLI](https://ampcode.com)
  - [Claude Code](https://docs.anthropic.com/en/docs/claude-code) (`npm install -g @anthropic-ai/claude-code`)
- `jq` installed (`brew install jq` on macOS)
- A git repository for your project

## Go CLI Preview

This repository now includes an in-progress Go implementation of Ralph:

```bash
go run ./cmd/ralph init
go run ./cmd/ralph doctor
go run ./cmd/ralph run
```

`go run ./cmd/ralph init` does not require a path. If you omit it, Ralph initializes `.ralph/` in the current git repository.

Current scope:

- default tool: Codex
- default run mode: 10 iterations
- `.ralph/` layout

The shell workflow remains the stable default while the Go CLI catches up.

## Setup

### Option 1: One-command init

Initialize Ralph in any git repository with one command:

```bash
/path/to/ralph/init-ralph.sh /path/to/your-repo
```

By default this creates:

- `.ralph/prd.json`
- `.ralph/progress.txt`
- `.ralph/tasks/*.md`
- `.ralph/ralph.sh`
- `.ralph/CODEX.md`
- `.ralph/CLAUDE.md`
- `.ralph/prompt.md`

The init command defaults to Codex. You can pick a different preferred tool or overwrite existing files:

```bash
/path/to/ralph/init-ralph.sh /path/to/your-repo --tool claude
/path/to/ralph/init-ralph.sh /path/to/your-repo --force
```

### Option 2: Copy to your project manually

Copy the Ralph files into your project yourself:

```bash
mkdir -p .ralph/tasks
cp /path/to/ralph/ralph.sh .ralph/
cp /path/to/ralph/prd.json.example .ralph/prd.json
cp /path/to/ralph/CODEX.md .ralph/CODEX.md
cp /path/to/ralph/CLAUDE.md .ralph/CLAUDE.md
cp /path/to/ralph/prompt.md .ralph/prompt.md
cp tasks/prd-[feature-name].md .ralph/tasks/
chmod +x .ralph/ralph.sh
```

### Option 3: Install skills globally (Amp)

Copy the skills to your Amp or Claude config for use across all projects:

For AMP
```bash
cp -r skills/prd ~/.config/amp/skills/
cp -r skills/ralph ~/.config/amp/skills/
```

For Claude Code (manual)
```bash
cp -r skills/prd ~/.claude/skills/
cp -r skills/ralph ~/.claude/skills/
```

### Option 4: Use as Claude Code Marketplace

Add the Ralph marketplace to Claude Code:

```bash
/plugin marketplace add snarktank/ralph
```

Then install the skills:

```bash
/plugin install ralph-skills@ralph-marketplace
```

Available skills after installation:
- `/prd` - Generate Product Requirements Documents
- `/ralph` - Convert PRDs to prd.json format

Skills are automatically invoked when you ask Claude to:
- "create a prd", "write prd for", "plan this feature"
- "convert this prd", "turn into ralph format", "create prd.json"

### Configure Amp auto-handoff (recommended)

Add to `~/.config/amp/settings.json`:

```json
{
  "amp.experimental.autoHandoff": { "context": 90 }
}
```

This enables automatic handoff when context fills up, allowing Ralph to handle large stories that exceed a single context window.

## Workflow

### 1. Create a PRD

Use the PRD skill to generate a detailed requirements document:

```
Load the prd skill and create a PRD for [your feature description]
```

Answer the clarifying questions. Save the output to `.ralph/tasks/prd-[feature-name].md`.

### 2. Convert PRD to Ralph format

Use the Ralph skill to convert the markdown PRD to JSON:

```
Load the ralph skill and convert .ralph/tasks/prd-[feature-name].md to .ralph/prd.json
```

This creates `.ralph/prd.json` with user stories structured for autonomous execution.

### 3. Run Ralph

```bash
# Using Codex (default)
./.ralph/ralph.sh [max_iterations]

# Using Amp
./.ralph/ralph.sh --tool amp [max_iterations]

# Using Claude Code
./.ralph/ralph.sh --tool claude [max_iterations]

# Using Codex explicitly
./.ralph/ralph.sh --tool codex [max_iterations]
```

Default is 10 iterations. Use `--tool amp`, `--tool claude`, or `--tool codex` to select your AI coding tool.
The Codex path runs with `--dangerously-bypass-approvals-and-sandbox`, so use it only in a trusted local repository.

Ralph will:
1. Create a feature branch (from PRD `branchName`)
2. Pick the highest priority story where `passes: false`
3. Implement that single story
4. Run quality checks (typecheck, tests)
5. Commit if checks pass, using `feat: US-001 Story title`
6. Update `.ralph/prd.json` to mark story as `passes: true`
7. Append learnings to `.ralph/progress.txt`
8. Repeat until all stories pass or max iterations reached

Before the first iteration, Ralph validates that:
- `.ralph/prd.json` exists and contains a branch name plus at least one story
- `.ralph/tasks/` exists and contains at least one markdown PRD source file
- The selected tool and prompt file are available

If those checks fail, Ralph exits immediately instead of starting a broken run.

## Key Files

| File | Purpose |
|------|---------|
| `ralph.sh` | The bash loop that spawns fresh AI instances (supports `--tool amp`, `--tool claude`, or `--tool codex`) |
| `prompt.md` | Prompt template for Amp |
| `CLAUDE.md` | Prompt template for Claude Code |
| `CODEX.md` | Prompt template for Codex |
| `.ralph/prd.json` | User stories with `passes` status (the task list) |
| `prd.json.example` | Example PRD format for reference |
| `.ralph/progress.txt` | Append-only learnings for future iterations |
| `.ralph/tasks/` | Source markdown PRDs for the current loop |
| `skills/prd/` | Skill for generating PRDs (works with Amp and Claude Code) |
| `skills/ralph/` | Skill for converting PRDs to JSON (works with Amp and Claude Code) |
| `.claude-plugin/` | Plugin manifest for Claude Code marketplace discovery |
| `flowchart/` | Interactive visualization of how Ralph works |

## Flowchart

[![Ralph Flowchart](ralph-flowchart.png)](https://snarktank.github.io/ralph/)

**[View Interactive Flowchart](https://snarktank.github.io/ralph/)** - Click through to see each step with animations.

The `flowchart/` directory contains the source code. To run locally:

```bash
cd flowchart
npm install
npm run dev
```

## Critical Concepts

### Each Iteration = Fresh Context

Each iteration spawns a **new AI instance** (Amp, Claude Code, or Codex) with clean context. The only memory between iterations is:
- Git history (commits from previous iterations)
- `.ralph/progress.txt` (learnings and context)
- `.ralph/prd.json` (which stories are done)

### Small Tasks

Each PRD item should be small enough to complete in one context window. If a task is too big, the LLM runs out of context before finishing and produces poor code.

Right-sized stories:
- Add a database column and migration
- Add a UI component to an existing page
- Update a server action with new logic
- Add a filter dropdown to a list

Too big (split these):
- "Build the entire dashboard"
- "Add authentication"
- "Refactor the API"

### AGENTS.md Updates Are Critical

After each iteration, Ralph updates the relevant `AGENTS.md` files with learnings. This is key because AI coding tools automatically read these files, so future iterations (and future human developers) benefit from discovered patterns, gotchas, and conventions.

Examples of what to add to AGENTS.md:
- Patterns discovered ("this codebase uses X for Y")
- Gotchas ("do not forget to update Z when changing W")
- Useful context ("the settings panel is in component X")

### Feedback Loops

Ralph only works if there are feedback loops:
- Typecheck catches type errors
- Tests verify behavior
- CI must stay green (broken code compounds across iterations)

### Browser Verification for UI Stories

Frontend stories must include "Verify in browser using dev-browser skill" in acceptance criteria. Ralph will use the dev-browser skill to navigate to the page, interact with the UI, and confirm changes work.

### Stop Condition

When all stories have `passes: true`, Ralph outputs `<promise>COMPLETE</promise>` and the loop exits.

## Debugging

Check current state:

```bash
# See which stories are done
cat .ralph/prd.json | jq '.userStories[] | {id, title, passes}'

# See learnings from previous iterations
cat .ralph/progress.txt

# Check git history
git log --oneline -10
```

## Customizing the Prompt

After copying `prompt.md` (for Amp), `CLAUDE.md` (for Claude Code), or `CODEX.md` (for Codex) to your project, customize it for your project:
- Add project-specific quality check commands
- Include codebase conventions
- Add common gotchas for your stack

## Archiving

Ralph automatically archives previous runs when you start a new feature (different `branchName`). Archives are saved to `archive/YYYY-MM-DD-feature-name/`.

## References

- [Geoffrey Huntley's Ralph article](https://ghuntley.com/ralph/)
- [Amp documentation](https://ampcode.com/manual)
- [Claude Code documentation](https://docs.anthropic.com/en/docs/claude-code)
- [Codex documentation](https://developers.openai.com/codex/)
