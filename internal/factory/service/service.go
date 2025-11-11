package service

import (
	notesvc "internal-work-management-service/internal/service/note"
	ordersvc "internal-work-management-service/internal/service/order"
	tasksvc "internal-work-management-service/internal/service/task"
)

// RepositorySet groups repository dependencies required for building services
type RepositorySet struct {
	Notes  notesvc.Repository
	Tasks  tasksvc.Repository
	Orders ordersvc.Repository
}

// Factory which also stores required repositories
type Factory struct {
	repos RepositorySet
}

func New(repos RepositorySet) Factory {
	return Factory{repos: repos}
}

func (f Factory) NewNoteService() *notesvc.Service {
	return notesvc.New(f.repos.Notes, nil)
}

func (f Factory) NewTaskService() *tasksvc.Service {
	return tasksvc.New(f.repos.Tasks, nil)
}

func (f Factory) NewOrderService() *ordersvc.Service {
	return ordersvc.New(f.repos.Orders, nil)
}
