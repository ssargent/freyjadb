package store

import (
	"bufio"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ssargent/freyjadb/pkg/codec"
)

// LogWriter handles append-only writes to the active data file
type LogWriter struct {
	file       *os.File
	writer     *bufio.Writer
	codec      *codec.RecordCodec
	fsyncTimer *time.Timer
	config     LogWriterConfig
	mutex      sync.Mutex
	offset     int64 // Current write offset
}

// NewLogWriter creates a new log writer with the given configuration
func NewLogWriter(config LogWriterConfig) (*LogWriter, error) {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(config.FilePath), 0750); err != nil {
		return nil, err
	}

	// Open file in write-only mode, create if doesn't exist
	file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}

	// Seek to end for append behavior
	if _, err := file.Seek(0, 2); err != nil {
		if closeErr := file.Close(); closeErr != nil {
			// Log or handle
		}
		return nil, err
	}

	// Get current file size for offset tracking
	stat, err := file.Stat()
	if err != nil {
		if closeErr := file.Close(); closeErr != nil {
			// Log or handle
		}
		return nil, err
	}

	writer := &LogWriter{
		file:   file,
		writer: bufio.NewWriterSize(file, config.BufferSize),
		codec:  codec.NewRecordCodec(),
		config: config,
		offset: stat.Size(),
	}

	// Set up fsync timer if interval is configured
	if config.FsyncInterval > 0 {
		writer.fsyncTimer = time.AfterFunc(config.FsyncInterval, func() {
			writer.mutex.Lock()
			defer writer.mutex.Unlock()
			writer.sync() // Ignore error in timer callback
		})
	}

	return writer, nil
}

// Put appends a key-value pair to the log file and returns the record offset
func (w *LogWriter) Put(key, value []byte) (int64, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Encode the record
	data, err := w.codec.Encode(key, value)
	if err != nil {
		return 0, err
	}

	// Write to buffer
	n, err := w.writer.Write(data)
	if err != nil {
		return 0, err
	}

	// Calculate the offset where this record starts
	recordOffset := w.offset

	// Update offset
	w.offset += int64(n)

	// Sync immediately if no fsync interval configured
	if w.config.FsyncInterval == 0 {
		if err := w.sync(); err != nil {
			return 0, err
		}
	} else {
		// Reset fsync timer
		if w.fsyncTimer != nil {
			w.fsyncTimer.Reset(w.config.FsyncInterval)
		}
	}

	return recordOffset, nil
}

// Sync forces a fsync to disk
func (w *LogWriter) Sync() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.sync()
}

// sync performs the actual fsync operation (internal method)
func (w *LogWriter) sync() error {
	// Flush buffered writes
	if err := w.writer.Flush(); err != nil {
		return err
	}

	// Fsync to disk
	return w.file.Sync()
}

// Close closes the log writer and ensures all data is synced
func (w *LogWriter) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Cancel fsync timer
	if w.fsyncTimer != nil {
		w.fsyncTimer.Stop()
	}

	// Final sync
	if err := w.sync(); err != nil {
		if closeErr := w.file.Close(); closeErr != nil {
			// Log or handle
		}
		return err
	}

	return w.file.Close()
}

// Size returns the current size of the log file
func (w *LogWriter) Size() int64 {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.offset
}

// Path returns the file path
func (w *LogWriter) Path() string {
	return w.config.FilePath
}
