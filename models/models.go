package models

import "jira-go/pkg/jira"

type Config struct {
	JiraToken  string
	JiraURL    string
	OllamaHost string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Структура для данных шаблона
type Stats struct {
	ModelCount int `json:"modelCount"`
	TaskCount  int `json:"taskCount"`
}

type TemplateData struct {
	Models        []map[string]interface{} `json:"models"`
	Tasks         []jira.JiraTask          `json:"tasks"`
	Error         string                   `json:"error"`
	SelectedModel string                   `json:"selectedModel"`
	Stats         Stats                    `json:"stats"`
}
