#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_DIR="${1:-.}"
TOOL="codex"
FORCE=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --tool)
      TOOL="$2"
      shift 2
      ;;
    --tool=*)
      TOOL="${1#*=}"
      shift
      ;;
    --force)
      FORCE=1
      shift
      ;;
    *)
      TARGET_DIR="$1"
      shift
      ;;
  esac
done

if [[ "$TOOL" != "codex" && "$TOOL" != "amp" && "$TOOL" != "claude" ]]; then
  echo "Error: Invalid tool '$TOOL'. Must be 'codex', 'amp', or 'claude'."
  exit 1
fi

TARGET_DIR="$(cd "$TARGET_DIR" && pwd)"
RALPH_DIR="$TARGET_DIR/.ralph"
TASKS_DIR="$RALPH_DIR/tasks"

if ! git -C "$TARGET_DIR" rev-parse --show-toplevel >/dev/null 2>&1; then
  echo "Error: $TARGET_DIR is not inside a git repository."
  exit 1
fi

mkdir -p "$TASKS_DIR"

copy_file() {
  local src="$1"
  local dest="$2"

  if [[ -e "$dest" && "$FORCE" -ne 1 ]]; then
    echo "Skip existing file: $dest"
    return
  fi

  cp "$src" "$dest"
  echo "Wrote $dest"
}

copy_file "$SCRIPT_DIR/ralph.sh" "$RALPH_DIR/ralph.sh"
copy_file "$SCRIPT_DIR/CODEX.md" "$RALPH_DIR/CODEX.md"
copy_file "$SCRIPT_DIR/CLAUDE.md" "$RALPH_DIR/CLAUDE.md"
copy_file "$SCRIPT_DIR/prompt.md" "$RALPH_DIR/prompt.md"
copy_file "$SCRIPT_DIR/templates/prd.template.json" "$RALPH_DIR/prd.json"
copy_file "$SCRIPT_DIR/templates/prd-template.md" "$TASKS_DIR/prd-template.md"

chmod +x "$RALPH_DIR/ralph.sh"

if [[ ! -f "$RALPH_DIR/progress.txt" || "$FORCE" -eq 1 ]]; then
  printf '# Ralph Progress Log\nStarted: %s\n---\n' "$(date)" > "$RALPH_DIR/progress.txt"
  echo "Wrote $RALPH_DIR/progress.txt"
else
  echo "Skip existing file: $RALPH_DIR/progress.txt"
fi

cat <<EOF

Ralph initialized in:
  $RALPH_DIR

Next steps:
  1. Edit .ralph/tasks/prd-template.md or replace it with your feature PRD.
  2. Convert that PRD into .ralph/prd.json.
  3. Run ./.ralph/ralph.sh

Files created:
  - .ralph/ralph.sh
  - .ralph/prd.json
  - .ralph/progress.txt
  - .ralph/tasks/prd-template.md
  - .ralph/CODEX.md
  - .ralph/CLAUDE.md
  - .ralph/prompt.md
EOF
