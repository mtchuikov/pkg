package logtools

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

const (
	// the owner has read, write, and execute permissions, while the
	// group and others have read and execute permissions only
	defaulFilePerm = 0755

	// os.O_CREATE creates the file if it doesn't exist, os.O_WRONLY
	// opens the file for write-only access, and os.O_APPEND ensures
	// that data is always written at the end of the file
	defaulFileFlags = os.O_CREATE | os.O_WRONLY | os.O_APPEND

	// the maximum size of the log file in bytes, by the default it
	// equals to 64 MB
	defaulMaxFileSize = 64 * 1024 * 1024

	// indicates whether log files should be compressed using gzip,
	// when set to true, logs will be compressed before being saved to
	// the file
	defaulCompress = false

	// the maximum number of log entries that can be buffered before
	// the logs are flushed
	defaulMaxBatchSize = 32

	// the interval at which the log buffer is flushed to disk, helps
	// to ensure that logs are written periodically even if the batch
	// size is not reached
	defaulFlushInterval = 10 * time.Second
)

const (
	failedToOpenLogFile   = "failed to open log file: %v"
	failedToRenameLogFile = "failed to rename log file: %v"
	failedToGetFileStats  = "failed to get file stats: %v"
	failedToWriteLogFile  = "failed to write log file: %v"
	failedToFlushLogBuf   = "failed to flush log buffer: %v"
	failedToRotateLogFile = "failed to rotate log file: %v"
	failedToCloseLogFile  = "failed to close log file: %v"
)

var _ io.WriteCloser = (*FileWriter)(nil)

type FileWriter struct {
	mu sync.Mutex

	mode     os.FileMode
	flags    int
	file     *os.File
	maxSize  int64
	compress bool

	maxBatchSize int
	batchSize    int
	writer       *bufio.Writer

	flushTicker  *time.Ticker
	errorHandler func(error)
	done         chan struct{}
}

type FileWriterOption func(*FileWriter)

func WithFileWriterFileMode(mode int) FileWriterOption {
	return func(fw *FileWriter) {
		fw.mode = os.FileMode(mode)
	}
}

func WithFileWriterMaxSize(size int) FileWriterOption {
	return func(fw *FileWriter) {
		fw.maxSize = int64(size * 1024 * 1024)
	}
}

func WithFileWriterCompress(compress bool) FileWriterOption {
	return func(fw *FileWriter) {
		fw.compress = compress
	}
}

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

func NewFileWriter(file string, opts ...FileWriterOption) (*FileWriter, error) {
	fw := &FileWriter{
		mode:         defaulFilePerm,
		flags:        defaulFileFlags,
		maxSize:      defaulMaxFileSize,
		compress:     defaulCompress,
		maxBatchSize: defaulMaxBatchSize,
		flushTicker:  time.NewTicker(defaulFlushInterval),
		errorHandler: func(err error) {},
	}

	for _, opt := range opts {
		opt(fw)
	}

	f, err := os.OpenFile(file, fw.flags, fw.mode)
	if err != nil {
		err = errors.Unwrap(err)
		return nil, fmt.Errorf(failedToOpenLogFile, err)
	}

	fw.mu = sync.Mutex{}
	fw.file = f
	fw.batchSize = 0
	fw.writer = bufio.NewWriterSize(fw.file, fw.maxBatchSize)
	fw.done = make(chan struct{})

	return fw, nil
}

// Open opens a new log file with the specified name, using the
// flags and permissions set in the FileWriter. It resets the
// internal writer to work with the new file. Before calling
// this function, ensure that the Close method is called to
// properly close the previous log file and avoid resource
// leaks or data corruption
func (fw *FileWriter) Open(file string, mode int) error {
	m := os.FileMode(mode)
	f, err := os.OpenFile(file, fw.flags, m)
	if err != nil {
		err = errors.Unwrap(err)
		return fmt.Errorf(failedToOpenLogFile, err)
	}

	fw.mode = m
	fw.file = f
	fw.writer.Reset(fw.file)

	return nil
}

func (fw *FileWriter) closeFile() {
	fw.file.Close()
	fw.file = nil
}

// rotate performs log file rotation. It closes the current log
// file, renames it with a timestamp postfix, and opens a new
// one the original name
func (fw *FileWriter) rotate() error {
	fileName := fw.file.Name()
	fw.closeFile()

	postfix := time.Now().Format(time.RFC3339)
	backupName := fileName + "." + postfix

	err := os.Rename(fileName, backupName)
	if err != nil {
		err = errors.Unwrap(err)
		return fmt.Errorf(failedToRenameLogFile, err)
	}

	fw.file, err = os.OpenFile(fileName, fw.flags, fw.mode)
	if err != nil {
		err = errors.Unwrap(err)
		return fmt.Errorf(failedToOpenLogFile, err)
	}

	fw.writer.Reset(fw.file)

	return nil
}

// flushWriter ensures that all buffered log data is written to the
// file. If the rotate parameter is true, it first rotates the
// log file to manage file size and archive old logs. After
// optional rotation, it flushes the writer's buffer to write any
// remaining data to the current log file
func (fw *FileWriter) flush() error {
	err := fw.writer.Flush()
	if err != nil {
		err = errors.Unwrap(err)
		return fmt.Errorf(failedToFlushLogBuf, err)
	}

	return nil
}

// Write writes the provided data to the log file while ensuring
// that the total size of the file, the buffered data, and the
// new data does not exceed the maximum allowed size. If the new
// data would cause the size to surpass this limit, the log file
// is rotated and any buffered data is flushed before proceeding.
// After writing, if the number of batched entries reaches the
// predefined threshold, the buffer is flushed
func (fw *FileWriter) getFileSize() (int64, error) {
	stat, err := fw.file.Stat()
	if err != nil {
		err = errors.Unwrap(err)
		return 0, fmt.Errorf(failedToGetFileStats, err)
	}

	return stat.Size(), nil
}

func (fw *FileWriter) Write(p []byte) (int, error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	fileSize, err := fw.getFileSize()
	if err != nil {
		return 0, err
	}

	bufSize := int64(fw.writer.Buffered())
	pSize := int64(len(p))
	postWriteSize := fileSize + bufSize + pSize

	if postWriteSize > fw.maxSize {
		fw.batchSize = 0
		err := fw.rotate()
		if err != nil {
			return 0, err
		}

		err = fw.flush()
		if err != nil {
			return 0, err
		}
	}

	n, err := fw.writer.Write(p)
	if err != nil {
		err = errors.Unwrap(err)
		return n, fmt.Errorf(failedToWriteLogFile, err)
	}

	fw.batchSize++
	if fw.batchSize >= fw.maxBatchSize {
		fw.batchSize = 0
		// flush the buffer without rotating, because after the
		// postWriteSize calculation we assume that there will be enough
		// space after rotation, but in some cases (e.g., when an
		// excessively large number of bytes is passed), this assumption
		// might not hold; it is the user's responsibility to ensure that
		// the input size remains within acceptable limits
		err = fw.flush()
	}

	return n, err
}

func (fw *FileWriter) Rotate() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	fileSize, err := fw.getFileSize()
	if err != nil {
		return err
	}

	bufSize := int64(fw.writer.Buffered())
	postWriteSize := fileSize + bufSize
	rotateBeforeFlush := postWriteSize > fw.maxSize

	if rotateBeforeFlush {
		err = fw.rotate()
		if err != nil {
			return err
		}

		err = fw.flush()
	} else {
		err = fw.flush()
		if err != nil {
			return err
		}

		err = fw.rotate()
	}

	return err
}

func (fw *FileWriter) Close() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	fw.flushTicker.Stop()
	close(fw.done)

	fileSize, err := fw.getFileSize()
	if err != nil {
		return err
	}

	bufSize := int64(fw.writer.Buffered())
	rotateBeforeFlush := fileSize+bufSize > fw.maxSize

	if rotateBeforeFlush {
		err = fw.rotate()
	}

	if err == nil {
		fw.flush()
	}

	fw.closeFile()

	return nil
}
