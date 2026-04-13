package app

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/snarktank/ralph/internal/doctor"
	"github.com/snarktank/ralph/internal/project"
	"github.com/snarktank/ralph/internal/runner"
)

func Run(args []string) error {
	if len(args) == 0 {
		return usage()
	}

	switch args[0] {
	case "init":
		return runInit(args[1:])
	case "doctor":
		return runDoctor(args[1:])
	case "run":
		return runRalph(args[1:])
	case "prd":
		return runPRD(args[1:])
	case "help", "--help", "-h":
		return usage()
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usageText)
	}
}

const usageText = `Ralph

Usage:
  ralph init [path] [--force]
  ralph doctor
  ralph prd [--convert-only .ralph/tasks/prd-example.md]
  ralph run [max-iterations] [--tool codex]

Commands:
  init    Create .ralph/ in the target git repository. Defaults to the current directory.
  doctor  Validate the current repository and .ralph setup.
  prd     Start an interactive Codex PRD session, then auto-convert to .ralph/prd.json.
  run     Execute Ralph with Codex. Defaults to 10 iterations.

Examples:
  ralph init
  ralph init /path/to/repo --force
  ralph doctor
  ralph prd
  ralph prd --convert-only .ralph/tasks/prd-example.md
  ralph run
  ralph run 3
`

func usage() error {
	fmt.Print(usageText)
	return nil
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	force := fs.Bool("force", false, "overwrite existing .ralph files")
	fs.SetOutput(os.Stdout)
	if err := fs.Parse(args); err != nil {
		return err
	}

	repoHint := "."
	if fs.NArg() > 0 {
		repoHint = fs.Arg(0)
	}
	paths, err := project.Discover(repoHint)
	if err != nil {
		return err
	}
	if err := project.EnsureInitialized(paths, *force); err != nil {
		return err
	}
	fmt.Printf("Ralph initialized in %s\n", paths.RalphDir)
	fmt.Printf("Codex skills installed in %s\n", paths.CodexSkillsDir)
	return nil
}

func runDoctor(args []string) error {
	if len(args) > 0 {
		return errors.New("doctor does not accept positional arguments")
	}
	paths, err := project.Discover(".")
	if err != nil {
		return err
	}
	checks, err := doctor.Run(paths)
	if err != nil {
		return err
	}
	doctor.Print(checks)
	if doctor.HasFatalFailure(checks) {
		return errors.New("doctor found fatal issues")
	}
	return nil
}

func runRalph(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	tool := fs.String("tool", "codex", "execution backend")
	fs.SetOutput(os.Stdout)
	if err := fs.Parse(args); err != nil {
		return err
	}

	maxIterations := 10
	if fs.NArg() > 0 {
		_, err := fmt.Sscanf(fs.Arg(0), "%d", &maxIterations)
		if err != nil || maxIterations < 1 {
			return fmt.Errorf("invalid max iterations %q", fs.Arg(0))
		}
	}

	paths, err := project.Discover(".")
	if err != nil {
		return err
	}
	checks, err := doctor.Run(paths)
	if err != nil {
		return err
	}
	if doctor.HasFatalFailure(checks) {
		doctor.Print(checks)
		return errors.New("doctor found fatal issues")
	}

	return runner.Run(runner.Config{
		Paths:         paths,
		MaxIterations: maxIterations,
		Tool:          *tool,
	})
}

func runPRD(args []string) error {
	fs := flag.NewFlagSet("prd", flag.ContinueOnError)
	convertOnly := fs.String("convert-only", "", "convert an existing markdown PRD into .ralph/prd.json")
	fs.SetOutput(os.Stdout)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return errors.New("prd does not accept positional arguments")
	}

	paths, err := project.Discover(".")
	if err != nil {
		return err
	}
	if err := project.EnsureInitialized(paths, false); err != nil {
		return err
	}
	checks, err := doctor.Run(paths)
	if err != nil {
		return err
	}
	if doctor.HasFatalFailure(checks) {
		doctor.Print(checks)
		return errors.New("doctor found fatal issues")
	}

	if *convertOnly != "" {
		return runner.ConvertPRD(paths, *convertOnly)
	}
	return runner.RunInteractivePRD(paths)
}
