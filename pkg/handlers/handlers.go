package handlers

import (
	"encoding/json"
	"html/template"
	"jira-go/pkg/config"
	"jira-go/pkg/jira"
	"jira-go/pkg/ollama"
	"log"
	"net/http"
	"sync"
)

var (
	tmpl      *template.Template
	appData   *AppData
	configObj *config.Config
	mu        sync.RWMutex
)

type AppData struct {
	Models        []map[string]interface{}
	Tasks         []jira.JiraTask
	Error         string
	SelectedModel string
}

type ProjectForm struct {
	ProjectKey string `json:"projectKey"`
	Messages   string `json:"messages"`
}

func NewHandlerService(cfg *config.Config) *AppData {
	configObj = cfg
	models, err := ollama.GetOllamaModels(configObj.OllamaHost)
	return &AppData{
		Models: models,
		Error:  err.Error(),
	}
}
func InitHandlers(cfg *config.Config) {
	configObj = cfg
	configObj = cfg

	// Загружаем модели при запуске
	models, err := ollama.GetOllamaModels(configObj.OllamaHost)
	if err != nil {
		log.Printf("Ошибка загрузки моделей: %v", err)
		models = []map[string]interface{}{}
	}

	appData = &AppData{
		Models: models, // это плоский массив моделей
		Tasks:  []jira.JiraTask{},
	}
	// Загружаем шаблоны
	var errParse error
	tmpl, errParse = loadTemplates()
	if errParse != nil {
		log.Fatal("Ошибка загрузки шаблонов:", errParse)
	}
	printTemplateNames(tmpl)
	// Регистрируем handlers
	http.HandleFunc("/select-model", selectModelHandler) // обработчик для выбора модели
	http.HandleFunc("/get-tasks", getTasksHandler)
	http.HandleFunc("/send", sendHandler)
	http.HandleFunc("/api/models", modelsHandler)
	http.HandleFunc("/api/tasks", tasksHandler)
	http.HandleFunc("/", indexHandler)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	data := map[string]interface{}{
		"models": appData.Models,
		"tasks":  appData.Tasks,
		"error":  appData.Error,
	}

	err := tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var formData ProjectForm
	if err := json.NewDecoder(r.Body).Decode(&formData); err != nil {
		http.Error(w, "Ошибка parsing JSON", http.StatusBadRequest)
		return
	}

	if formData.ProjectKey == "" {
		http.Error(w, "Ключ проекта обязателен", http.StatusBadRequest)
		return
	}

	// Получаем задачи
	tasks, err := jira.GetJiraTask(configObj.JiraURL, configObj.JiraToken, formData.ProjectKey)
	if err != nil {
		log.Printf("Ошибка получения задач: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Обновляем данные
	mu.Lock()
	appData.Tasks = tasks
	appData.Error = ""
	mu.Unlock()

	// Возвращаем JSON ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tasks":   tasks,
		"count":   len(tasks),
	})
}

func sendHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Здесь будет логика отправки запросов к ИИ
	w.Write([]byte("Функция отправки запросов к ИИ будет реализована позже"))
}

func modelsHandler(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
	models, err := ollama.GetOllamaModels(configObj.OllamaHost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	appData.Models = models
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(appData.Models)
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(appData.Tasks)
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

// Загружаем шаблоны
func loadTemplates() (*template.Template, error) {
	//tmpl := template.New("base")

	// Загружаем все шаблоны
	templates := []string{
		"templates/base.html",
		"templates/static/header.html",
		"templates/static/stats.html",
		"templates/static/models.html",
		"templates/static/task_form.html",
		"templates/static/tasks.html",
		"templates/static/ai_form.html",
		"templates/index.html",
	}
	// Парсим все файлы и используем имя файла как имя шаблона
	tmpl := template.New("").Funcs(template.FuncMap{
		"len": func(items interface{}) int {
			switch v := items.(type) {
			case []map[string]interface{}:
				return len(v)
			case []jira.JiraTask:
				return len(v)
			default:
				return 0
			}
		},
	})

	return tmpl.ParseFiles(templates...)
}

// вспомогательную функцию для отладки
func printTemplateNames(tmpl *template.Template) {
	log.Printf("Загруженные шаблоны:")
	for _, t := range tmpl.Templates() {
		log.Printf("  - %s", t.Name())
	}
}
