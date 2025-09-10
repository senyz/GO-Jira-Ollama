package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	"jira-go/models"
	"jira-go/pkg/config"
	"jira-go/pkg/jira"
	"jira-go/pkg/ollama"
	"log"
	"net/http"
	"os"
	"strings"
)

func mainv1() {
	log.Println("Запуск приложения")
	// Загружаем конфигурацию
	config := config.LoadConfig()

	if config.JiraToken == "" {
		log.Fatal("JIRA_TOKEN не установлен. Добавьте его в .env файл или переменные окружения")
	}
	// Получаем список моделей
	log.Println("Получение списка моделей Ollama")
	tagsURL := "/api/tags"
	resp, err := http.Get(config.OllamaHost + tagsURL)
	if err != nil {
		log.Fatalf("Ошибка запроса списка моделей: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Ошибка от Ollama API: %s, тело ответа: %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Ошибка чтения тела ответа: %v", err)
	}

	var tags map[string]interface{}
	if err := json.Unmarshal(body, &tags); err != nil {
		log.Fatalf("Ошибка декодирования JSON: %v", err)
	}

	aimodels, ok := tags["models"].([]interface{})
	if !ok {
		log.Fatal("Неверный формат списка моделей")
	}

	fmt.Println("Доступные модели ИИ:")
	for i, m := range aimodels {
		model, ok := m.(map[string]interface{})
		if !ok {
			continue
		}
		name, ok := model["name"].(string)
		if !ok {
			continue
		}
		fmt.Printf("%d - %s\n", i+1, name)
	}

	var choice int
	fmt.Print("Выберите модель (номер): ")
	if _, err := fmt.Scanln(&choice); err != nil {
		log.Fatalf("Ошибка ввода: %v", err)
	}

	if choice < 1 || choice > len(aimodels) {
		log.Fatal("Неверный номер модели")
	}

	selectedModel, ok := aimodels[choice-1].(map[string]interface{})
	if !ok {
		log.Fatal("Неверный формат модели")
	}

	modelName, ok := selectedModel["name"].(string)
	if !ok {
		log.Fatal("Не удалось получить название модели")
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

	// Выводим информацию о всех задачах
	fmt.Println("\nНайденные задачи:")
	for i, task := range tasks {
		fmt.Printf("\nЗадача #%d (%s):\n", i+1, task["key"].(string))
		fmt.Printf("Заголовок: %s\n", task["summary"])
		fmt.Printf("Описание: %s\n", task["description"])
		fmt.Printf("Решение: %s\n", task["resolution"])
	}

	// Находим первую невыполненную задачу
	var activeTask map[string]interface{}
	for _, task := range tasks {
		if task["resolution"].(string) != "Done" {
			activeTask = task
			break
		}
	}

	if activeTask == nil {
		fmt.Println("\nВсе задачи выполнены (Решение: Done)")
		return
	}

	fmt.Println("\nРаботаем с задачей:", activeTask["key"].(string))
	fmt.Printf("Заголовок: %s\n", activeTask["summary"])
	fmt.Printf("Описание: %s\n", activeTask["description"])
	// Создаем reader для чтения всей строки с пробелами
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("\nВведите ваш вопрос к ИИ: ")
	message, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Ошибка ввода: %v", err)
	}

	// Удаляем символ новой строки в конце
	message = strings.TrimSpace(message)

	// Добавляем информацию о задаче к сообщению
	fullMessage := models.Message{
		Content: fmt.Sprintf(
			`%s\nЗадача %s.  %s.`,
			message,
			activeTask["key"].(string),
			activeTask["description"].(string)),
	}

	fmt.Println("\nОтправляем ИИ следующий запрос:")
	fmt.Println(fullMessage)

	response, err := ollama.SendOllamaMessage(config.OllamaHost, modelName, fullMessage)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nОтвет ИИ:%s\n", response)

	log.Println("Приложение завершило работу")
}
