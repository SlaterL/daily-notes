package notes

import (
	"bytes"
	_ "embed"
	"text/template"

	"github.com/SlaterL/daily-notes/internal/jira"
)

//go:embed templates/note.tmpl
var dailyTemplate string

type DailyNoteData struct {
	Date   string
	Issues []jira.Issue
}

func Render(date string, issues []jira.Issue) (string, error) {
	tmpl, err := template.New("daily").Parse(dailyTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, DailyNoteData{
		Date:   date,
		Issues: issues,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
