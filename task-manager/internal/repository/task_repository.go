package repository

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"task-manager/internal/domain"
	"time"
)

type TaskRepository interface {
	Create(task *domain.Task) error
	Update(task *domain.Task) error
	Delete(taskID, userID string) error
	FindByID(taskID string) (*domain.Task, error)
	FindByUser(userID string) ([]domain.Task, error)
	FindByUserAndStatus(userID string, status domain.TaskStatus) ([]domain.Task, error)
}

type JSONTaskRepository struct {
	mu       sync.RWMutex
	filePath string
	tasks    map[string]*domain.Task // taskID -> Task
}

func NewTaskRepository(dataDir string) (*JSONTaskRepository, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	repo := &JSONTaskRepository{
		filePath: filepath.Join(dataDir, "tasks.json"),
		tasks:    make(map[string]*domain.Task),
	}

	if err := repo.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return repo, nil
}

func (r *JSONTaskRepository) Create(task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tasks[task.ID] = task
	return r.save()
}

func (r *JSONTaskRepository) Update(task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[task.ID]; !exists {
		return errors.New("tarefa não encontrada")
	}

	task.UpdatedAt = time.Now()
	r.tasks[task.ID] = task
	return r.save()
}

func (r *JSONTaskRepository) Delete(taskID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, exists := r.tasks[taskID]
	if !exists {
		return errors.New("tarefa não encontrada")
	}

	if task.UserID != userID {
		return errors.New("não autorizado")
	}

	delete(r.tasks, taskID)
	return r.save()
}

func (r *JSONTaskRepository) FindByID(taskID string) (*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, exists := r.tasks[taskID]
	if !exists {
		return nil, errors.New("tarefa não encontrada")
	}
	
	taskCopy := *task
	return &taskCopy, nil
}

func (r *JSONTaskRepository) FindByUser(userID string) ([]domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Task
	for _, task := range r.tasks {
		if task.UserID == userID {
			taskCopy := *task
			result = append(result, taskCopy)
		}
	}

	// Ordenar por prioridade (maior primeiro) e data de criação
	sort.Slice(result, func(i, j int) bool {
		if result[i].Priority != result[j].Priority {
			return result[i].Priority > result[j].Priority
		}
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result, nil
}

func (r *JSONTaskRepository) FindByUserAndStatus(userID string, status domain.TaskStatus) ([]domain.Task, error) {
	tasks, err := r.FindByUser(userID)
	if err != nil {
		return nil, err
	}

	var filtered []domain.Task
	for _, task := range tasks {
		if task.Status == status {
			filtered = append(filtered, task)
		}
	}

	return filtered, nil
}

func (r *JSONTaskRepository) save() error {
	data, err := json.MarshalIndent(r.tasks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.filePath, data, 0644)
}

func (r *JSONTaskRepository) load() error {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &r.tasks)
}
