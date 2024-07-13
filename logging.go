package main

import (
	"fmt"
	"log"
)

const (
	LevelTrace   = "TRACE"
	LevelDebug   = "DEBUG"
	LevelInfo    = "INFO"
	LevelWarning = "WARN"
	LevelError   = "ERROR"
)

func logF(level string, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	colorized := colorize(fmt.Sprintf("[%s] - %s", level, msg), level)
	log.Println(colorized)
}

func colorize(s string, level string) string {
	var levelColor int

	// Pick a color for the log level
	switch level {
	case LevelTrace:
		levelColor = 36 // cyan
	case LevelDebug:
		levelColor = 37 // gray
	case LevelInfo:
		levelColor = 34 // blue
	case LevelWarning:
		levelColor = 33 // yellow
	case LevelError:
		levelColor = 31 // red
	default:
		levelColor = 0
	}

	return fmt.Sprintf("\u001B[%dm%s\u001B[0m", levelColor, s)
}
