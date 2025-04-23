package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

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

func sendOllamaMessage(OllamaHost string, model string, messages []string) error {
	log.Printf("Отправка сообщения в модель %s", model)

	data := map[string]interface{}{
		"model": model,
		"messages": []interface{}{
			map[string]string{
				"role":    "user",
				"content": messages[0],
			},
		},
		"stream": false,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Ошибка сериализации JSON: %v", err)
		return fmt.Errorf("ошибка сериализации JSON: %v", err)
	}

	url := OllamaHost + "/api/chat"
	log.Printf("Отправка запроса на %s", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Ошибка создания запроса: %v", err)
		return fmt.Errorf("ошибка создания запроса: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка отправки запроса: %v", err)
		return fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Ошибка от API: %s, тело ответа: %s", resp.Status, string(body))
		return fmt.Errorf("ошибка от API: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка чтения тела ответа: %v", err)
		return fmt.Errorf("ошибка чтения тела ответа: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Ошибка декодирования JSON: %v", err)
		return fmt.Errorf("ошибка декодирования JSON: %v", err)
	}

	if generatedResponse, ok := response["message"].(map[string]interface{})["content"].(string); ok {
		fmt.Println("\nОтвет ИИ:")
		fmt.Println(generatedResponse)
	} else {
		log.Printf("Неожиданный формат ответа: %v", response)
		return fmt.Errorf("неожиданный формат ответа от API")
	}

	return nil
}
func getJiraTask(JiraURL string, JiraToken string, projectKey string) ([]map[string]interface{}, error) {
	log.Printf("Получение задач для проекта %s", projectKey)

	url := fmt.Sprintf(JiraURL+"/rest/api/2/search?jql=project=%s+AND+created>=startOfDay()", projectKey)
	log.Printf("Отправка запроса к Jira API: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+JiraToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения тела ответа: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка от Jira API: %s, тело ответа: %s", resp.Status, string(body))
		return nil, fmt.Errorf("ошибка от Jira API: %s", resp.Status)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("ошибка декодирования JSON: %v", err)
	}

	issues, ok := data["issues"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("неверный формат ответа, отсутствуют задачи")
	}

	var tasks []map[string]interface{}
	for _, issue := range issues {
		task, ok := issue.(map[string]interface{})
		if !ok {
			continue
		}

		fields, ok := task["fields"].(map[string]interface{})
		if !ok {
			continue
		}

		resolution := ""
		if res, ok := fields["resolution"].(map[string]interface{}); ok {
			if name, ok := res["name"].(string); ok {
				resolution = name
			}
		}

		description, _ := fields["description"].(string)
		summary, _ := fields["summary"].(string)
		key, _ := task["key"].(string)

		tasks = append(tasks, map[string]interface{}{
			"key":         key,
			"description": description,
			"summary":     summary,
			"resolution":  resolution,
		})
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("не найдено задач для проекта %s", projectKey)
	}

	return tasks, nil
}

func main() {
	log.Println("Запуск приложения")
	// Загружаем конфигурацию
	config := LoadConfig()

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
		body, _ := ioutil.ReadAll(resp.Body)
		log.Fatalf("Ошибка от Ollama API: %s, тело ответа: %s", resp.Status, string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Ошибка чтения тела ответа: %v", err)
	}

	var tags map[string]interface{}
	if err := json.Unmarshal(body, &tags); err != nil {
		log.Fatalf("Ошибка декодирования JSON: %v", err)
	}

	models, ok := tags["models"].([]interface{})
	if !ok {
		log.Fatal("Неверный формат списка моделей")
	}

	fmt.Println("Доступные модели ИИ:")
	for i, m := range models {
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

	if choice < 1 || choice > len(models) {
		log.Fatal("Неверный номер модели")
	}

	selectedModel, ok := models[choice-1].(map[string]interface{})
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

	tasks, err := getJiraTask(config.JiraURL, config.JiraToken, projectKey)
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
	fullMessage := fmt.Sprintf(

		`%s\nЗадача %s.  %s.`,
		message,
		activeTask["key"].(string),
		activeTask["description"].(string))

	fmt.Println("\nОтправляем ИИ следующий запрос:")
	fmt.Println(fullMessage)

	if err := sendOllamaMessage(config.OllamaHost, modelName, []string{fullMessage}); err != nil {
		log.Fatal(err)
	}

	log.Println("Приложение завершило работу")
}
