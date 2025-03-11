package logger

import (
	"io"
	"log"
	"os"
)

type LogLevel int

const (
	LevelError LogLevel = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

var (
	// Default loggers
	ErrorLogger *log.Logger
	WarnLogger  *log.Logger
	InfoLogger  *log.Logger
	DebugLogger *log.Logger

	currentLevel LogLevel = LevelInfo
)

func Initialize() {
	flags := log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile

	ErrorLogger = log.New(os.Stderr, "[ERROR] ", flags)
	WarnLogger = log.New(os.Stdout, "[WARN]  ", flags)
	InfoLogger = log.New(os.Stdout, "[INFO]  ", flags)
	DebugLogger = log.New(os.Stdout, "[DEBUG] ", flags)

	log.SetFlags(flags)
	log.SetPrefix("[INFO]  ")

	InfoLogger.Println("Logger initialized")
}

func SetLevel(level LogLevel) {
	currentLevel = level
	InfoLogger.Printf("Log level set to %v", level)
}

func SetOutput(w io.Writer) {
	ErrorLogger.SetOutput(w)
	WarnLogger.SetOutput(w)
	InfoLogger.SetOutput(w)
	DebugLogger.SetOutput(w)
	log.SetOutput(w)
}

func Error(format string, v ...interface{}) {
	ErrorLogger.Printf(format, v...)
}

func Warn(format string, v ...interface{}) {
	if currentLevel >= LevelWarn {
		WarnLogger.Printf(format, v...)
	}
}

func Info(format string, v ...interface{}) {
	if currentLevel >= LevelInfo {
		InfoLogger.Printf(format, v...)
	}
}

func Debug(format string, v ...interface{}) {
	if currentLevel >= LevelDebug {
		DebugLogger.Printf(format, v...)
	}
}

func Fatal(format string, v ...interface{}) {
	ErrorLogger.Fatalf(format, v...)
}
