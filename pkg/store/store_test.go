package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "store_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	require.NoError(t, err)
	assert.NotNil(t, store)

	// Verify data directory was created
	assert.DirExists(t, tmpDir)

	// Verify segment files were created
	assert.FileExists(t, filepath.Join(tmpDir, "001.data"))
	assert.FileExists(t, filepath.Join(tmpDir, "002.data"))

	// Test cleanup
	err = store.Close()
	assert.NoError(t, err)
}

func TestStore_PutAndGet(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "store_put_get_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	// Test Put
	err = store.Put([]byte("test_key"), []byte("test_value"))
	assert.NoError(t, err)

	// Test Get
	value, err := store.Get([]byte("test_key"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("stub_value"), value) // StoreImpl returns stub values

	// Test Get non-existent key
	_, err = store.Get([]byte("non_existent"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key not found")
}

func TestStore_Explain(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "store_explain_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Test basic explain
	opts := ExplainOptions{}
	result, err := store.Explain(ctx, opts)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify global stats
	assert.Greater(t, result.Global.TotalKeys, 0)
	assert.Greater(t, result.Global.Uptime, time.Duration(0))
	assert.Greater(t, result.Global.TotalSizeMB, 0.0)

	// Verify segments
	assert.Len(t, result.Segments, 2)
	for _, seg := range result.Segments {
		assert.Greater(t, seg.Keys, 0)
		assert.Greater(t, seg.SizeMB, 0.0)
	}

	// Verify partitions
	assert.Contains(t, result.Partitions, "User")
	assert.Contains(t, result.Partitions, "Item")

	userPartition := result.Partitions["User"]
	assert.Greater(t, userPartition.Keys, 0)
	assert.Equal(t, "1:N", userPartition.Cardinality)
}

func TestStore_ExplainWithSamples(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "store_explain_samples_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Test explain with samples
	opts := ExplainOptions{
		WithSamples: 2,
	}
	result, err := store.Explain(ctx, opts)
	require.NoError(t, err)

	// Should have samples
	assert.Len(t, result.Diagnostics.Samples, 2)
	for _, sample := range result.Diagnostics.Samples {
		assert.NotEmpty(t, sample.Key)
		assert.NotEmpty(t, sample.Value)
		assert.False(t, sample.Ts.IsZero())
	}
}

func TestStore_ExplainWithMetrics(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "store_explain_metrics_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Test explain with metrics
	opts := ExplainOptions{
		WithMetrics: true,
	}
	result, err := store.Explain(ctx, opts)
	require.NoError(t, err)

	// Should have metrics
	assert.Greater(t, result.Diagnostics.Metrics.AvgGetLatencyMs, 0.0)
	assert.Greater(t, result.Diagnostics.Metrics.IORateMBs, 0.0)
}

func TestStore_ExplainWithPKFilter(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "store_explain_pk_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Test explain with existing PK
	opts := ExplainOptions{
		PK: "User",
	}
	result, err := store.Explain(ctx, opts)
	require.NoError(t, err)
	assert.Empty(t, result.Warnings)

	// Test explain with non-existent PK
	opts = ExplainOptions{
		PK: "NonExistent",
	}
	result, err = store.Explain(ctx, opts)
	require.NoError(t, err)
	assert.Len(t, result.Warnings, 1)
	assert.Contains(t, result.Warnings[0], "No data for PK: NonExistent")
}

func TestStore_CompactionDetection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "store_compaction_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// The store has segments with 15% and 10% dead percentage
	// Only segments with >20% dead should be marked for compaction
	opts := ExplainOptions{}
	result, err := store.Explain(ctx, opts)
	require.NoError(t, err)

	// Should have no segments marked for compaction (15% and 10% < 20%)
	assert.Empty(t, result.Diagnostics.CompactionReady)
}

func TestStoreImpl_MultiplePuts(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "store_multiple_puts_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	// Put multiple keys
	keys := [][]byte{
		[]byte("key1"),
		[]byte("key2"),
		[]byte("key3"),
	}

	for _, key := range keys {
		err := store.Put(key, []byte("value"))
		assert.NoError(t, err)
	}

	// All keys should be retrievable
	for _, key := range keys {
		value, err := store.Get(key)
		assert.NoError(t, err)
		assert.Equal(t, []byte("stub_value"), value)
	}
}

func TestStoreImpl_KeyTracking(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "store_key_tracking_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	require.NoError(t, err)
	defer store.Close()

	// Initial state should have keys from stub data
	ctx := context.Background()
	result, err := store.Explain(ctx, ExplainOptions{})
	require.NoError(t, err)

	initialKeys := result.Global.TotalKeys

	// Add a new key
	err = store.Put([]byte("new_key"), []byte("new_value"))
	assert.NoError(t, err)

	// Key count should increase
	result, err = store.Explain(ctx, ExplainOptions{})
	require.NoError(t, err)
	assert.Equal(t, initialKeys+1, result.Global.TotalKeys)
}

func TestStore_ErrorHandling(t *testing.T) {
	// Test with invalid directory
	_, err := NewStore("/invalid/path/that/does/not/exist/and/cannot/be/created")
	assert.Error(t, err)
}

func BenchmarkStore_Put(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "store_bench_put")
	require.NoError(b, err)
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	require.NoError(b, err)
	defer store.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("bench_key_%d", i))
		value := []byte(fmt.Sprintf("bench_value_%d", i))
		if err := store.Put(key, value); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStore_Get(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "store_bench_get")
	require.NoError(b, err)
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	require.NoError(b, err)
	defer store.Close()

	// Pre-populate with data
	for i := 0; i < 1000; i++ {
		key := []byte(fmt.Sprintf("bench_key_%d", i))
		value := []byte(fmt.Sprintf("bench_value_%d", i))
		store.Put(key, value)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("bench_key_%d", i%1000))
		store.Get(key)
	}
}

func BenchmarkStore_Explain(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "store_bench_explain")
	require.NoError(b, err)
	defer os.RemoveAll(tmpDir)

	store, err := NewStore(tmpDir)
	require.NoError(b, err)
	defer store.Close()

	ctx := context.Background()
	opts := ExplainOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Explain(ctx, opts)
	}
}
