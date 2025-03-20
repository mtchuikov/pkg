package logtools

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// TODO:
// 1. Add file rotation when the maxSizeMB is exceeded.
// 2. Compress old log files when creating a new one.
// 3. Add error handling.

const (
	defaulFileMode      = 0644
	defaulFileFlags     = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	defaulMaxSizeMB     = 512
	defaulCompress      = true
	defaulMaxBatchSize  = 30
	defaulFlushInterval = 5 * time.Second
)

var _ io.WriteCloser = (*FileWriter)(nil)

type FileWriter struct {
	mode os.FileMode
	file *os.File
	// maxSizeMB int
	// compress     bool
	maxBatchSize int
	batchSize    int
	writer       *bufio.Writer
	mu           sync.Mutex
	flushTicker  *time.Ticker
	done         chan struct{}
}

type FileWriterOption func(*FileWriter)

func WithFileWriterFileMode(mode int) FileWriterOption {
	return func(fw *FileWriter) {
		fw.mode = os.FileMode(mode)
	}
}

// func WithFileWriterMaxSizeMB(size int) FileWriterOption {
// 	return func(fw *FileWriter) {
// 		fw.maxSizeMB = size
// 	}
// }

// func WithFileWriterCompress(compress bool) FileWriterOption {
// 	return func(fw *FileWriter) {
// 		fw.compress = compress
// 	}
// }

func WithFileWriterFlushInterval(interval time.Duration) FileWriterOption {
	return func(fw *FileWriter) {
		fw.flushTicker = time.NewTicker(interval)
	}
}

func WithFileWriterMaxBatchSize(size int) FileWriterOption {
	return func(fw *FileWriter) {
		fw.maxBatchSize = size
	}
}

func (fw *FileWriter) runTicker() {
	for {
		select {
		case <-fw.flushTicker.C:
			fw.mu.Lock()
			fw.writer.Flush()
			fw.mu.Unlock()
		case <-fw.done:
			return
		}
	}
}

func NewFileWriter(file string, opts ...FileWriterOption) (*FileWriter, error) {
	fw := &FileWriter{
		mode: defaulFileMode,
		// maxSizeMB:    fileWriterMaxSizeMB,
		// compress:     fileWriterCompress,
		maxBatchSize: defaulMaxBatchSize,
		flushTicker:  time.NewTicker(defaulFlushInterval),
	}

	for _, opt := range opts {
		opt(fw)
	}

	f, err := os.OpenFile(file, defaulFileFlags, fw.mode)
	if err != nil {
		err = fmt.Errorf("failed to open log file: %w", err)
		return nil, err
	}

	fw.file = f
	fw.batchSize = 0
	fw.writer = bufio.NewWriter(fw.file)
	fw.mu = sync.Mutex{}
	fw.done = make(chan struct{})

	go fw.runTicker()

	return fw, nil
}

func (fw *FileWriter) Write(p []byte) (int, error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	n, err := fw.writer.Write(p)
	if err != nil {
		err = fmt.Errorf("failed to write to log file: %w", err)
		return n, err
	}

	fw.batchSize++
	if fw.batchSize >= fw.maxBatchSize {
		err = fw.writer.Flush()
		fw.batchSize = 0
		if err != nil {
			err = fmt.Errorf("failed to flush log buffer: %w", err)
			return n, err
		}
	}

	return n, nil
}

func (fw *FileWriter) Close() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	fw.flushTicker.Stop()
	close(fw.done)

	err := fw.writer.Flush()
	if err != nil {
		err = fmt.Errorf("failed to flush log buffer: %w", err)
		return err
	}

	err = fw.file.Close()
	return fmt.Errorf("failed to close log file: %w", err)
}
