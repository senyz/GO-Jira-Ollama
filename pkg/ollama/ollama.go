package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"jira-go/models"
	"log"
	"net/http"
)

// sendOllamaMessage - отправка сообщения в модель Ollama
func SendOllamaMessage(OllamaHost string, model string, messages models.Message) (string, error) {
	log.Printf("Отправка сообщения в модель %s", model)

	data := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"stream":   false,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("ошибка сериализации JSON: %v", err)
	}

	url := fmt.Sprintf("http://%s/api/chat", OllamaHost)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ошибка от API: %s, %s", resp.Status, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ошибка декодирования ответа: %v", err)
	}

	if message, ok := result["message"].(map[string]interface{}); ok {
		if content, ok := message["content"].(string); ok {
			return content, nil
		}
	}
	return "", fmt.Errorf("не удалось извлечь ответ")
}

func GetOllamaModels(OllamaHost string) ([]map[string]interface{}, error) {
	log.Printf("Получение списка моделей Ollama")
	var models []map[string]interface{}
	tagsURL := "/api/tags"
	resp, err := http.Get(OllamaHost + tagsURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса списка моделей: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Ошибка от Ollama API: %s, тело ответа: %s", resp.Status, string(body))
		return nil, fmt.Errorf("ошибка от Ollama API: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка чтения тела ответа: %v", err)
		return nil, fmt.Errorf("ошибка чтения тела ответа: %v", err)
	}

	var tags map[string]interface{}
	if err := json.Unmarshal(body, &tags); err != nil {
		log.Printf("Ошибка декодирования JSON: %v", err)
		return nil, fmt.Errorf("ошибка декодирования JSON: %v", err)
	}

	models = append(models, tags)
	return models, nil
}
