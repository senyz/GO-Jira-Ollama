package handlers

import (
	"encoding/json"
	"jira-go/models"
	"jira-go/pkg/ollama"
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
	var formData AIRequest
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
	mess := models.Message{
		Role:    "user",
		Content: formData.Messages,
	}
	response, err := ollama.SendOllamaMessage(configObj.OllamaHost, appData.SelectedModel, mess)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Здесь будет логика отправки запросов к ИИ
	w.Write([]byte(response))
}

// Обновление моделей (можно вызывать периодически)
func RefreshModels() error {
	models, err := ollama.GetOllamaModels(configObj.OllamaHost)
	if err != nil {
		return err
	}

	mu.Lock()
	appData.Models = models
	mu.Unlock()

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
