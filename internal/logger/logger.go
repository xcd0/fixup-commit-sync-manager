package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var logLevelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
}

var logLevelColors = map[LogLevel]string{
	DEBUG: "\033[36m", // Cyan
	INFO:  "\033[32m", // Green
	WARN:  "\033[33m", // Yellow
	ERROR: "\033[31m", // Red
}

const colorReset = "\033[0m"

type Logger struct {
	level      LogLevel
	fileLogger *log.Logger
	verbose    bool
	logFile    *os.File
}

func NewLogger(levelStr, filePath string, verbose bool) (*Logger, error) {
	level := parseLogLevel(levelStr)

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	fileLogger := log.New(logFile, "", log.LstdFlags)

	return &Logger{
		level:      level,
		fileLogger: fileLogger,
		verbose:    verbose,
		logFile:    logFile,
	}, nil
}

func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelName := logLevelNames[level]
	message := fmt.Sprintf(format, args...)

	logEntry := fmt.Sprintf("[%s] [%s] %s", timestamp, levelName, message)

	if l.fileLogger != nil {
		l.fileLogger.Println(logEntry)
	}

	if l.verbose {
		color := logLevelColors[level]
		coloredLevel := color + levelName + colorReset
		coloredEntry := fmt.Sprintf("[%s] [%s] %s", timestamp, coloredLevel, message)
		fmt.Println(coloredEntry)
	}
}

func (l *Logger) LogOperationStart(operation string) {
	l.Info("Starting operation: %s", operation)
}

func (l *Logger) LogOperationEnd(operation string, duration time.Duration) {
	l.Info("Completed operation: %s (duration: %v)", operation, duration)
}

func (l *Logger) LogOperationError(operation string, err error) {
	l.Error("Operation failed: %s - %v", operation, err)
}

func (l *Logger) LogFileOperation(operation, filePath string) {
	l.Debug("File operation: %s - %s", operation, filePath)
}

func (l *Logger) LogGitOperation(operation string, args []string) {
	l.Debug("Git operation: %s %s", operation, strings.Join(args, " "))
}

func (l *Logger) LogConfigLoad(configPath string) {
	l.Info("Configuration loaded from: %s", configPath)
}

func (l *Logger) LogSyncResult(filesAdded, filesModified, filesDeleted int, commitHash string) {
	if filesAdded+filesModified+filesDeleted == 0 {
		l.Info("Sync completed - no changes detected")
	} else {
		l.Info("Sync completed - Files: +%d ~%d -%d, Commit: %s",
			filesAdded, filesModified, filesDeleted, commitHash[:8])
	}
}

func (l *Logger) LogFixupResult(filesModified int, commitHash string) {
	if filesModified == 0 {
		l.Info("Fixup completed - no changes to commit")
	} else {
		l.Info("Fixup completed - %d files modified, Commit: %s",
			filesModified, commitHash[:8])
	}
}

func (l *Logger) LogVHDXOperation(operation, vhdxPath string) {
	l.Info("VHDX operation: %s - %s", operation, vhdxPath)
}

func parseLogLevel(levelStr string) LogLevel {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

func SetupGlobalLogger(levelStr, filePath string, verbose bool) (*Logger, error) {
	return NewLogger(levelStr, filePath, verbose)
}

type MultiWriter struct {
	writers []io.Writer
}

func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{writers: writers}
}

func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return
		}
	}
	return len(p), nil
}
