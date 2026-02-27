package service

import (
	"errors"
	"time"
	"task-manager/internal/domain"
	"task-manager/internal/repository"
)

type TaskService struct {
	taskRepo *repository.JSONTaskRepository
}

func NewTaskService(taskRepo *repository.JSONTaskRepository) *TaskService {
	return &TaskService{taskRepo: taskRepo}
}

func (s *TaskService) Create(userID string, req domain.CreateTaskRequest) (*domain.Task, error) {
	if req.Title == "" {
		return nil, errors.New("título é obrigatório")
	}

	if req.Priority < 1 || req.Priority > 5 {
		req.Priority = 1
	}

	task := &domain.Task{
		ID:          generateID(),
		Title:       req.Title,
		Description: req.Description,
		Status:      domain.StatusPending,
		Priority:    req.Priority,
		UserID:      userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		DueDate:     req.DueDate,
	}

	if err := s.taskRepo.Create(task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskService) GetUserTasks(userID string, status string) ([]domain.Task, error) {
	if status != "" {
		return s.taskRepo.FindByUserAndStatus(userID, domain.TaskStatus(status))
	}
	return s.taskRepo.FindByUser(userID)
}

func (s *TaskService) Update(userID, taskID string, req domain.UpdateTaskRequest) (*domain.Task, error) {
	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return nil, err
	}

	if task.UserID != userID {
		return nil, errors.New("não autorizado")
	}

	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Status != "" {
		task.Status = req.Status
	}
	if req.Priority != 0 {
		task.Priority = req.Priority
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}

	if err := s.taskRepo.Update(task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskService) Delete(userID, taskID string) error {
	return s.taskRepo.Delete(taskID, userID)
}

func (s *TaskService) GetTask(userID, taskID string) (*domain.Task, error) {
	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return nil, err
	}

	if task.UserID != userID {
		return nil, errors.New("não autorizado")
	}

	return task, nil
}
