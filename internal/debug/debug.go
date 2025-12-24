package debug

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

var (
	debugMode bool
	mu        sync.RWMutex
	logger    *log.Logger
	file      *os.File
)

// Init initializes debug mode
func Init(enabled bool, logFile string) error {
	mu.Lock()
	defer mu.Unlock()
	
	debugMode = enabled
	
	if enabled && logFile != "" {
		var err error
		file, err = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open debug log file: %w", err)
		}
		logger = log.New(file, "DEBUG: ", log.LstdFlags|log.Lmicroseconds)
	} else {
		logger = log.New(os.Stderr, "DEBUG: ", log.LstdFlags|log.Lmicroseconds)
	}
	
	return nil
}

// Enabled returns whether debug mode is enabled
func Enabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return debugMode
}

// Log logs a debug message if debug mode is enabled
func Log(msg string) {
	if !Enabled() {
		return
	}
	
	if logger != nil {
		logger.Println(msg)
	}
}

// Logf logs a formatted debug message
func Logf(format string, args ...interface{}) {
	if !Enabled() {
		return
	}
	
	if logger != nil {
		logger.Printf(format, args...)
	}
}

// LogKey logs a key event
func LogKey(key string, source string) {
	Logf("KEY: %s from %s at %v", key, source, time.Now())
}

// LogMsg logs a tea message
func LogMsg(msg interface{}, source string) {
	Logf("MSG: %T %+v from %s at %v", msg, msg, source, time.Now())
}

// LogState logs application state
func LogState(state string, details interface{}) {
	Logf("STATE: %s - %+v at %v", state, details, time.Now())
}

// LogError logs an error
func LogError(err error, source string) {
	if !Enabled() {
		return
	}
	
	if logger != nil {
		logger.Printf("ERROR in %s: %v at %v", source, err, time.Now())
	}
}

// Close closes the debug log file
func Close() {
	if file != nil {
		file.Close()
	}
}

// SetMode enables or disables debug mode
func SetMode(enabled bool) {
	mu.Lock()
	defer mu.Unlock()
	debugMode = enabled
}
