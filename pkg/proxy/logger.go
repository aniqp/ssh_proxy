package proxy

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aniqp/formal_assessment/pkg/config"
)

type Logger struct {
	cfg *config.Config
}

func NewLogger(cfg *config.Config) *Logger {
	return &Logger{cfg: cfg}
}

func (l *Logger) createLogFile(username string, timestamp time.Time, logDir string) (*os.File, error) {
	if err := os.MkdirAll(logDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	timestampString := timestamp.Format("2006-01-02 15:04:05")
	timestampStringReplace := strings.ReplaceAll(timestampString, ":", "_")

	filename := fmt.Sprintf("session_%s_%s.txt", username, timestampStringReplace)
	filePath := filepath.Join(logDir, filename)

	logFile, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %v", err)
	}

	_, err = logFile.WriteString(fmt.Sprintf("user: %s\ntimestamp: %s\n\n", username, timestampString))
	if err != nil {
		return nil, fmt.Errorf("failed to write metadata to log file: %v", err)
	}

	log.Printf("Logging stdin to: %s", filePath)
	return logFile, nil
}
