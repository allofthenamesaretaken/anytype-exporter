package utils

import (
	"log"
	"os"
)

type Logger struct {
	infoStdoutLogger  *log.Logger
	infoFileLogger    *log.Logger
	warnStdoutLogger  *log.Logger
	warnFileLogger    *log.Logger
	errorStdoutLogger *log.Logger
	errorFileLogger   *log.Logger
	file              *os.File
}

func NewLogger(filePath string) (*Logger, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		infoStdoutLogger: log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		infoFileLogger:   log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),

		warnStdoutLogger: log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
		warnFileLogger:   log.New(file, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),

		errorStdoutLogger: log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorFileLogger:   log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),

		file: file,
	}, nil
}

func (logger *Logger) Info(msg string) {
	logger.infoStdoutLogger.Printf("%s", msg)
	logger.infoFileLogger.Printf("%s", msg)
}

func (logger *Logger) Warn(msg string) {
	logger.warnStdoutLogger.Printf("%s", msg)
	logger.warnFileLogger.Printf("%s", msg)
}

func (logger *Logger) Error(msg string, err error) {
	logger.errorStdoutLogger.Printf("%s: %v", msg, err)
	logger.errorFileLogger.Printf("%s: %v", msg, err)
}

func (logger *Logger) Close() {
	logger.file.Close()
}
