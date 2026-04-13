#!/bin/bash
# Ralph Wiggum - Long-running AI agent loop
# Usage: ./ralph.sh [--tool codex|amp|claude] [max_iterations]

set -e
set -o pipefail

# Parse arguments
TOOL="codex"  # Default to codex
MAX_ITERATIONS=10

while [[ $# -gt 0 ]]; do
  case $1 in
    --tool)
      TOOL="$2"
      shift 2
      ;;
    --tool=*)
      TOOL="${1#*=}"
      shift
      ;;
    *)
      # Assume it's max_iterations if it's a number
      if [[ "$1" =~ ^[0-9]+$ ]]; then
        MAX_ITERATIONS="$1"
      fi
      shift
      ;;
  esac
done

# Validate tool choice
if [[ "$TOOL" != "amp" && "$TOOL" != "claude" && "$TOOL" != "codex" ]]; then
  echo "Error: Invalid tool '$TOOL'. Must be 'amp', 'claude', or 'codex'."
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if PROJECT_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel 2>/dev/null)"; then
  :
else
  PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
fi
PRD_FILE="$SCRIPT_DIR/prd.json"
PROGRESS_FILE="$SCRIPT_DIR/progress.txt"
TASKS_DIR="$SCRIPT_DIR/tasks"
ARCHIVE_DIR="$SCRIPT_DIR/archive"
LAST_BRANCH_FILE="$SCRIPT_DIR/.last-branch"

validate_environment() {
  if ! command -v jq >/dev/null 2>&1; then
    echo "Error: jq is required but not installed."
    exit 1
  fi

  if [ ! -d "$TASKS_DIR" ]; then
    echo "Error: Ralph tasks directory not found at $TASKS_DIR"
    exit 1
  fi

  if ! find "$TASKS_DIR" -maxdepth 1 -type f \( -name '*.md' -o -name '*.markdown' \) | grep -q .; then
    echo "Error: No markdown PRD source files found in $TASKS_DIR"
    exit 1
  fi

  if [ ! -f "$PRD_FILE" ]; then
    echo "Error: PRD file not found at $PRD_FILE"
    exit 1
  fi

  if ! jq -e '.branchName and (.branchName | type == "string") and (.branchName | length > 0)' "$PRD_FILE" >/dev/null; then
    echo "Error: $PRD_FILE is missing a non-empty branchName."
    exit 1
  fi

  if ! jq -e '.userStories and (.userStories | type == "array") and (.userStories | length > 0)' "$PRD_FILE" >/dev/null; then
    echo "Error: $PRD_FILE must contain a non-empty userStories array."
    exit 1
  fi

  case "$TOOL" in
    amp)
      if [ ! -f "$SCRIPT_DIR/prompt.md" ]; then
        echo "Error: Amp prompt not found at $SCRIPT_DIR/prompt.md"
        exit 1
      fi
      if ! command -v amp >/dev/null 2>&1; then
        echo "Error: amp is required for --tool amp but was not found."
        exit 1
      fi
      ;;
    claude)
      if [ ! -f "$SCRIPT_DIR/CLAUDE.md" ]; then
        echo "Error: Claude prompt not found at $SCRIPT_DIR/CLAUDE.md"
        exit 1
      fi
      if ! command -v claude >/dev/null 2>&1; then
        echo "Error: claude is required for --tool claude but was not found."
        exit 1
      fi
      ;;
    codex)
      if [ ! -f "$SCRIPT_DIR/CODEX.md" ]; then
        echo "Error: Codex prompt not found at $SCRIPT_DIR/CODEX.md"
        exit 1
      fi
      if ! command -v codex >/dev/null 2>&1; then
        echo "Error: codex is required for --tool codex but was not found."
        exit 1
      fi
      ;;
  esac
}

validate_environment

if ! jq -e '.userStories[] | select(.passes == false)' "$PRD_FILE" >/dev/null; then
  echo "Ralph found no unfinished stories in $PRD_FILE. Exiting."
  exit 0
fi

# Archive previous run if branch changed
if [ -f "$PRD_FILE" ] && [ -f "$LAST_BRANCH_FILE" ]; then
  CURRENT_BRANCH=$(jq -r '.branchName // empty' "$PRD_FILE" 2>/dev/null || echo "")
  LAST_BRANCH=$(cat "$LAST_BRANCH_FILE" 2>/dev/null || echo "")
  
  if [ -n "$CURRENT_BRANCH" ] && [ -n "$LAST_BRANCH" ] && [ "$CURRENT_BRANCH" != "$LAST_BRANCH" ]; then
    # Archive the previous run
    DATE=$(date +%Y-%m-%d)
    # Strip "ralph/" prefix from branch name for folder
    FOLDER_NAME=$(echo "$LAST_BRANCH" | sed 's|^ralph/||')
    ARCHIVE_FOLDER="$ARCHIVE_DIR/$DATE-$FOLDER_NAME"
    
    echo "Archiving previous run: $LAST_BRANCH"
    mkdir -p "$ARCHIVE_FOLDER"
    [ -f "$PRD_FILE" ] && cp "$PRD_FILE" "$ARCHIVE_FOLDER/"
    [ -f "$PROGRESS_FILE" ] && cp "$PROGRESS_FILE" "$ARCHIVE_FOLDER/"
    echo "   Archived to: $ARCHIVE_FOLDER"
    
    # Reset progress file for new run
    printf '# Ralph Progress Log\nStarted: %s\n---\n' "$(date)" > "$PROGRESS_FILE"
  fi
fi

# Track current branch
if [ -f "$PRD_FILE" ]; then
  CURRENT_BRANCH=$(jq -r '.branchName // empty' "$PRD_FILE" 2>/dev/null || echo "")
  if [ -n "$CURRENT_BRANCH" ]; then
    echo "$CURRENT_BRANCH" > "$LAST_BRANCH_FILE"
  fi
fi

# Initialize progress file if it doesn't exist
if [ ! -f "$PROGRESS_FILE" ]; then
  printf '# Ralph Progress Log\nStarted: %s\n---\n' "$(date)" > "$PROGRESS_FILE"
fi

echo "Starting Ralph - Tool: $TOOL - Max iterations: $MAX_ITERATIONS"

for i in $(seq 1 $MAX_ITERATIONS); do
  echo ""
  echo "==============================================================="
  echo "  Ralph Iteration $i of $MAX_ITERATIONS ($TOOL)"
  echo "==============================================================="
  OUTPUT_FILE=$(mktemp "${TMPDIR:-/tmp}/ralph-output.XXXXXX")

  # Run the selected tool with the ralph prompt
  if [[ "$TOOL" == "amp" ]]; then
    cat "$SCRIPT_DIR/prompt.md" | amp --dangerously-allow-all 2>&1 | tee "$OUTPUT_FILE" || true
  elif [[ "$TOOL" == "claude" ]]; then
    # Claude Code: use --dangerously-skip-permissions for autonomous operation, --print for output
    claude --dangerously-skip-permissions --print < "$SCRIPT_DIR/CLAUDE.md" 2>&1 | tee "$OUTPUT_FILE" || true
  else
    # Codex CLI: run from the repository root so repo-relative paths in the PRD and prompts resolve consistently
    codex exec --dangerously-bypass-approvals-and-sandbox -C "$PROJECT_ROOT" - < "$SCRIPT_DIR/CODEX.md" 2>&1 | tee "$OUTPUT_FILE" || true
  fi
  
  # Check for completion signal
  if grep -q "<promise>COMPLETE</promise>" "$OUTPUT_FILE"; then
    if jq -e '.userStories[] | select(.passes == false)' "$PRD_FILE" >/dev/null; then
      echo "Completion signal received, but unfinished stories remain in $PRD_FILE. Continuing..."
      rm -f "$OUTPUT_FILE"
      echo "Iteration $i complete. Continuing..."
      sleep 2
      continue
    fi
    rm -f "$OUTPUT_FILE"
    echo ""
    echo "Ralph completed all tasks!"
    echo "Completed at iteration $i of $MAX_ITERATIONS"
    exit 0
  fi

  rm -f "$OUTPUT_FILE"
  
  echo "Iteration $i complete. Continuing..."
  sleep 2
done

echo ""
echo "Ralph reached max iterations ($MAX_ITERATIONS) without completing all tasks."
echo "Check $PROGRESS_FILE for status."
exit 1
