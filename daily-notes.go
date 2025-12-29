package main

import (
	"fmt"
	"log"
	"time"

	"daily-notes/internal/config"
	"daily-notes/internal/jira"
	"daily-notes/internal/notes"
)

func main() {
	log.SetFlags(0)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

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

	content := notes.Render(today, issues)

	if err := notes.Write(path, content); err != nil {
		log.Fatalf("write error: %v", err)
	}

	if len(issues) == 0 {
		fmt.Printf("Created daily note: %s (no active Jira tasks)\n", today+".md")
	} else {
		fmt.Printf("Created daily note: %s (%d Jira tasks)\n", today+".md", len(issues))
	}
}
