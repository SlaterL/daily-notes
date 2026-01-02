package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/SlaterL/daily-notes/internal/config"
	"github.com/SlaterL/daily-notes/internal/jira"
	"github.com/SlaterL/daily-notes/internal/notes"
)

var (
	readmeLinks       = flag.Bool("readme", false, "Indicates if links to readmes should be included with each jira issue that's found. You probably arent set up for this to be useful.")
	todayFlag         = flag.String("day", time.Now().Local().Format("2006-01-02"), "Indicates the day that will be used for note generation and summary. In yyyy-mm-dd format")
	cmd               = flag.String("cmd", "start", "Which command should run (start, commit, summarize)")
	repo              = flag.String("repo", "", "The repo where a commit message was added")
	msg               = flag.String("msg", "", "The commit message")
	summaryPromptBase = `
You are summarizing a single daily note for a Senior Software Engineer.

TASKS:
- Review daily notes and extract information from the bullet points and checked tasks. Present it as a concise list of bullet points containing what happened today.

RULES:
- Output ONLY a single bullet list
- Each bullet must describe one concrete event, task, or decision
- Use past tense
- Be factual and neutral
- Be sure to include points related to commit messages if available
- Do NOT add explanations, headings, or commentary
- Do NOT infer or invent events
- Do NOT include goals, plans, or future work unless explicitly stated in the text
- Do NOT repeat information across bullets

FORMAT:
- One markdown bullet list
- Each bullet starts with "- "
- No blank lines before or after the list
- For commit message bullets, use this format:
"""- (repo): 
	- <commit message summary>
	- <commit message summary>"""

INPUT:
<<<
%s
>>>
`
)

func main() {
	flag.Parse()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	cfg.ReadmeLinks = *readmeLinks || cfg.ReadmeLinks

	var today string
	if todayFlag != nil {
		today = *todayFlag
	} else {
		log.Fatalf("nil date error")
	}

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

	switch *cmd {
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
	case "sum":
		if !exists {
			log.Fatal("Nothing to summarize.")
		}
		summaryPath, errSummaryPath := notes.DailyNoteSummaryPath(cfg, today)
		if errSummaryPath != nil {
			log.Fatal(errSummaryPath)
		}
		summarizeDay(path, summaryPath, cfg.OllamaModel)
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
	if *repo == "" || *msg == "" {
		return errors.New("appendCommit: repo or commit message not specified")
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

	if matchesAnyRegex(*msg, exclude) {
		return nil
	}

	content := notes.RenderAppendCommit(*repo, *msg)
	errAppend := notes.Append(path, []byte(content))
	if errAppend != nil {
		return fmt.Errorf("appendCommit: append error: %v", errAppend)
	}

	return nil
}

type generateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func summarizeDay(notesPath, summaryPath, model string) error {
	data, errRead := os.ReadFile(notesPath)
	if errRead != nil {
		return fmt.Errorf("summarizeDay: read error: %v", errRead)
	}

	prompt := fmt.Sprintf(summaryPromptBase, string(data))

	payload := map[string]any{
		// "model":  "qwen3-coder:30b",
		"model":  model,
		"prompt": prompt,
		"stream": false,
	}

	body, _ := json.Marshal(payload)

	fmt.Println("Generating note summary. This may take some time.")
	resp, err := http.Post(
		"http://localhost:11434/api/generate",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var gr generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return fmt.Errorf("summarizeDay: error parsing ollama response: %v", err)
	}

	output := stripThinkBlock(gr.Response)
	notes.Write(summaryPath, output)

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

func stripThinkBlock(s string) string {
	const marker = "</think>"
	if idx := strings.Index(s, marker); idx != -1 {
		return strings.TrimSpace(s[idx+len(marker):])
	}
	return strings.TrimSpace(s)
}
