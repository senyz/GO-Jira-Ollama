package main

import (
	"jira-go/pkg/config"
	"jira-go/pkg/handlers"
	"log"
	"net/http"
)

func main() {
	log.Println("Запуск приложения")
	config := config.LoadConfig()
	if config.JiraToken == "" {
		log.Fatal("JIRA_TOKEN не установлен. Добавьте его в .env файл или переменные окружения")
	}

	// Инициализируем handlers с конфигом
	handlers.InitHandlers(config)

	log.Println("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
