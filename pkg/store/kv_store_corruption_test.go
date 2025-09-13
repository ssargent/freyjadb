package store

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestDataCorruptionScenarios tests various scenarios that can cause data corruption
func TestDataCorruptionScenarios(t *testing.T) {
	t.Run("ImmediateReadAfterWrite", func(t *testing.T) {
		testImmediateReadAfterWrite(t)
	})

	t.Run("ConcurrentReadWrite", func(t *testing.T) {
		testConcurrentReadWrite(t)
	})

	t.Run("BufferStateConsistency", func(t *testing.T) {
		testBufferStateConsistency(t)
	})

	t.Run("OffsetCalculationAccuracy", func(t *testing.T) {
		testOffsetCalculationAccuracy(t)
	})

	t.Run("FileHandleSynchronization", func(t *testing.T) {
		testFileHandleSynchronization(t)
	})

	t.Run("ReaderWriterRaceCondition", func(t *testing.T) {
		testReaderWriterRaceCondition(t)
	})
}

// testImmediateReadAfterWrite tests reading data immediately after writing
func testImmediateReadAfterWrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kv_corruption_immediate")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: 0, // Immediate sync
	}

	store, err := NewKVStore(config)
	assert.NoError(t, err)

	_, err = store.Open()
	assert.NoError(t, err)
	defer store.Close()

	// Test multiple immediate read-after-write scenarios
	testCases := []struct {
		key   string
		value string
		desc  string
	}{
		{"key1", "value1", "simple key-value"},
		{"key2", "a_longer_value_that_might_cause_buffer_issues", "long value"},
		{"key3", " ", "space value (not empty)"},
		{"key4", string(make([]byte, 1024)), "large value"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Write data
			err := store.Put([]byte(tc.key), []byte(tc.value))
			assert.NoError(t, err, "Failed to put %s", tc.desc)

			// Immediately read data
			readValue, err := store.Get([]byte(tc.key))
			assert.NoError(t, err, "Failed to get %s immediately after write", tc.desc)
			assert.Equal(t, tc.value, string(readValue), "Value mismatch for %s", tc.desc)
		})
	}
}

// testConcurrentReadWrite tests concurrent read/write operations
func testConcurrentReadWrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kv_corruption_concurrent")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: time.Millisecond * 10, // Short fsync interval
	}

	store, err := NewKVStore(config)
	assert.NoError(t, err)

	_, err = store.Open()
	assert.NoError(t, err)
	defer store.Close()

	const numGoroutines = 10
	const numOperations = 50

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperations)

	// Start concurrent readers and writers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", goroutineID, j)
				value := fmt.Sprintf("value_%d_%d", goroutineID, j)

				// Write operation
				err := store.Put([]byte(key), []byte(value))
				if err != nil {
					errors <- fmt.Errorf("write error goroutine %d op %d: %v", goroutineID, j, err)
					continue
				}

				// Immediate read operation
				readValue, err := store.Get([]byte(key))
				if err != nil {
					errors <- fmt.Errorf("read error goroutine %d op %d: %v", goroutineID, j, err)
					continue
				}

				if string(readValue) != value {
					errors <- fmt.Errorf("data corruption goroutine %d op %d: expected %s, got %s",
						goroutineID, j, value, string(readValue))
					continue
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	var errorCount int
	for err := range errors {
		t.Errorf("Concurrent operation error: %v", err)
		errorCount++
	}

	assert.Equal(t, 0, errorCount, "Found %d errors in concurrent operations", errorCount)
}

// testBufferStateConsistency tests that buffer state remains consistent
func testBufferStateConsistency(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kv_corruption_buffer")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: time.Millisecond * 50, // Moderate fsync interval
	}

	store, err := NewKVStore(config)
	assert.NoError(t, err)

	_, err = store.Open()
	assert.NoError(t, err)
	defer store.Close()

	// Write a series of records
	const numRecords = 100
	keys := make([]string, numRecords)
	values := make([]string, numRecords)

	for i := 0; i < numRecords; i++ {
		keys[i] = fmt.Sprintf("buffer_test_key_%d", i)
		values[i] = fmt.Sprintf("buffer_test_value_%d_with_some_extra_data_to_make_it_longer", i)

		err := store.Put([]byte(keys[i]), []byte(values[i]))
		assert.NoError(t, err, "Failed to write record %d", i)
	}

	// Force sync to ensure all data is on disk
	err = store.writer.Sync()
	assert.NoError(t, err)

	// Read all records back
	for i := 0; i < numRecords; i++ {
		readValue, err := store.Get([]byte(keys[i]))
		assert.NoError(t, err, "Failed to read record %d after sync", i)
		assert.Equal(t, values[i], string(readValue), "Buffer state corruption in record %d", i)
	}

	// Test reading in reverse order to check for buffer state issues
	for i := numRecords - 1; i >= 0; i-- {
		readValue, err := store.Get([]byte(keys[i]))
		assert.NoError(t, err, "Failed to read record %d in reverse", i)
		assert.Equal(t, values[i], string(readValue), "Reverse read corruption in record %d", i)
	}
}

// testOffsetCalculationAccuracy tests that offset calculations are accurate
func testOffsetCalculationAccuracy(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kv_corruption_offset")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: 0, // Immediate sync
	}

	store, err := NewKVStore(config)
	assert.NoError(t, err)

	_, err = store.Open()
	assert.NoError(t, err)
	defer store.Close()

	// Test with various data sizes to check offset calculations
	testSizes := []int{1, 10, 100, 1000, 10000}

	for _, size := range testSizes {
		t.Run(fmt.Sprintf("Size_%d", size), func(t *testing.T) {
			key := fmt.Sprintf("offset_test_%d", size)
			value := string(make([]byte, size))

			// Write data
			err := store.Put([]byte(key), []byte(value))
			assert.NoError(t, err, "Failed to write %d byte value", size)

			// Read data back
			readValue, err := store.Get([]byte(key))
			assert.NoError(t, err, "Failed to read %d byte value", size)
			assert.Equal(t, size, len(readValue), "Size mismatch for %d byte value", size)
			assert.Equal(t, value, string(readValue), "Content mismatch for %d byte value", size)
		})
	}
}

// testFileHandleSynchronization tests file handle management
func testFileHandleSynchronization(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kv_corruption_handles")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: time.Millisecond * 100,
	}

	store, err := NewKVStore(config)
	assert.NoError(t, err)

	_, err = store.Open()
	assert.NoError(t, err)
	defer store.Close()

	// Test multiple open/close cycles
	for cycle := 0; cycle < 5; cycle++ {
		t.Run(fmt.Sprintf("Cycle_%d", cycle), func(t *testing.T) {
			// Write data
			key := fmt.Sprintf("handle_test_%d", cycle)
			value := fmt.Sprintf("handle_test_value_%d", cycle)

			err := store.Put([]byte(key), []byte(value))
			assert.NoError(t, err, "Failed to write in cycle %d", cycle)

			// Read data
			readValue, err := store.Get([]byte(key))
			assert.NoError(t, err, "Failed to read in cycle %d", cycle)
			assert.Equal(t, value, string(readValue), "Value mismatch in cycle %d", cycle)

			// Force sync
			err = store.writer.Sync()
			assert.NoError(t, err, "Failed to sync in cycle %d", cycle)
		})
	}
}

// testReaderWriterRaceCondition tests for race conditions between reader and writer
func testReaderWriterRaceCondition(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "kv_corruption_race")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	config := KVStoreConfig{
		DataDir:       tmpDir,
		FsyncInterval: time.Millisecond * 5, // Very short fsync interval
	}

	store, err := NewKVStore(config)
	assert.NoError(t, err)

	_, err = store.Open()
	assert.NoError(t, err)
	defer store.Close()

	const numOperations = 100
	done := make(chan bool, 2)

	// Writer goroutine
	go func() {
		defer func() { done <- true }()

		for i := 0; i < numOperations; i++ {
			key := fmt.Sprintf("race_key_%d", i)
			value := fmt.Sprintf("race_value_%d_with_extra_data_to_increase_size", i)

			err := store.Put([]byte(key), []byte(value))
			if err != nil {
				t.Errorf("Write error in operation %d: %v", i, err)
				return
			}

			// Small delay to allow reader to potentially race
			time.Sleep(time.Microsecond * 10)
		}
	}()

	// Reader goroutine
	go func() {
		defer func() { done <- true }()

		for i := 0; i < numOperations; i++ {
			key := fmt.Sprintf("race_key_%d", i)

			// Try to read (might fail if not written yet, that's OK)
			readValue, err := store.Get([]byte(key))
			if err != nil {
				// Expected if not written yet or data corruption
				if err != ErrKeyNotFound {
					t.Errorf("Read error in operation %d: %v", i, err)
				}
			} else {
				expectedValue := fmt.Sprintf("race_value_%d_with_extra_data_to_increase_size", i)
				if string(readValue) != expectedValue {
					t.Errorf("Race condition corruption in operation %d: expected %s, got %s",
						i, expectedValue, string(readValue))
				}
			}

			// Small delay to allow writer to potentially race
			time.Sleep(time.Microsecond * 10)
		}
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Final verification - read all values
	for i := 0; i < numOperations; i++ {
		key := fmt.Sprintf("race_key_%d", i)
		expectedValue := fmt.Sprintf("race_value_%d_with_extra_data_to_increase_size", i)

		readValue, err := store.Get([]byte(key))
		assert.NoError(t, err, "Final read failed for key %d", i)
		assert.Equal(t, expectedValue, string(readValue), "Final value mismatch for key %d", i)
	}
}
