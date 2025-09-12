package jira

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type JiraTask struct {
	Key    string `json:"key"`
	Fields struct {
		Summary     string `json:"summary"`
		Description string `json:"description"`
		Status      struct {
			Name string `json:"name"`
		} `json:"status"`
		Resolution struct {
			Name string `json:"name"`
		} `json:"resolution"`
	} `json:"fields"`
}

func GetJiraTask(JiraURL string, JiraToken string, projectKey string) ([]JiraTask, error) {
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

	var response struct {
		Issues []JiraTask `json:"issues"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("ошибка декодирования JSON: %v", err)
	}

	if len(response.Issues) == 0 {
		return nil, fmt.Errorf("не найдено задач для проекта %s", projectKey)
	}

	return response.Issues, nil
}

// makeRequest creates and executes an HTTP request with optional authorization and content-type headers.
//
// @param method The HTTP method to use for the request (e.g., GET, POST).
// @param url The URL to which the request is sent.
// @param token The bearer token for authorization (optional).
// @param body The body of the request (optional).
// @return The HTTP response and any error encountered during the request.
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
