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

const RalphPRDSkill = `---
name: ralph-prd
description: Generate Product Requirements Documents for Ralph workflows.
user-invocable: true
---

# Ralph PRD

Create a markdown PRD for the requested feature and save it under .ralph/tasks/.

Rules:
- Keep the PRD concise and implementation-aware.
- Prefer outcome-oriented requirements.
- Make the resulting PRD easy to convert into .ralph/prd.json.
- Save the file as .ralph/tasks/prd-[feature-name].md.
`

const PRDSessionPrompt = `You are helping the user create a Ralph feature PRD.

Goals:
- Follow the repository-local ralph-prd skill behavior if it is available.
- Ask clarifying questions until the feature is clear enough to implement.
- Keep the discussion focused on user outcomes, constraints, and validation.
- When the requirements are clear, write a markdown PRD to .ralph/tasks/prd-[feature-name].md.

Rules:
- Do not write .ralph/prd.json in this interactive session.
- Use a stable, descriptive file name under .ralph/tasks/ that starts with prd- and ends with .md.
- If you revise the plan during the conversation, update the markdown PRD file.
- Before the user exits, make sure the markdown PRD file exists on disk.
`

const PRDConvertPromptTemplate = `Convert the markdown PRD at %s into %s.

Requirements:
- Follow the repository-local ralph-prd-converter skill behavior if it is available.
- Read the markdown PRD and write valid Ralph JSON to %s.
- Keep stories small enough for one Ralph iteration each.
- Order stories by dependency and execution priority.
- Add "Typecheck passes" to every story.
- For UI work, add "Verify in browser using dev-browser skill".
- Preserve the feature intent from the markdown PRD.
- Do not modify unrelated files.
`

const RalphPRDConverterSkill = `---
name: ralph-prd-converter
description: Convert markdown PRDs into .ralph/prd.json for the Ralph autonomous agent loop.
user-invocable: true
---

# Ralph PRD Converter

Convert a markdown PRD into .ralph/prd.json.

Rules:
- Write valid JSON.
- Keep stories small enough for one Ralph iteration each.
- Order stories by dependency and execution priority.
- Add "Typecheck passes" to every story.
- For UI work, add "Verify in browser using dev-browser skill".
- Output must target .ralph/prd.json.
`
