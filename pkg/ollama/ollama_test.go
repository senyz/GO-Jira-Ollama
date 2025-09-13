package ollama

import (
	"encoding/json"
	"jira-go/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetOllamaModels(t *testing.T) {
	// Создаем тестовый HTTP сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что путь запроса правильный
		if r.URL.Path != "/api/tags" {
			t.Errorf("Expected request to /api/tags, got %s", r.URL.Path)
		}

		// Подготавливаем тестовые данные
		response := map[string]interface{}{
			"models": []map[string]interface{}{
				{
					"name": "llama2",
					"size": 3825819519,
				},
				{
					"name": "mistral",
					"size": 4683083392,
				},
			},
		}

		// Отправляем JSON ответ
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Тест успешного получения моделей
	t.Run("Successful model retrieval", func(t *testing.T) {
		models, err := GetOllamaModels(server.URL)

		// Проверяем, что ошибка отсутствует
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Проверяем, что получили данные
		if len(models) == 0 {
			t.Error("Expected to receive models, got empty slice")
		}

		// Проверяем структуру полученных данных
		if len(models) != 2 {
			t.Errorf("Expected 1 model map, got %d", len(models))
		}
	})

	// Тест ошибки при недоступном сервере
	t.Run("Server unavailable", func(t *testing.T) {
		// Используем заведомо неправильный URL
		_, err := GetOllamaModels("http://localhost:99999")

		// Проверяем, что получили ошибку
		if err == nil {
			t.Error("Expected error for unavailable server, got nil")
		}
	})

	// Тест ошибки при HTTP статусе не 200
	t.Run("Non-200 HTTP status", func(t *testing.T) {
		// Создаем сервер, который возвращает 500 ошибку
		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer errorServer.Close()

		_, err := GetOllamaModels(errorServer.URL)

		// Проверяем, что получили ошибку
		if err == nil {
			t.Error("Expected error for 500 status, got nil")
		}
	})

	// Тест ошибки при невалидном JSON
	t.Run("Invalid JSON response", func(t *testing.T) {
		// Создаем сервер, который возвращает невалидный JSON
		invalidJSONServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{ invalid json }"))
		}))
		defer invalidJSONServer.Close()

		_, err := GetOllamaModels(invalidJSONServer.URL)

		// Проверяем, что получили ошибку
		if err == nil {
			t.Error("Expected error for invalid JSON, got nil")
		}
	})
}

// Тест для функции SendOllamaMessage (дополнительный тест для полноты покрытия)
func TestSendOllamaMessage(t *testing.T) {
	// Создаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем метод и путь
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/chat" {
			t.Errorf("Expected request to /api/chat, got %s", r.URL.Path)
		}

		// Отправляем тестовый ответ
		response := map[string]interface{}{
			"message": map[string]interface{}{
				"content": "Test response from model",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Тест успешной отправки сообщения
	t.Run("Successful message sending", func(t *testing.T) {
		messages := []models.Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		}

		response, err := SendOllamaMessage(server.URL[7:], "test-model", messages) // Убираем "http://" из URL

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if response != "Test response from model" {
			t.Errorf("Expected 'Test response from model', got %s", response)
		}
	})

	// Тест ошибки при недоступном сервере
	t.Run("Server unavailable", func(t *testing.T) {
		messages := []models.Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		}

		_, err := SendOllamaMessage("localhost:99999", "test-model", messages)

		if err == nil {
			t.Error("Expected error for unavailable server, got nil")
		}
	})
}
