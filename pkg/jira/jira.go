package jira

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// makeRequest - создание запроса
func makeRequest(method, url, token string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	return client.Do(req)
}

func GetJiraTask(JiraURL string, JiraToken string, projectKey string) ([]map[string]interface{}, error) {
	log.Printf("Получение задач для проекта %s", projectKey)

	url := fmt.Sprintf(JiraURL+"/rest/api/2/search?jql=project=%s+AND+created>=startOfDay()", projectKey)
	log.Printf("Отправка запроса к Jira API: %s", url)

	resp, err := makeRequest("GET", url, JiraToken, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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
