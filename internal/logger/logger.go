package logger

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2/widget"
)

// Logger provides a simple logging mechanism that writes to a Fyne widget.
type Logger struct {
	logWidget *widget.Label
	messages  []string
}

// New creates a new Logger.
func New(logWidget *widget.Label) *Logger {
	return &Logger{
		logWidget: logWidget,
	}
}

// Info logs an informational message.
func (l *Logger) Info(msg string) {
	l.addMessage("INFO: " + msg)
}

// Error logs an error message.
func (l *Logger) Error(err error) {
	l.addMessage("ERROR: " + err.Error())
}

func (l *Logger) addMessage(msg string) {
	timestamp := time.Now().Format("15:04:05")
	l.messages = append(l.messages, fmt.Sprintf("[%s] %s", timestamp, msg))
	if len(l.messages) > 100 { // Keep only the last 100 messages
		l.messages = l.messages[len(l.messages)-100:]
	}
	l.logWidget.SetText(strings.Join(l.messages, "\n"))
}