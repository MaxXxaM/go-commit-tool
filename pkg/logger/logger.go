package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

// Level определяет уровень логирования
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Logger представляет простой логгер для CLI приложения
type Logger struct {
	level  Level
	debug  *log.Logger
	info   *log.Logger
	warn   *log.Logger
	error  *log.Logger
}

// New создает новый логгер
func New(level Level, output io.Writer) *Logger {
	if output == nil {
		output = os.Stdout
	}

	return &Logger{
		level:  level,
		debug:  log.New(output, "[DEBUG] ", log.Ltime),
		info:   log.New(output, "[INFO]  ", log.Ltime),
		warn:   log.New(output, "[WARN]  ", log.Ltime),
		error:  log.New(output, "[ERROR] ", log.Ltime),
	}
}

// Debug логирует сообщение уровня Debug
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= LevelDebug {
		l.debug.Printf(format, v...)
	}
}

// Info логирует сообщение уровня Info
func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= LevelInfo {
		l.info.Printf(format, v...)
	}
}

// Warn логирует сообщение уровня Warn
func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= LevelWarn {
		l.warn.Printf(format, v...)
	}
}

// Error логирует сообщение уровня Error
func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= LevelError {
		l.error.Printf(format, v...)
	}
}

// Fatal логирует сообщение и завершает программу
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.error.Printf(format, v...)
	os.Exit(1)
}

// Success печатает сообщение об успехе (зеленый цвет для поддерживающих терминалов)
func (l *Logger) Success(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	fmt.Printf("\033[32m✓ %s\033[0m\n", msg)
}

// Warning печатает предупреждение (желтый цвет)
func (l *Logger) Warning(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	fmt.Printf("\033[33m⚠ %s\033[0m\n", msg)
}

// ErrorMsg печатает ошибку (красный цвет)
func (l *Logger) ErrorMsg(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	fmt.Printf("\033[31m✗ %s\033[0m\n", msg)
}

// Default возвращает логгер по умолчанию
var Default = New(LevelInfo, os.Stdout)
