package handlers

import (
	"html/template"
	"jira-go/models"
	"jira-go/pkg/config"
	"jira-go/pkg/jira"
	"jira-go/pkg/ollama"
	"log"
	"net/http"
	"sync"
)

type AIRequest struct {
	Model       string  `json:"model"`
	Messages    string  `json:"messages"`
	Temperature float64 `json:"temperature,omitempty"`
}

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

// InitHandlers инициализирует обработчики HTTP запросов
func InitHandlers(cfg *config.Config) {
	configObj = cfg

	// Загружаем модели при запуске
	models, err := ollama.GetOllamaModels(configObj.OllamaHost)
	if err != nil {
		log.Printf("Ошибка загрузки моделей: %v", err)
		models = []map[string]interface{}{}
	}

	appData = &AppData{
		Models: models,
		Tasks:  []jira.JiraTask{},
	}

	// Загружаем шаблоны
	var errParse error
	tmpl, errParse = loadTemplates()
	if errParse != nil {
		log.Fatal("Ошибка загрузки шаблонов:", errParse)
	}

	// Отладочная информация
	printTemplateNames(tmpl)
	fs := http.FileServer(http.Dir("./templates/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	// Регистрируем handlers
	
	http.HandleFunc("/get-tasks", getTasksHandler)
	http.HandleFunc("/send-to-ai", sendAIHandler)
	http.HandleFunc("/select-model", selectModelHandler)
	http.HandleFunc("/api/models", modelsHandler)
	http.HandleFunc("/api/tasks", tasksHandler)
	http.HandleFunc("/", indexHandler)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	data := models.TemplateData{
		Models:        appData.Models,
		Tasks:         appData.Tasks,
		Error:         appData.Error,
		SelectedModel: appData.SelectedModel,
		Stats: models.Stats{
			ModelCount: len(appData.Models),
			TaskCount:  len(appData.Tasks),
		},
	}

	err := tmpl.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Загружаем шаблоны
func loadTemplates() (*template.Template, error) {
	// Сначала загружаем основной шаблон
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		return nil, err
	}

	// Затем парсим остальные шаблоны
	componentTemplates := []string{
		"templates/static/head.html",
		"templates/static/header.html",
		"templates/static/stats.html",
		"templates/static/models.html",
		"templates/static/task_form.html",
		"templates/static/tasks.html",
		"templates/static/ai_form.html",
	}

	return tmpl.ParseFiles(componentTemplates...)
}

// вспомогательную функцию для отладки
func printTemplateNames(tmpl *template.Template) {
	log.Printf("Загруженные шаблоны:")
	for key, t := range tmpl.Templates() {
		log.Printf("%d  - %s", key, t.Name())
	}
}
