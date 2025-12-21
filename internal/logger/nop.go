package logger

type NopLogger struct{}

func (NopLogger) Info(msg string, fields ...Field)  {}
func (NopLogger) Error(msg string, fields ...Field) {}
func (NopLogger) Warn(msg string, fields ...Field)  {}
func (NopLogger) Debug(msg string, fields ...Field) {}
func (NopLogger) Fatal(msg string, fields ...Field) {}
func (NopLogger) Sync() error                       { return nil }
