package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"internal-work-management-service/internal/app"
	"internal-work-management-service/internal/app/config"
	"internal-work-management-service/internal/app/handlers"
	"internal-work-management-service/internal/app/middleware"
	"internal-work-management-service/internal/infrastructure/logger"
)

const startupTimeout = 10 * time.Second

func run() int {
	logr := logger.Decorate(logger.New(logger.LogConfig{AppName: "app"}), "component", "app")
	middleware.SetLogger(logr)
	handlers.SetLogger(logger.Decorate(logr, "component", "handlers"))

	cfg, err := config.Load()
	if err != nil {
		logr.Error("msg", "config_load_failed", "error", err)
		return 1
	}
	logr.With(
		"http_addr", cfg.HTTPAddr,
		"read_timeout", cfg.HTTPReadTimeout.String(),
		"write_timeout", cfg.HTTPWriteTimeout.String(),
		"shutdown_timeout", cfg.HTTPShutdownTimeout.String(),
	).Info("msg", "config_loaded")

	startupCtx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()

	appHandler, closeApp, err := app.BuildApp(startupCtx, cfg)
	if err != nil {
		logr.Error("msg", "bootstrap_failed", "error", err)
		return 1
	}
	defer func() {
		if err := closeApp(); err != nil {
			logr.Error("msg", "cleanup_failed", "error", err)
		}
	}()

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           appHandler,
		ReadHeaderTimeout: cfg.HTTPReadTimeout,
		ReadTimeout:       cfg.HTTPReadTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		logr.Info("msg", "http_listen_start", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logr.Info("msg", "shutdown_signal_received", "signal", sig.String())
	case err := <-serverErr:
		if err != nil {
			logr.Error("msg", "server_failed", "error", err)
			return 1
		}
		logr.Info("msg", "server_stopped")
		return 0
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTPShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logr.Error("msg", "server_shutdown_failed", "error", err)
		return 1
	}

	if err := <-serverErr; err != nil {
		logr.Error("msg", "server_failed", "error", err)
		return 1
	}

	logr.Info("msg", "server_shutdown_complete")
	return 0
}

func main() {
	os.Exit(run())
}
