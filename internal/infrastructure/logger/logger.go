package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// Logger defines the minimal contract for logging everything
type Logger interface {
	With(args ...any) Logger
	Info(args ...any)
	Error(args ...any)
}

// LogConfig configures basic fields for the logger
type LogConfig struct {
	AppName string
	Stdout  *os.File
}

// New constructs logger
func New(cfg LogConfig) Logger {
	stdout := cfg.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	return &stdLogger{
		base: map[string]any{
			"app": cfg.AppName,
		},
		log: log.New(stdout, "", 0),
	}
}

type stdLogger struct {
	mu   sync.Mutex
	log  *log.Logger
	base map[string]any
}

func (l *stdLogger) With(args ...any) Logger {
	child := make(map[string]any, len(l.base)+len(args)/2)
	for k, v := range l.base {
		child[k] = v
	}
	merge(child, args)
	return &stdLogger{
		mu:   sync.Mutex{},
		log:  l.log,
		base: child,
	}
}

func (l *stdLogger) Info(args ...any) {
	l.write("INFO", args...)
}

func (l *stdLogger) Error(args ...any) {
	l.write("ERROR", args...)
}

func (l *stdLogger) write(level string, args ...any) {
	entry := make(map[string]any, len(l.base)+len(args)/2+2)
	for k, v := range l.base {
		entry[k] = v
	}
	entry["level"] = level
	entry["ts"] = time.Now().UTC().Format(time.RFC3339Nano)
	merge(entry, args)
	l.mu.Lock()
	defer l.mu.Unlock()
	l.log.Println(format(entry))
}

func merge(target map[string]any, args []any) {
	for i := 0; i+1 < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			continue
		}
		target[key] = args[i+1]
	}
}

func format(fields map[string]any) string {
	parts := make([]string, 0, len(fields))
	for k, v := range fields {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return strings.Join(parts, " ")
}
