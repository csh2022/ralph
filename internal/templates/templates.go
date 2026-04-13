package templates

const CodePrompt = `# Ralph Agent Instructions for Codex

You are an autonomous coding agent working on a software project.

## Your Task

1. Read the PRD at .ralph/prd.json
2. Read the progress log at .ralph/progress.txt (check Codebase Patterns section first)
3. Check you're on the correct branch from PRD branchName. If not, check it out or create from main.
4. Pick the highest priority user story where passes: false
5. Implement that single user story
6. Run quality checks (e.g. typecheck, lint, test - use whatever the project requires)
7. Update AGENTS.md files if you discover reusable patterns
8. If checks pass, commit all changes with message: feat: US-001 Story title
9. Update .ralph/prd.json to set passes: true for the completed story
10. Append progress to .ralph/progress.txt

## Progress Report Format

Append to .ralph/progress.txt (never replace, always append):

## [Date/Time] - [Story ID]
Session: [Codex session id if available]
- What was implemented
- Files changed
- Learnings for future iterations
---

The markdown source PRDs for this loop live under .ralph/tasks/.

## Quality Requirements

- All commits must pass project quality checks
- Do not commit broken code
- Keep changes focused and minimal
- Use one commit message format only: feat: US-001 Story title

## Stop Condition

After completing a user story, check if all stories have passes: true.

If all stories are complete and passing, reply with:
<promise>COMPLETE</promise>

If there are still stories with passes: false, end your response normally.
`

const TaskTemplate = `# Feature PRD

## Summary

Describe the feature in 2-3 sentences.

## Users

- Who is this for?
- What do they need?

## Requirements

- Requirement 1
- Requirement 2
- Requirement 3

## Constraints

- Technical constraint 1
- Validation requirement 1

## Notes

- Any rollout details, dependencies, or links
`

const PRDTemplate = `{
  "project": "Replace Me",
  "branchName": "ralph/replace-me",
  "description": "Replace this description with the feature you want Ralph to implement.",
  "userStories": [
    {
      "id": "US-001",
      "title": "Replace this placeholder story",
      "description": "As a user, I want you to replace this placeholder story so that Ralph has a real task list.",
      "acceptanceCriteria": [
        "Replace this placeholder story with real user stories",
        "Typecheck passes"
      ],
      "priority": 1,
      "passes": false,
      "notes": ""
    }
  ]
}
`
