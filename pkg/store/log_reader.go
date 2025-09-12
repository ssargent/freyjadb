package store

import (
	"bufio"
	"io"
	"os"

	"github.com/ssargent/freyjadb/pkg/codec"
)

// LogReader provides sequential access to records in a log file
type LogReader struct {
	file   *os.File
	reader *bufio.Reader
	codec  *codec.RecordCodec
	offset int64
	config LogReaderConfig
}

// NewLogReader creates a new log reader for the specified file
func NewLogReader(config LogReaderConfig) (*LogReader, error) {
	file, err := os.Open(config.FilePath)
	if err != nil {
		return nil, err
	}

	// Seek to start offset if specified
	if config.StartOffset > 0 {
		if _, err := file.Seek(config.StartOffset, 0); err != nil {
			file.Close()
			return nil, err
		}
	}

	return &LogReader{
		file:   file,
		reader: bufio.NewReader(file),
		codec:  codec.NewRecordCodec(),
		offset: config.StartOffset,
		config: config,
	}, nil
}

// ReadNext reads the next record from the current offset
func (r *LogReader) ReadNext() (*codec.Record, error) {
	// Read the record header (20 bytes: CRC32 + KeySize + ValueSize + Timestamp)
	header := make([]byte, 20)
	n, err := io.ReadFull(r.reader, header)
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, io.EOF
		}
		return nil, err
	}
	r.offset += int64(n)

	// Decode header to get sizes
	if len(header) < 20 {
		return nil, ErrCorruption
	}

	keySize := int(uint32(header[4]) | uint32(header[5])<<8 | uint32(header[6])<<16 | uint32(header[7])<<24)
	valueSize := int(uint32(header[8]) | uint32(header[9])<<8 | uint32(header[10])<<16 | uint32(header[11])<<24)

	// Read key and value data
	dataSize := keySize + valueSize
	if dataSize == 0 {
		// This might be a tombstone or empty record
		record := &codec.Record{
			CRC32:     uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16 | uint32(header[3])<<24,
			KeySize:   uint32(keySize),
			ValueSize: uint32(valueSize),
			Timestamp: uint64(header[12]) | uint64(header[13])<<8 | uint64(header[14])<<16 | uint64(header[15])<<24 | uint64(header[16])<<32 | uint64(header[17])<<40 | uint64(header[18])<<48 | uint64(header[19])<<56,
			Key:       []byte{},
			Value:     []byte{},
		}
		return record, nil
	}

	data := make([]byte, dataSize)
	n, err = io.ReadFull(r.reader, data)
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, ErrCorruption
		}
		return nil, err
	}
	r.offset += int64(n)

	// Construct full record data for decoding
	fullData := make([]byte, 20+dataSize)
	copy(fullData[0:20], header)
	copy(fullData[20:], data)

	// Decode the complete record
	record, err := r.codec.Decode(fullData)
	if err != nil {
		return nil, err
	}

	// Validate CRC
	if err := record.Validate(); err != nil {
		return nil, ErrCorruption
	}

	return record, nil
}

// ReadAt reads a record at a specific offset
func (r *LogReader) ReadAt(offset int64) (*codec.Record, error) {
	// Always reopen the file to ensure we see the latest data
	if r.file != nil {
		r.file.Close()
	}

	file, err := os.Open(r.config.FilePath)
	if err != nil {
		return nil, err
	}

	// Seek to the specified offset
	if _, err := file.Seek(offset, 0); err != nil {
		file.Close()
		return nil, err
	}

	// Read the record header (20 bytes: CRC32 + KeySize + ValueSize + Timestamp)
	header := make([]byte, 20)
	n, err := file.Read(header)
	if err != nil {
		file.Close()
		if err == io.EOF || n < 20 {
			return nil, ErrCorruption
		}
		return nil, err
	}

	// Decode header to get sizes
	if len(header) < 20 {
		file.Close()
		return nil, ErrCorruption
	}

	keySize := int(uint32(header[4]) | uint32(header[5])<<8 | uint32(header[6])<<16 | uint32(header[7])<<24)
	valueSize := int(uint32(header[8]) | uint32(header[9])<<8 | uint32(header[10])<<16 | uint32(header[11])<<24)

	// Read key and value data
	dataSize := keySize + valueSize
	if dataSize == 0 {
		// This might be a tombstone or empty record
		file.Close()
		record := &codec.Record{
			CRC32:     uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16 | uint32(header[3])<<24,
			KeySize:   uint32(keySize),
			ValueSize: uint32(valueSize),
			Timestamp: uint64(header[12]) | uint64(header[13])<<8 | uint64(header[14])<<16 | uint64(header[15])<<24 | uint64(header[16])<<32 | uint64(header[17])<<40 | uint64(header[18])<<48 | uint64(header[19])<<56,
			Key:       []byte{},
			Value:     []byte{},
		}
		return record, nil
	}

	data := make([]byte, dataSize)
	n, err = file.Read(data)
	if err != nil {
		file.Close()
		if err == io.EOF || n < dataSize {
			return nil, ErrCorruption
		}
		return nil, err
	}

	file.Close()

	// Construct full record data for decoding
	fullData := make([]byte, 20+dataSize)
	copy(fullData[0:20], header)
	copy(fullData[20:], data)

	// Decode the complete record
	record, err := r.codec.Decode(fullData)
	if err != nil {
		return nil, err
	}

	// Validate CRC
	if err := record.Validate(); err != nil {
		return nil, ErrCorruption
	}

	return record, nil
}

// Seek sets the read offset
func (r *LogReader) Seek(offset int64) error {
	if _, err := r.file.Seek(offset, 0); err != nil {
		return err
	}

	r.reader = bufio.NewReader(r.file) // Recreate reader to clear buffer
	r.offset = offset
	return nil
}

// Offset returns the current read offset
func (r *LogReader) Offset() int64 {
	return r.offset
}

// Iterator returns a streaming iterator for records
func (r *LogReader) Iterator() RecordIterator {
	return &logRecordIterator{reader: r}
}

// Close closes the log reader
func (r *LogReader) Close() error {
	return r.file.Close()
}

// logRecordIterator implements RecordIterator for streaming access
type logRecordIterator struct {
	reader *LogReader
	record *codec.Record
	err    error
}

func (it *logRecordIterator) Next() bool {
	it.record, it.err = it.reader.ReadNext()
	return it.err == nil
}

func (it *logRecordIterator) Record() *codec.Record {
	return it.record
}

func (it *logRecordIterator) Close() error {
	// Don't close the underlying reader as it's owned by the caller
	return nil
}
