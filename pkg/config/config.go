package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	JiraToken  string
	JiraURL    string
	OllamaHost string
}

func LoadConfig() *Config {
	// Загружаем .env файл
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, using environment variables")
	}

	return &Config{
		JiraToken:  getEnv("JIRA_TOKEN", ""),
		JiraURL:    getEnv("JIRA_URL", "https://jira.officesvc.bz"),
		OllamaHost: getEnv("OLLAMA_HOST", "host.docker.internal:11434"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
