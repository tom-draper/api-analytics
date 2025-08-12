package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

type Logger struct {
	mu         sync.Mutex
	file       *os.File
	logger     *log.Logger
	buffer     chan string
	done       chan bool
	bufferSize int
}

var (
	defaultLogger *Logger
	once          sync.Once
)

const (
	defaultBufferSize = 1000
	flushInterval     = 1 * time.Second
	maxLogFileSize    = 100 * 1024 * 1024 // 100MB
)

// GetLogger returns the singleton logger instance
func GetLogger() *Logger {
	once.Do(func() {
		defaultLogger = NewLogger("./requests.log", defaultBufferSize)
	})
	return defaultLogger
}

// NewLogger creates a new logger with buffering
func NewLogger(filename string, bufferSize int) *Logger {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}

	logger := &Logger{
		file:       file,
		logger:     log.New(file, "", log.LstdFlags),
		buffer:     make(chan string, bufferSize),
		done:       make(chan bool),
		bufferSize: bufferSize,
	}

	// Start background writer
	go logger.writer()

	return logger
}

// writer handles buffered writes in background
func (l *Logger) writer() {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	batch := make([]string, 0, 100) // Batch up to 100 messages

	for {
		select {
		case msg := <-l.buffer:
			batch = append(batch, msg)
			
			// Flush if batch is full
			if len(batch) >= cap(batch) {
				l.flushBatch(batch)
				batch = batch[:0] // Reset slice but keep capacity
			}

		case <-ticker.C:
			// Flush any pending messages periodically
			if len(batch) > 0 {
				l.flushBatch(batch)
				batch = batch[:0]
			}
			
			// Check if log rotation is needed
			l.rotateIfNeeded()

		case <-l.done:
			// Flush remaining messages and exit
			if len(batch) > 0 {
				l.flushBatch(batch)
			}
			// Drain any remaining messages
			for {
				select {
				case msg := <-l.buffer:
					l.writeMessage(msg)
				default:
					return
				}
			}
		}
	}
}

// flushBatch writes multiple messages at once
func (l *Logger) flushBatch(messages []string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	for _, msg := range messages {
		l.logger.Println(msg)
	}
	
	// Force sync to disk for important logs
	if l.file != nil {
		l.file.Sync()
	}
}

// writeMessage writes a single message
func (l *Logger) writeMessage(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Println(msg)
}

// rotateIfNeeded rotates log file if it gets too large
func (l *Logger) rotateIfNeeded() {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.file == nil {
		return
	}
	
	stat, err := l.file.Stat()
	if err != nil {
		return
	}
	
	if stat.Size() > maxLogFileSize {
		// Rotate log file
		l.file.Close()
		
		// Move current log to backup
		os.Rename("./requests.log", fmt.Sprintf("./requests.log.%d", time.Now().Unix()))
		
		// Create new log file
		file, err := os.OpenFile("./requests.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("Error rotating log file: %v", err)
			return
		}
		
		l.file = file
		l.logger = log.New(file, "", log.LstdFlags)
	}
}

// logAsync sends message to buffer for async writing
func (l *Logger) logAsync(msg string) {
	select {
	case l.buffer <- msg:
		// Message queued successfully
	default:
		// Buffer is full, write synchronously to avoid blocking
		l.writeMessage(msg)
	}
}

// Close gracefully shuts down the logger
func (l *Logger) Close() {
	close(l.done)
	time.Sleep(100 * time.Millisecond) // Give writer time to finish
	
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.file != nil {
		l.file.Close()
		l.file = nil
	}
}

// SetOutput allows changing log destination (useful for testing)
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.SetOutput(w)
}

// Public API functions - these maintain backward compatibility

// LogErrorToFile logs error messages with context
func LogErrorToFile(ipAddress string, apiKey string, msg string) {
	text := fmt.Sprintf("%s %s :: %s", ipAddress, apiKey, msg)
	GetLogger().logAsync(text)
}

// LogRequestsToFile logs request statistics
func LogRequestsToFile(apiKey string, inserted int, totalRequests int) {
	text := fmt.Sprintf("key=%s: %d/%d", apiKey, inserted, totalRequests)
	GetLogger().logAsync(text)
}

// LogToFile logs a message asynchronously
func LogToFile(msg string) {
	GetLogger().logAsync(msg)
}

// LogToFileSync logs a message synchronously (use sparingly)
func LogToFileSync(msg string) {
	GetLogger().writeMessage(msg)
}

// Close gracefully shuts down the default logger
func Close() {
	if defaultLogger != nil {
		defaultLogger.Close()
	}
}

// SetLogOutput changes the log output destination
func SetLogOutput(w io.Writer) {
	GetLogger().SetOutput(w)
}