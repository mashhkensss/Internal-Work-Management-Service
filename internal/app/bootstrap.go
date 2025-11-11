package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"internal-work-management-service/internal/app/config"
	"internal-work-management-service/internal/app/handlers/notes"
	"internal-work-management-service/internal/app/handlers/orders"
	"internal-work-management-service/internal/app/handlers/tasks"
	"internal-work-management-service/internal/app/middleware"
	"internal-work-management-service/internal/app/router"
	repositoryfactory "internal-work-management-service/internal/factory/repository"
	servicefactory "internal-work-management-service/internal/factory/service"
	"internal-work-management-service/internal/infrastructure/persistence/postgres"
)

// BuildApp wires the application dependencies and returns an HTTP handler. Also there are factories for services and
// repositories.
func BuildApp(ctx context.Context, cfg config.Config) (http.Handler, func() error, error) {
	pool, err := postgres.NewPool(ctx, cfg.PostgresDSN, postgres.PoolConfig{
		MaxConns:          20,
		MinConns:          2,
		MaxConnLifetime:   30 * time.Minute,
		MaxConnIdleTime:   5 * time.Minute,
		HealthCheckPeriod: time.Minute,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("bootstrap: create pool: %w", err)
	}

	repoFactory := repositoryfactory.NewPostgresFactory(pool)
	noteRepo := repoFactory.NewNoteRepository()
	taskRepo := repoFactory.NewTaskRepository()
	orderRepo := repoFactory.NewOrderRepository()

	svcFactory := servicefactory.New(servicefactory.RepositorySet{
		Notes:  noteRepo,
		Tasks:  taskRepo,
		Orders: orderRepo,
	})

	noteSvc := svcFactory.NewNoteService()
	taskSvc := svcFactory.NewTaskService()
	orderSvc := svcFactory.NewOrderService()

    deps := router.Deps{
        Notes:  notes.New(noteSvc),
        Tasks:  tasks.New(taskSvc),
        Orders: orders.New(orderSvc),
        Middlewares: router.Middlewares{
            Auth: middleware.JWT(cfg.JWTSecret),
        },
    }

	chiRouter := chi.NewRouter()
	chiRouter.Get("/live", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	chiRouter.Mount("/", router.Router(deps))

	closeFn := func() error {
		pool.Close()
		return nil
	}

	log.Printf("bootstrap complete: connected to database and router initialised")

	return chiRouter, closeFn, nil
}
