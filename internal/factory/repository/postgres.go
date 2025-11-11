package repository

import (
	postgresrepo "internal-work-management-service/internal/infrastructure/persistence/postgres"
	notesvc "internal-work-management-service/internal/service/note"
	ordersvc "internal-work-management-service/internal/service/order"
	tasksvc "internal-work-management-service/internal/service/task"
)

// PostgresFactory creates repository implementations (for work with Postgres)
type PostgresFactory struct {
	pool postgresrepo.ConnPool
}

func NewPostgresFactory(pool postgresrepo.ConnPool) PostgresFactory {
	return PostgresFactory{pool: pool}
}

func (f PostgresFactory) NewNoteRepository() notesvc.Repository {
	return postgresrepo.NewNoteRepo(f.pool)
}

func (f PostgresFactory) NewTaskRepository() tasksvc.Repository {
	return postgresrepo.NewTaskRepo(f.pool)
}

func (f PostgresFactory) NewOrderRepository() ordersvc.Repository {
	return postgresrepo.NewOrderRepo(f.pool)
}
