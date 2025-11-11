package handlers

import "internal-work-management-service/internal/infrastructure/logger"

var handlerLogger logger.Logger = logger.Decorate(
	logger.New(logger.LogConfig{AppName: "handlers"}),
	"component", "handlers",
)

// SetLogger overrides the default structured logger for handlers
func SetLogger(l logger.Logger) {
	if l == nil {
		return
	}
	handlerLogger = l
}
