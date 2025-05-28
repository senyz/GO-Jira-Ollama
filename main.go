package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	JiraToken  string
	JiraURL    string
	OllamaHost string
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func LoadConfig() *Config {
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

func getOllamaModels(OllamaHost string) ([]map[string]interface{}, error) {
	log.Printf("Получение списка моделей Ollama")
	tagsURL := "/api/tags"
	resp, err := http.Get(OllamaHost + tagsURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса списка моделей: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Ошибка от Ollama API: %s, тело ответа: %s", resp.Status, string(body))
		return nil, fmt.Errorf("ошибка от Ollama API: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка чтения тела ответа: %v", err)
		return nil, fmt.Errorf("ошибка чтения тела ответа: %v", err)
	}

	var tags map[string]interface{}
	if err := json.Unmarshal(body, &tags); err != nil {
		log.Printf("Ошибка декодирования JSON: %v", err)
		return nil, fmt.Errorf("ошибка декодирования JSON: %v", err)
	}

	models, ok := tags["models"].([]interface{})
	if !ok {
		log.Fatal("Неверный формат списка моделей")
	}
	return models, nil
}

func getJiraProjectKey(JiraURL string, JiraToken string) (string, error) {
	log.Printf("Получение ключа проекта Jira")
	url := fmt.Sprintf(JiraURL+"/rest/api/2/search?jql=project=%s+AND+created>=startOfDay()", "")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+JiraToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка чтения тела ответа: %v", err)
		return "", fmt.Errorf("ошибка чтения тела ответа: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("Ошибка декодирования JSON: %v", err)
		return "", fmt.Errorf("ошибка декодирования JSON: %v", err)
	}

	projectKey, ok := data["issues"].([]interface{})[0].(map[string]interface{})["key"].(string)
	if !ok {
		log.Fatal("Неверный формат ключа проекта")
	}
	return projectKey, nil
}

func getJiraTask(JiraURL string, JiraToken string, projectKey string) ([]map[string]interface{}, error) {
	log.Printf("Получение задач для проекта %s", projectKey)
	url := fmt.Sprintf(JiraURL+"/rest/api/2/search?jql=project=%s+AND+created>=startOfDay()", projectKey)
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
		log.Printf("Ошибка чтения тела ответа: %v", err)
		return nil, fmt.Errorf("ошибка чтения тела ответа: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка от Jira API: %s, тело ответа: %s", resp.Status, string(body))
		return nil, fmt.Errorf("ошибка от Jira API: %s", resp.Status)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("Ошибка декодирования JSON: %v", err)
		return nil, fmt.Errorf("ошибка декодирования JSON: %v", err)
	}

	issues, ok := data["issues"].([]interface{})
	if !ok {
		log.Fatal("Неверный формат ответа")
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

	return tasks, nil
}

func sendOllamaMessage(OllamaHost string, model string, messages []string) error {
	log.Printf("Отправка сообщения в модель %s", model)
	data := map[string]interface{}{
		"model":    model,
		"messages": []interface{}{},
	}
	for _, message := range messages {
		data["messages"] = append(data["messages"], map[string]string{
			"role":    "user",
			"content": message,
		})
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Ошибка сериализации JSON: %v", err)
		return fmt.Errorf("ошибка сериализации JSON: %v", err)
	}

	url := OllamaHost + "/api/chat"
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

	return nil
}

func main() {
	log.Println("Запуск приложения")
	config := LoadConfig()
	if config.JiraToken == "" {
		log.Fatal("JIRA_TOKEN не установлен. Добавьте его в .env файл или переменные окружения")
	}

	models, err := getOllamaModels(config.OllamaHost)
	if err != nil {
		log.Fatal(err)
	}
	tasks, err := getJiraTask(config.JiraURL, config.JiraToken, "PROJ-123") // hardcoded project key
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
