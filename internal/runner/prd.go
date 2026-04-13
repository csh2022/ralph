package runner

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/snarktank/ralph/internal/project"
	"github.com/snarktank/ralph/internal/templates"
)

const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiRed    = "\033[31m"
	ansiYellow = "\033[33m"
	ansiGreen  = "\033[32m"
	ansiCyan   = "\033[36m"
)

func printPRDBanner() {
	border := ansiBold + ansiCyan + "+----------------------------------------------------------------------------+" + ansiReset
	fmt.Println(border)
	fmt.Printf("%s%s|%-76s|%s\n", ansiBold, ansiCyan, " Ralph interactive PRD session ", ansiReset)
	fmt.Println(border)
	fmt.Printf("%s| %-74s |%s\n", ansiYellow, "If Codex asks whether you trust this directory, choose 'Yes, continue'.", ansiReset)
	fmt.Printf("%s| %-74s |%s\n", ansiGreen, "When you're done:", ansiReset)
	fmt.Printf("%s|   %-72s |%s\n", ansiGreen, "- press Ctrl-D", ansiReset)
	fmt.Printf("%s|   %-72s |%s\n", ansiGreen, "- type exit", ansiReset)
	fmt.Printf("%s| %-74s |%s\n", ansiCyan, "After the session ends, Ralph will look for the latest", ansiReset)
	fmt.Printf("%s| %-74s |%s\n", ansiCyan, ".ralph/tasks/prd-*.md and automatically convert it to", ansiReset)
	fmt.Printf("%s| %-74s |%s\n", ansiCyan, ".ralph/prd.json.", ansiReset)
	fmt.Println(border)
	fmt.Println()
}

func printPRDErrorBanner(message string) {
	border := ansiBold + ansiRed + "+----------------------------------------------------------------------------+" + ansiReset
	fmt.Println()
	fmt.Println(border)
	fmt.Printf("%s%s|%-76s|%s\n", ansiBold, ansiRed, " Ralph PRD flow failed ", ansiReset)
	fmt.Println(border)
	fmt.Printf("%s| %-74s |%s\n", ansiRed, message, ansiReset)
	fmt.Println(border)
}

func RunInteractivePRD(paths project.Paths) error {
	snapshot, err := project.SnapshotTaskFiles(paths.TaskDir)
	if err != nil {
		return err
	}

	printPRDBanner()

	cmd := exec.Command(
		"codex",
		"--dangerously-bypass-approvals-and-sandbox",
		"-C", paths.RepoRoot,
		templates.PRDSessionPrompt,
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("interactive codex session failed: %w", err)
	}

	taskFile, err := project.DetectLatestChangedTaskFile(paths.TaskDir, snapshot)
	if err != nil {
		printPRDErrorBanner("No new or updated .ralph/tasks/prd-*.md file was found.")
		return fmt.Errorf("interactive session ended, but automatic conversion cannot continue: %w", err)
	}

	fmt.Printf("Detected PRD markdown: %s\n", taskFile)
	if err := ConvertPRD(paths, taskFile); err != nil {
		return err
	}
	fmt.Printf("Converted %s to %s\n", taskFile, paths.PRDFile)
	return nil
}

func ConvertPRD(paths project.Paths, taskFile string) error {
	if !filepath.IsAbs(taskFile) {
		taskFile = filepath.Join(paths.RepoRoot, taskFile)
	}
	if _, err := os.Stat(taskFile); err != nil {
		return fmt.Errorf("markdown PRD not found: %w", err)
	}
	prompt := fmt.Sprintf(templates.PRDConvertPromptTemplate, taskFile, paths.PRDFile, paths.PRDFile)

	cmd := exec.Command(
		"codex",
		"exec",
		"--dangerously-bypass-approvals-and-sandbox",
		"-C", paths.RepoRoot,
		"-",
	)
	cmd.Stdin = bytes.NewReader([]byte(prompt))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("codex conversion failed: %w", err)
	}
	prd, err := project.LoadPRD(paths.PRDFile)
	if err != nil {
		return fmt.Errorf("conversion finished, but %s is not valid JSON: %w", paths.PRDFile, err)
	}
	if err := project.Validate(prd); err != nil {
		return fmt.Errorf("conversion finished, but %s is invalid: %w", paths.PRDFile, err)
	}
	return nil
}
