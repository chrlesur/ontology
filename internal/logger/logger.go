package logger

type Logger struct{}

func GetLogger() *Logger {
    return &Logger{}
}

func (l *Logger) Debug(msg string) {}
func (l *Logger) Info(msg string)  {}
func (l *Logger) Warning(msg string, err error) {}
func (l *Logger) Error(msg string, err error)   {}