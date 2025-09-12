package handlers

import (
	"encoding/json"
	"fmt"
	"jira-go/models"
	"jira-go/pkg/ollama"
	"log"
	"net/http"
)

func modelsHandler(w http.ResponseWriter, r *http.Request) {
	models, err := ollama.GetOllamaModels(configObj.OllamaHost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mu.Lock()
	appData.Models = models
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}

func sendAIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var formData struct {
		Model       string  `json:"model"`
		Messages    string  `json:"messages"`
		Temperature float64 `json:"temperature,omitempty"`
		TaskKey     string  `json:"taskKey,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&formData); err != nil {
		http.Error(w, "Ошибка parsing JSON", http.StatusBadRequest)
		return
	}

	if formData.Messages == "" {
		http.Error(w, "Сообщение обязательно", http.StatusBadRequest)
		return
	}

	if appData.SelectedModel == "" {
		http.Error(w, "Модель не выбрана", http.StatusBadRequest)
		return
	}

	// Если есть ключ задачи, добавляем информацию о задаче
	var fullMessage string
	if formData.TaskKey != "" {
		// Находим задачу в кэше
		var taskInfo string
		mu.RLock()
		for _, task := range appData.Tasks {
			if task.Key == formData.TaskKey {
				taskInfo = fmt.Sprintf("Задача: %s\nОписание: %s\n",
					task.Fields.Summary, task.Fields.Description)
				break
			}
		}
		mu.RUnlock()

		fullMessage = formData.Messages + "\n\n" + taskInfo
	} else {
		fullMessage = formData.Messages
	}

	mess := models.Message{
		Role:    "user",
		Content: fullMessage,
	}

	response, err := ollama.SendOllamaMessage(configObj.OllamaHost, appData.SelectedModel, mess)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"answer":  response,
		"taskKey": formData.TaskKey,
	})
}

// Обновление моделей
func RefreshModels() error {
	models, err := ollama.GetOllamaModels(configObj.OllamaHost)
	if err != nil {
		return err
	}

	mu.Lock()
	appData.Models = models
	mu.Unlock()

	log.Printf("Обновление моделей: %d моделей загружено", len(models))
	return nil
}

func selectModelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	modelName := r.FormValue("model")
	mu.Lock()
	appData.SelectedModel = modelName
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"model":   modelName,
	})
}
