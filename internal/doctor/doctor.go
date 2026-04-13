package doctor

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/snarktank/ralph/internal/project"
)

type Check struct {
	Name   string
	OK     bool
	Detail string
	Fatal  bool
}

func Run(paths project.Paths) ([]Check, error) {
	checks := []Check{
		{Name: "git repository", OK: paths.RepoRoot != "", Detail: paths.RepoRoot, Fatal: true},
		checkCommand("codex installed", "codex", true),
		checkPath(".ralph directory", paths.RalphDir, true),
		checkPath(".ralph/prd.json", paths.PRDFile, true),
		checkPath(".ralph/progress.txt", paths.ProgressFile, true),
		checkPath(".ralph/tasks", paths.TaskDir, true),
		checkPath(".codex/skills/ralph-prd/SKILL.md", paths.RalphPRDSkillFile, true),
		checkPath(".codex/skills/ralph-prd-converter/SKILL.md", paths.RalphPRDConverterSkillFile, true),
	}

	if fileExists(paths.PRDFile) {
		prd, err := project.LoadPRD(paths.PRDFile)
		if err != nil {
			checks = append(checks, Check{Name: "prd.json parse", OK: false, Detail: err.Error(), Fatal: true})
		} else {
			checks = append(checks, Check{Name: "prd.json valid", OK: project.Validate(prd) == nil, Detail: prd.BranchName, Fatal: true})
			checks = append(checks, Check{Name: "unfinished stories remain", OK: project.HasUnfinishedStories(prd), Detail: unfinishedDetail(prd), Fatal: false})
			if story, ok := project.HighestPriorityUnfinishedStory(prd); ok {
				checks = append(checks, Check{Name: "next story", OK: true, Detail: fmt.Sprintf("%s %s", story.ID, story.Title), Fatal: false})
			}
		}
	}

	return checks, nil
}

func Print(checks []Check) {
	for _, check := range checks {
		status := "OK"
		if !check.OK {
			status = "FAIL"
		}
		if check.Detail != "" {
			fmt.Printf("%s  %s - %s\n", status, check.Name, check.Detail)
		} else {
			fmt.Printf("%s  %s\n", status, check.Name)
		}
	}
}

func HasFatalFailure(checks []Check) bool {
	for _, check := range checks {
		if check.Fatal && !check.OK {
			return true
		}
	}
	return false
}

func checkCommand(name, command string, fatal bool) Check {
	_, err := exec.LookPath(command)
	return Check{Name: name, OK: err == nil, Detail: command, Fatal: fatal}
}

func checkPath(name, path string, fatal bool) Check {
	_, err := os.Stat(path)
	return Check{Name: name, OK: err == nil, Detail: path, Fatal: fatal}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func unfinishedDetail(prd project.PRD) string {
	count := 0
	for _, story := range prd.UserStories {
		if !story.Passes {
			count++
		}
	}
	if count == 0 {
		return "none"
	}
	return fmt.Sprintf("%d remaining", count)
}
