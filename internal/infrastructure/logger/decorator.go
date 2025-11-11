package logger

// Decorate wraps a logger with additional key/value pairs appended to each log entry
func Decorate(base Logger, extras ...any) Logger {
	return &decorator{base: base, extras: extras}
}

type decorator struct {
	base   Logger
	extras []any
}

func (d *decorator) With(args ...any) Logger {
	combined := append(append([]any{}, d.extras...), args...)
	return &decorator{
		base:   d.base,
		extras: combined,
	}
}

func (d *decorator) Info(args ...any) {
	combined := append(append([]any{}, d.extras...), args...)
	d.base.Info(combined...)
}

func (d *decorator) Error(args ...any) {
	combined := append(append([]any{}, d.extras...), args...)
	d.base.Error(combined...)
}
