package handlers

import (
	"encoding/json"
	"jira-go/pkg/jira"
	"log"
	"net/http"
)

/**
* Handles the retrieval of Jira tasks for a given project.
* This function accepts a POST request with a JSON payload containing the project key,
* fetches the tasks from Jira, and returns them in JSON format.
*
* @param w The HTTP response writer.
* @param r The HTTP request object.
 */
func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем Content-Type
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Неверный Content-Type. Ожидается application/json", http.StatusBadRequest)
		return
	}

	var formData struct {
		ProjectKey string `json:"projectKey"`
	}

	if err := json.NewDecoder(r.Body).Decode(&formData); err != nil {
		log.Printf("Ошибка декодирования JSON: %v", err)
		http.Error(w, "Ошибка parsing JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if formData.ProjectKey == "" {
		http.Error(w, "Ключ проекта обязателен", http.StatusBadRequest)
		return
	}

	log.Printf("Получение задач для проекта: %s", formData.ProjectKey)

	tasks, err := jira.GetJiraTask(configObj.JiraURL, configObj.JiraToken, formData.ProjectKey)
	if err != nil {
		log.Printf("Ошибка получения задач: %v", err)
		http.Error(w, "Ошибка получения задач: "+err.Error(), http.StatusInternalServerError)
		return
	}

	mu.Lock()
	appData.Tasks = tasks
	appData.Error = ""
	mu.Unlock()

	log.Printf("Получено %d задач для проекта %s", len(tasks), formData.ProjectKey)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tasks":   tasks,
		"count":   len(tasks),
	})
}

/**
* Handles HTTP requests to retrieve the list of tasks.
* It acquires a read lock to ensure thread-safe access to the shared data,
* sets the response content type to JSON, and encodes the tasks data as JSON in the response.
*
* @param w The HTTP response writer to write the response to.
* @param r The HTTP request object containing the request details.
 */
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(appData.Tasks)
}
