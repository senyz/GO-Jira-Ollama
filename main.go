package main

import (
	"fmt"
	"html/template"

	"jira-go/pkg/config"
	"jira-go/pkg/jira"
	"jira-go/pkg/ollama"
	"log"
	"net/http"
)

func main() {
	log.Println("Запуск приложения")
	config := config.LoadConfig()
	if config.JiraToken == "" {
		log.Fatal("JIRA_TOKEN не установлен. Добавьте его в .env файл или переменные окружения")
	}

	models, err := ollama.GetOllamaModels(config.OllamaHost)
	if err != nil {
		log.Fatal(err)
	}
	var projectKey string
	fmt.Print("Введите ключ проекта Jira: ")
	if _, err := fmt.Scanln(&projectKey); err != nil {
		log.Fatalf("Ошибка ввода: %v", err)
	}

	tasks, err := jira.GetJiraTask(config.JiraURL, config.JiraToken, projectKey)
	if err != nil {
		log.Fatal(err)
	}

	templateFiles := map[string]interface{}{
		"models": models,
		"tasks":  tasks,
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err = tmpl.Execute(w, templateFiles)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	fmt.Println("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
