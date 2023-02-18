package internal

import (
	"log"
	"os"
)

type Logger struct {
	stdLogger *log.Logger
	errLogger *log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		log.New(os.Stdout, "", log.LstdFlags),
		log.New(os.Stderr, "ERROR", log.LstdFlags),
	}
}

func (l *Logger) Info(args ...any) {
	l.stdLogger.Println(args...)
}

func (l *Logger) Error(args ...any) {
	l.errLogger.Println(args...)
}

func (l *Logger) Fatal(args ...any) {
	l.errLogger.Fatalln(args...)
}
