package notes

import (
	"strings"

	"daily-notes/internal/jira"
)

func Render(date string, issues []jira.Issue) string {
	var b strings.Builder

	b.WriteString("## 🏆 Major Accomplishments\n")
	b.WriteString("- \n\n")
	b.WriteString("## 📋 Jira Tasks\n")

	if len(issues) == 0 {
		b.WriteString("\n")
	} else {
		for _, i := range issues {
			b.WriteString("[**" + i.Key + "**](" + i.URL + ") (" + i.Status + ") — " + i.Summary + "\n")
			b.WriteString("- [ ] \n\n")
		}
	}

	b.WriteString("## 📋 Other Tasks\n")
	b.WriteString("- [ ] Review MRs\n\n")

	b.WriteString("## 📝 Notes\n\n")

	return b.String()
}
