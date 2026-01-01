package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/SlaterL/daily-notes/internal/config"
	"github.com/SlaterL/daily-notes/internal/jira"
	"github.com/SlaterL/daily-notes/internal/notes"
)

var (
	readmeLinks = flag.Bool("readme", false, "Indicates if links to readmes should be included with each jira issue that's found. You probably arent set up for this to be useful.")
	usage       = `Usage:
	start: Create a new daily note
	commit <repo> <commit message>: Append a commit message to your daily note`
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		return
	}
	flag.Parse()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	cfg.ReadmeLinks = *readmeLinks || cfg.ReadmeLinks

	today := time.Now().Local().Format("2006-01-02")

	path, err := notes.DailyNotePath(cfg, today)
	if err != nil {
		log.Fatalf("path error: %v", err)
	}

	exists := notes.Exists(path)
	client, err := jira.NewClient(cfg)
	if err != nil {
		log.Fatalf("jira client error: %v", err)
	}

	issues := []jira.Issue{}
	if !exists {
		issues, err = client.SearchIssues()
		if err != nil {
			log.Fatalf("jira search error: %v", err)
		}
	}

	switch os.Args[1] {
	case "start":
		errStartDay := startDay(path, today, issues)
		if errStartDay != nil {
			log.Fatal(errStartDay)
		}
	case "commit":
		errAppendCommit := appendCommit(exists, path, today, issues, cfg.ExcludeCommits)
		if errAppendCommit != nil {
			log.Fatal(errAppendCommit)
		}
		return
	}

}

func startDay(path, today string, issues []jira.Issue) error {
	content, err := notes.RenderBaseNote(today, issues)
	if err != nil {
		return fmt.Errorf("startDay: failed to build template: %v", err)
	}

	if err := notes.Write(path, content); err != nil {
		return fmt.Errorf("startDay: write error: %v", err)
	}

	if len(issues) == 0 {
		fmt.Printf("Created daily note: %s (no active Jira tasks)\n", today+".md")
	} else {
		fmt.Printf("Created daily note: %s (%d Jira tasks)\n", today+".md", len(issues))
	}

	return nil
}

func appendCommit(exists bool, path, today string, issues []jira.Issue, exclude []string) error {
	if len(os.Args) != 4 {
		return fmt.Errorf("appendCommit: incorrect number of args in commit command: (%d)", len(os.Args))
	}
	if !exists {
		content, err := notes.RenderBaseNote(today, issues)
		if err != nil {
			return fmt.Errorf("appendCommit: failed to build template: %v", err)
		}

		if err := notes.Write(path, content); err != nil {
			return fmt.Errorf("appendCommit: write error: %v", err)
		}
	}

	commitRepo := os.Args[2]
	commitMsg := os.Args[3]

	if matchesAnyRegex(commitMsg, exclude) {
		return nil
	}

	content := notes.RenderAppendCommit(commitRepo, commitMsg)
	errAppend := notes.Append(path, []byte(content))
	if errAppend != nil {
		return fmt.Errorf("appendCommit: append error: %v", errAppend)
	}

	return nil
}

func matchesAnyRegex(commitMsg string, patterns []string) bool {
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			// invalid regex — skip or log, depending on preference
			continue
		}

		if re.MatchString(commitMsg) {
			return true
		}
	}

	return false
}
