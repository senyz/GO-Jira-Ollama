package models

type Config struct {
	JiraToken  string
	JiraURL    string
	OllamaHost string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
