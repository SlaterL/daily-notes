package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/SlaterL/daily-notes/internal/config"
	"github.com/SlaterL/daily-notes/internal/jira"
	"github.com/SlaterL/daily-notes/internal/notes"
)

var (
	readmeLinks = flag.Bool("readme", false, "Indicates if links to readmes should be included with each jira issue that's found. You probably arent set up for this to be useful.")
)

func main() {
	flag.Parse()
	log.SetFlags(0)

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

	if notes.Exists(path) {
		fmt.Printf("Daily note already exists: %s\n", today+".md")
		return
	}

	client, err := jira.NewClient(cfg)
	if err != nil {
		log.Fatalf("jira client error: %v", err)
	}

	issues, err := client.SearchIssues()
	if err != nil {
		log.Fatalf("jira search error: %v", err)
	}

	content, err := notes.Render(today, issues)
	if err != nil {
		log.Fatalf("failed to build template: %v", err)
	}

	if err := notes.Write(path, content); err != nil {
		log.Fatalf("write error: %v", err)
	}

	if len(issues) == 0 {
		fmt.Printf("Created daily note: %s (no active Jira tasks)\n", today+".md")
	} else {
		fmt.Printf("Created daily note: %s (%d Jira tasks)\n", today+".md", len(issues))
	}
}
