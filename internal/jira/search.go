package jira

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Issue struct {
	Key        string
	Summary    string
	URL        string
	Status     string
	Components []string
}

type searchResponse struct {
	Issues []struct {
		Key    string `json:"key"`
		Fields struct {
			Summary string `json:"summary"`
			Status  struct {
				Name string `json:"name"`
			} `json:"status"`
			Components []struct {
				Name string `json:"name"`
			} `json:"components"`
		} `json:"fields"`
		Self string `json:"self"`
	} `json:"issues"`
}

func (c *Client) SearchIssues() ([]Issue, error) {
	jql := `assignee = currentUser() AND statusCategory = "In Progress"`

	if len(c.cfg.Jira.ProjectFilter) > 0 {
		quoted := make([]string, 0, len(c.cfg.Jira.ProjectFilter))
		for _, p := range c.cfg.Jira.ProjectFilter {
			quoted = append(quoted, `"`+p+`"`)
		}
		jql += " AND project IN (" + strings.Join(quoted, ",") + ")"
	}

	jql += " ORDER BY priority DESC, updated DESC"

	req, err := http.NewRequest(
		"GET",
		c.baseURL+"/rest/api/3/search/jql",
		nil,
	)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("jql", jql)
	q.Set("fields", "summary,status,components")
	q.Set("maxResults", "50")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", c.authHead)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jira search failed: %s", resp.Status)
	}

	var sr searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, err
	}

	var issues []Issue
	for _, i := range sr.Issues {
		if i.Key == "" || i.Fields.Summary == "" {
			continue
		}

		components := []string{}
		if c.cfg.ReadmeLinks {
			for _, comp := range i.Fields.Components {
				components = append(components, comp.Name)
			}
		}
		issues = append(issues, Issue{
			Key:        i.Key,
			Summary:    i.Fields.Summary,
			URL:        c.baseURL + "/browse/" + i.Key,
			Status:     i.Fields.Status.Name,
			Components: components,
		})
	}

	return issues, nil
}
