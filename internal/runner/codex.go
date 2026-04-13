package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/snarktank/ralph/internal/project"
)

type Config struct {
	Paths         project.Paths
	MaxIterations int
	Tool          string
}

func Run(cfg Config) error {
	if cfg.Tool != "codex" {
		return fmt.Errorf("only codex is supported in the Go runner for now")
	}
	prd, err := project.LoadPRD(cfg.Paths.PRDFile)
	if err != nil {
		return err
	}
	if err := project.Validate(prd); err != nil {
		return err
	}
	if !project.HasUnfinishedStories(prd) {
		fmt.Printf("Ralph found no unfinished stories in %s. Exiting.\n", cfg.Paths.PRDFile)
		return nil
	}
	if story, ok := project.HighestPriorityUnfinishedStory(prd); ok {
		fmt.Printf("Next story: %s - %s\n", story.ID, story.Title)
	}
	if err := project.ArchiveIfBranchChanged(cfg.Paths, prd); err != nil {
		return err
	}

	fmt.Printf("Starting Ralph - Tool: %s - Max iterations: %d\n", cfg.Tool, cfg.MaxIterations)
	for i := 1; i <= cfg.MaxIterations; i++ {
		fmt.Println()
		fmt.Println("===============================================================")
		fmt.Printf("  Ralph Iteration %d of %d (%s)\n", i, cfg.MaxIterations, cfg.Tool)
		fmt.Println("===============================================================")

		output, err := runCodex(cfg.Paths)
		if err != nil {
			fmt.Fprintf(os.Stderr, "codex exec failed: %v\n", err)
		}

		reloaded, reloadErr := project.LoadPRD(cfg.Paths.PRDFile)
		if reloadErr == nil && strings.Contains(output, "<promise>COMPLETE</promise>") && !project.HasUnfinishedStories(reloaded) {
			fmt.Println()
			fmt.Println("Ralph completed all tasks!")
			fmt.Printf("Completed at iteration %d of %d\n", i, cfg.MaxIterations)
			return nil
		}
		if strings.Contains(output, "<promise>COMPLETE</promise>") {
			fmt.Printf("Completion signal received, but unfinished stories remain in %s. Continuing...\n", cfg.Paths.PRDFile)
		}
		fmt.Printf("Iteration %d complete. Continuing...\n", i)
	}

	fmt.Println()
	fmt.Printf("Ralph reached max iterations (%d) without completing all tasks.\n", cfg.MaxIterations)
	fmt.Printf("Check %s for status.\n", cfg.Paths.ProgressFile)
	return nil
}

func runCodex(paths project.Paths) (string, error) {
	prompt, err := os.ReadFile(paths.CodePromptFile)
	if err != nil {
		return "", err
	}

	cmd := exec.Command(
		"codex",
		"exec",
		"--dangerously-bypass-approvals-and-sandbox",
		"-C", paths.RepoRoot,
		"-",
	)
	cmd.Stdin = bytes.NewReader(prompt)

	var buffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &buffer)
	cmd.Stdout = mw
	cmd.Stderr = mw

	err = cmd.Run()
	return buffer.String(), err
}
