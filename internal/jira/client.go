package jira

import (
	"encoding/base64"
	"net/http"

	"daily-notes/internal/config"
)

type Client struct {
	http     *http.Client
	baseURL  string
	authHead string
	cfg      *config.Config
}

func NewClient(cfg *config.Config) (*Client, error) {
	auth := cfg.Jira.Email + ":" + cfg.Token
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))

	return &Client{
		http:     &http.Client{},
		baseURL:  cfg.Jira.BaseURL,
		authHead: "Basic " + encoded,
		cfg:      cfg,
	}, nil
}
