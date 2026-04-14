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
	UntilDone     bool
	MaxNoProgress int
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

	if cfg.UntilDone {
		fmt.Printf("Starting Ralph - Tool: %s - Max iterations: until done\n", cfg.Tool)
		fmt.Printf("Circuit breaker: stop after %d consecutive iterations without reducing unfinished stories\n", cfg.MaxNoProgress)
	} else {
		fmt.Printf("Starting Ralph - Tool: %s - Max iterations: %d\n", cfg.Tool, cfg.MaxIterations)
	}

	previousUnfinished := unfinishedCount(prd)
	noProgressCount := 0
	for i := 1; cfg.UntilDone || i <= cfg.MaxIterations; i++ {
		fmt.Println()
		fmt.Println("===============================================================")
		if cfg.UntilDone {
			fmt.Printf("  Ralph Iteration %d (until done) (%s)\n", i, cfg.Tool)
		} else {
			fmt.Printf("  Ralph Iteration %d of %d (%s)\n", i, cfg.MaxIterations, cfg.Tool)
		}
		fmt.Println("===============================================================")

		output, err := runCodex(cfg.Paths)
		if err != nil {
			fmt.Fprintf(os.Stderr, "codex exec failed: %v\n", err)
		}

		reloaded, reloadErr := project.LoadPRD(cfg.Paths.PRDFile)
		if reloadErr == nil && strings.Contains(output, "<promise>COMPLETE</promise>") && !project.HasUnfinishedStories(reloaded) {
			fmt.Println()
			fmt.Println("Ralph completed all tasks!")
			if cfg.UntilDone {
				fmt.Printf("Completed at iteration %d\n", i)
			} else {
				fmt.Printf("Completed at iteration %d of %d\n", i, cfg.MaxIterations)
			}
			return nil
		}
		if strings.Contains(output, "<promise>COMPLETE</promise>") {
			fmt.Printf("Completion signal received, but unfinished stories remain in %s. Continuing...\n", cfg.Paths.PRDFile)
		}
		if reloadErr == nil {
			currentUnfinished := unfinishedCount(reloaded)
			if currentUnfinished < previousUnfinished {
				noProgressCount = 0
			} else {
				noProgressCount++
			}
			previousUnfinished = currentUnfinished
			if cfg.UntilDone && noProgressCount >= cfg.MaxNoProgress {
				fmt.Println()
				fmt.Printf("Ralph stopped after %d consecutive iterations without reducing unfinished stories.\n", cfg.MaxNoProgress)
				fmt.Printf("Check %s for status.\n", cfg.Paths.ProgressFile)
				return nil
			}
		}
		fmt.Printf("Iteration %d complete. Continuing...\n", i)
	}

	fmt.Println()
	fmt.Printf("Ralph reached max iterations (%d) without completing all tasks.\n", cfg.MaxIterations)
	fmt.Printf("Check %s for status.\n", cfg.Paths.ProgressFile)
	return nil
}

func unfinishedCount(prd project.PRD) int {
	count := 0
	for _, story := range prd.UserStories {
		if !story.Passes {
			count++
		}
	}
	return count
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
