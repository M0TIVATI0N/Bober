package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Task представляет структуру задачи
type Task struct {
	ID         int       `json:"id"`                   // Уникальный идентификатор задачи
	Expression string    `json:"expression"`           // Выражение для вычисления
	Status     string    `json:"status"`               // Статус задачи (pending, in_progress, completed)
	Result     float64   `json:"result,omitempty"`     // Результат вычисления
	StartTime  time.Time `json:"start_time,omitempty"` // Время начала выполнения задачи
}

// TaskRequest представляет структуру запроса на создание задачи
type TaskRequest struct {
	Expression string `json:"expression"` // Выражение для вычисления
}

// TaskResponse представляет структуру ответа с ID задачи
type TaskResponse struct {
	ID int `json:"id" ` // Уникальный идентификатор задачи
}

// Operation представляет структуру операции
type Operation struct {
	Operator string `json:"operator" ` // Оператор выражения
	Duration int    `json:"duration"`  // Продолжительность выполнения операции (в секундах)
}

var (
	tasks      []Task         // Список задач
	tasksMutex sync.Mutex     // Мьютекс для защиты списка задач
	operations = []Operation{ // Список доступных операций
		{"+", 2}, {"-", 2}, {"*", 4}, {"/", 4},
	}
	taskIDCounter = 0 // Счетчик ID задач
)

// addTask обрабатывает запрос на добавление новой задачи
func addTask(w http.ResponseWriter, r *http.Request) {
	var taskReq TaskRequest
	err := json.NewDecoder(r.Body).Decode(&taskReq)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Добавление задачи в список
	tasksMutex.Lock()
	defer tasksMutex.Unlock()
	taskIDCounter++
	newTask := Task{
		ID:         taskIDCounter,
		Expression: taskReq.Expression,
		Status:     "pending",
	}
	tasks = append(tasks, newTask)

	// Ответ с ID задачи
	resp := TaskResponse{ID: taskIDCounter}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// getTaskStatus обрабатывает запрос на получение статуса задачи
func getTaskStatus(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("id")

	tasksMutex.Lock()
	defer tasksMutex.Unlock()
	for _, task := range tasks {
		if strconv.Itoa(task.ID) == taskID {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(task)
			return
		}
	}

	http.NotFound(w, r)
}

// getOperations обрабатывает запрос на получение списка операций
func getOperations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(operations)
}

// getTaskForExecution обрабатывает запрос на получение задачи для выполнения
func getTaskForExecution(w http.ResponseWriter, r *http.Request) {
	// Здесь можно реализовать логику приоритизации задач на основе некоторых критериев
	// Для простоты выберем первую невыполненную задачу
	tasksMutex.Lock()
	defer tasksMutex.Unlock()
	for i, task := range tasks {
		if task.Status == "pending" {
			tasks[i].Status = "in_progress"
			tasks[i].StartTime = time.Now()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(task)
			return
		}
	}

	http.NotFound(w, r)
}

// handleResult обрабатывает результат выполнения задачи
func handleResult(w http.ResponseWriter, r *http.Request) {
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	tasksMutex.Lock()
	defer tasksMutex.Unlock()
	for i, t := range tasks {
		if t.ID == task.ID {
			tasks[i] = task
			break
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// indexHandler обрабатывает запрос на главную страницу
func indexHandler(w http.ResponseWriter, r *http.Request) {
	htmlBytes, err := ioutil.ReadFile("index.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(htmlBytes)
}

func main() {
	// Установка обработчиков маршрутов
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/addTask", addTask)
	http.HandleFunc("/getTaskStatus", getTaskStatus)
	http.HandleFunc("/getOperations", getOperations)
	http.HandleFunc("/getTaskForExecution", getTaskForExecution)
	http.HandleFunc("/handleResult", handleResult)

	// Запуск веб-сервера на порту 8080
	log.Fatal(http.ListenAndServe(":8080", nil))
}
