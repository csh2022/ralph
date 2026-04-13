package project

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/snarktank/ralph/internal/templates"
)

type PRD struct {
	Project     string      `json:"project"`
	BranchName  string      `json:"branchName"`
	Description string      `json:"description"`
	UserStories []UserStory `json:"userStories"`
}

type UserStory struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	AcceptanceCriteria []string `json:"acceptanceCriteria"`
	Priority           int      `json:"priority"`
	Passes             bool     `json:"passes"`
	Notes              string   `json:"notes"`
}

type Paths struct {
	RepoRoot                   string
	RalphDir                   string
	PRDFile                    string
	ProgressFile               string
	TaskDir                    string
	CodePromptFile             string
	LastBranchFile             string
	ArchiveDir                 string
	CodexDir                   string
	CodexSkillsDir             string
	RalphPRDSkillFile          string
	RalphPRDConverterSkillFile string
}

func Discover(repoHint string) (Paths, error) {
	root, err := gitRoot(repoHint)
	if err != nil {
		return Paths{}, err
	}
	ralphDir := filepath.Join(root, ".ralph")
	return Paths{
		RepoRoot:                   root,
		RalphDir:                   ralphDir,
		PRDFile:                    filepath.Join(ralphDir, "prd.json"),
		ProgressFile:               filepath.Join(ralphDir, "progress.txt"),
		TaskDir:                    filepath.Join(ralphDir, "tasks"),
		CodePromptFile:             filepath.Join(ralphDir, "CODEX.md"),
		LastBranchFile:             filepath.Join(ralphDir, ".last-branch"),
		ArchiveDir:                 filepath.Join(ralphDir, "archive"),
		CodexDir:                   filepath.Join(root, ".codex"),
		CodexSkillsDir:             filepath.Join(root, ".codex", "skills"),
		RalphPRDSkillFile:          filepath.Join(root, ".codex", "skills", "ralph-prd", "SKILL.md"),
		RalphPRDConverterSkillFile: filepath.Join(root, ".codex", "skills", "ralph-prd-converter", "SKILL.md"),
	}, nil
}

func gitRoot(dir string) (string, error) {
	target := dir
	if target == "" {
		var err error
		target, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	cmd := exec.Command("git", "-C", target, "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not inside a git repository: %w", err)
	}
	return filepath.Clean(string(bytes.TrimSpace(output))), nil
}

func EnsureInitialized(paths Paths, force bool) error {
	if err := os.MkdirAll(paths.TaskDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(paths.ArchiveDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(paths.RalphPRDSkillFile), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(paths.RalphPRDConverterSkillFile), 0o755); err != nil {
		return err
	}

	if err := writeFile(paths.PRDFile, []byte(templates.PRDTemplate), force); err != nil {
		return err
	}
	if err := writeFile(paths.CodePromptFile, []byte(templates.CodePrompt), force); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(paths.TaskDir, "prd-template.md"), []byte(templates.TaskTemplate), force); err != nil {
		return err
	}
	if err := writeFile(paths.RalphPRDSkillFile, []byte(templates.RalphPRDSkill), force); err != nil {
		return err
	}
	if err := writeFile(paths.RalphPRDConverterSkillFile, []byte(templates.RalphPRDConverterSkill), force); err != nil {
		return err
	}
	if err := ensureProgress(paths.ProgressFile, force); err != nil {
		return err
	}
	if err := ensureLastBranch(paths.LastBranchFile, force); err != nil {
		return err
	}
	return nil
}

func ensureProgress(path string, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}
	content := fmt.Sprintf("# Ralph Progress Log\nStarted: %s\n---\n", time.Now().Format(time.RFC3339))
	return os.WriteFile(path, []byte(content), 0o644)
}

func ensureLastBranch(path string, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}
	return os.WriteFile(path, []byte(""), 0o644)
}

func writeFile(path string, content []byte, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}
	return os.WriteFile(path, content, 0o644)
}

func LoadPRD(path string) (PRD, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PRD{}, err
	}
	var prd PRD
	if err := json.Unmarshal(data, &prd); err != nil {
		return PRD{}, err
	}
	return prd, nil
}

func HasUnfinishedStories(prd PRD) bool {
	for _, story := range prd.UserStories {
		if !story.Passes {
			return true
		}
	}
	return false
}

func Validate(prd PRD) error {
	if prd.BranchName == "" {
		return errors.New("prd.json missing branchName")
	}
	if len(prd.UserStories) == 0 {
		return errors.New("prd.json must contain at least one user story")
	}
	return nil
}

func ArchiveIfBranchChanged(paths Paths, prd PRD) error {
	lastBranch, err := os.ReadFile(paths.LastBranchFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	current := prd.BranchName
	previous := string(bytes.TrimSpace(lastBranch))
	if previous == "" || previous == current {
		return os.WriteFile(paths.LastBranchFile, []byte(current), 0o644)
	}

	folder := filepath.Join(paths.ArchiveDir, fmt.Sprintf("%s-%s", time.Now().Format("2006-01-02"), trimPrefix(previous, "ralph/")))
	if err := os.MkdirAll(folder, 0o755); err != nil {
		return err
	}
	if err := copyIfExists(paths.PRDFile, filepath.Join(folder, "prd.json")); err != nil {
		return err
	}
	if err := copyIfExists(paths.ProgressFile, filepath.Join(folder, "progress.txt")); err != nil {
		return err
	}
	fmt.Printf("Archiving previous run: %s\n", previous)
	fmt.Printf("   Archived to: %s\n", folder)
	if err := ensureProgress(paths.ProgressFile, true); err != nil {
		return err
	}
	return os.WriteFile(paths.LastBranchFile, []byte(current), 0o644)
}

func HighestPriorityUnfinishedStory(prd PRD) (UserStory, bool) {
	var selected UserStory
	found := false
	for _, story := range prd.UserStories {
		if story.Passes {
			continue
		}
		if !found || story.Priority < selected.Priority {
			selected = story
			found = true
		}
	}
	return selected, found
}

func copyIfExists(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

func trimPrefix(value, prefix string) string {
	if len(value) >= len(prefix) && value[:len(prefix)] == prefix {
		return value[len(prefix):]
	}
	return value
}

type TaskSnapshot struct {
	ModTimes map[string]time.Time
}

func SnapshotTaskFiles(taskDir string) (TaskSnapshot, error) {
	snapshot := TaskSnapshot{ModTimes: map[string]time.Time{}}
	matches, err := filepath.Glob(filepath.Join(taskDir, "prd-*.md"))
	if err != nil {
		return snapshot, err
	}
	for _, match := range matches {
		if filepath.Base(match) == "prd-template.md" {
			continue
		}
		info, err := os.Stat(match)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return snapshot, err
		}
		snapshot.ModTimes[match] = info.ModTime()
	}
	return snapshot, nil
}

func DetectLatestChangedTaskFile(taskDir string, before TaskSnapshot) (string, error) {
	matches, err := filepath.Glob(filepath.Join(taskDir, "prd-*.md"))
	if err != nil {
		return "", err
	}
	var selected string
	var selectedTime time.Time
	for _, match := range matches {
		if filepath.Base(match) == "prd-template.md" {
			continue
		}
		info, err := os.Stat(match)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return "", err
		}
		prev, existed := before.ModTimes[match]
		if existed && !info.ModTime().After(prev) {
			continue
		}
		if selected == "" || info.ModTime().After(selectedTime) {
			selected = match
			selectedTime = info.ModTime()
		}
	}
	if selected == "" {
		return "", errors.New("no new or updated .ralph/tasks/prd-*.md file found")
	}
	return selected, nil
}
