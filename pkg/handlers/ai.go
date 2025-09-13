package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

	// Читаем тело запроса для логирования
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения тела запроса", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Логируем сырое тело запроса
	log.Printf("Raw request body: %s", string(bodyBytes))

	// Восстанавливаем тело для декодирования
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	var formData struct {
		Model       string  `json:"model"`
		Messages    string  `json:"messages"`
		Temperature float64 `json:"temperature,omitempty,string"`
		TaskKey     string  `json:"taskKey,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&formData); err != nil {
		log.Printf("JSON parsing error: %v", err)
		log.Printf("Request body that caused error: %s", string(bodyBytes))
		http.Error(w, "Ошибка parsing JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Закрываем тело после чтения
	r.Body.Close()

	log.Printf("Parsed data: Model=%s, Messages=%s, Temperature=%f, TaskKey=%s",
		formData.Model, formData.Messages, formData.Temperature, formData.TaskKey)

	if formData.Messages == "" {
		http.Error(w, "Сообщение обязательно", http.StatusBadRequest)
		return
	}

	if appData.SelectedModel == "" {
		if formData.Model != "" {
			appData.SelectedModel = formData.Model
		} else {
			http.Error(w, "Модель не выбрана", http.StatusBadRequest)
			return
		}

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

	mess := []models.Message{
		{
			Role:    "user",
			Content: fullMessage,
		},
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
