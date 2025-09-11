package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// ExplainOptions configures the explain operation
type ExplainOptions struct {
	WithSamples int
	WithMetrics bool
	PK          string
}

// ExplainResult holds the results of an explain operation
type ExplainResult struct {
	Global struct {
		TotalKeys     int           `json:"total_keys"`
		ActiveKeys    int           `json:"active_keys"`
		Tombstones    int           `json:"tombstones"`
		TotalSizeMB   float64       `json:"total_size_mb"`
		LiveSizeMB    float64       `json:"live_size_mb"`
		IndexMemoryMB float64       `json:"index_memory_mb"`
		Uptime        time.Duration `json:"uptime"`
	} `json:"global"`

	Segments []Segment `json:"segments"`

	Partitions map[string]PKStats `json:"partitions"`

	Diagnostics struct {
		CompactionReady []string `json:"compaction_ready"`
		CRCErrors       int      `json:"crc_errors"`
		Samples         []Sample `json:"samples,omitempty"`
		Metrics         struct {
			AvgGetLatencyMs float64 `json:"avg_get_latency_ms,omitempty"`
			IORateMBs       float64 `json:"io_rate_mbs,omitempty"`
		} `json:"metrics,omitempty"`
	} `json:"diagnostics"`

	Warnings []string `json:"warnings,omitempty"`
}

type Segment struct {
	ID      string  `json:"id"`
	Keys    int     `json:"keys"`
	DeadPct float64 `json:"dead_pct"`
	SizeMB  float64 `json:"size_mb"`
}

type Sample struct {
	Key   string    `json:"key"`
	Value string    `json:"value_truncated"`
	Ts    time.Time `json:"timestamp"`
}

type SKRange struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Min   string `json:"min,omitempty"`
	Max   string `json:"max,omitempty"`
}

type PKStats struct {
	Keys        int       `json:"keys"`
	SKRanges    []SKRange `json:"sk_ranges"`
	Cardinality string    `json:"cardinality"`
}

// Store defines the basic store interface
type Store interface {
	Explain(ctx context.Context, opts ExplainOptions) (*ExplainResult, error)
	Put(key, value []byte) error
	Get(key []byte) ([]byte, error)
	Close() error
}

// StoreImpl is the concrete implementation
type StoreImpl struct {
	dataDir   string
	startTime time.Time
	keys      int
	keysMap   map[string]struct{}
}

// NewStore creates a new store
func NewStore(dataDir string) (Store, error) {
	s := &StoreImpl{
		dataDir:   dataDir,
		startTime: time.Now(),
		keys:      0,
		keysMap:   make(map[string]struct{}),
	}

	// Stub data
	for i := 1; i <= 1250; i++ {
		key := fmt.Sprintf("key:%d", i)
		s.keysMap[key] = struct{}{}
		if i%10 == 0 {
			delete(s.keysMap, key)
		}
	}
	s.keys = len(s.keysMap)

	// Create dummy segments
	if _, err := os.Stat(filepath.Join(dataDir, "001.data")); os.IsNotExist(err) {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return nil, err
		}
		for _, seg := range []string{"001.data", "002.data"} {
			f, err := os.Create(filepath.Join(dataDir, seg))
			if err != nil {
				return nil, err
			}
			f.Close()
		}
	}

	return s, nil
}

// Explain gathers stats
func (s *StoreImpl) Explain(ctx context.Context, opts ExplainOptions) (*ExplainResult, error) {
	res := &ExplainResult{}
	res.Global.TotalKeys = s.keys
	res.Global.ActiveKeys = s.keys * 9 / 10
	res.Global.Tombstones = s.keys / 10
	res.Global.TotalSizeMB = 5.2
	res.Global.LiveSizeMB = 4.1
	res.Global.Uptime = time.Since(s.startTime)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	res.Global.IndexMemoryMB = float64(m.Alloc) / (1024 * 1024)

	res.Segments = []Segment{
		{ID: "001", Keys: 600, DeadPct: 10.0, SizeMB: 2.1},
		{ID: "002", Keys: 650, DeadPct: 15.0, SizeMB: 3.1},
	}

	for _, seg := range res.Segments {
		if seg.DeadPct > 20.0 {
			res.Diagnostics.CompactionReady = append(res.Diagnostics.CompactionReady, seg.ID)
		}
	}

	res.Partitions = map[string]PKStats{
		"User": {Keys: 800, SKRanges: []SKRange{{Name: "Location", Count: 500, Min: "loc:1", Max: "loc:750"}}, Cardinality: "1:N"},
		"Item": {Keys: 450, SKRanges: []SKRange{{Name: "Category", Count: 250, Min: "cat:1", Max: "cat:300"}}, Cardinality: "N:1"},
	}

	if opts.WithSamples > 0 {
		samples := []Sample{
			{Key: "user:123", Value: "john_doe@example.com", Ts: time.Now().Add(-time.Hour)},
			{Key: "item:456", Value: "laptop", Ts: time.Now().Add(-2 * time.Hour)},
		}
		if opts.WithSamples < len(samples) {
			samples = samples[:opts.WithSamples]
		}
		res.Diagnostics.Samples = samples
	}

	if opts.PK != "" {
		if pkStats, exists := res.Partitions[opts.PK]; !exists || pkStats.Keys == 0 {
			res.Warnings = append(res.Warnings, fmt.Sprintf("No data for PK: %s", opts.PK))
		}
	}

	res.Diagnostics.CRCErrors = 0

	if opts.WithMetrics {
		res.Diagnostics.Metrics.AvgGetLatencyMs = 1.2
		res.Diagnostics.Metrics.IORateMBs = 10.5
	}

	return res, nil
}

// Put
func (s *StoreImpl) Put(key, value []byte) error {
	s.keysMap[string(key)] = struct{}{}
	s.keys = len(s.keysMap)
	return nil
}

// Get
func (s *StoreImpl) Get(key []byte) ([]byte, error) {
	if _, exists := s.keysMap[string(key)]; !exists {
		return nil, fmt.Errorf("key not found")
	}
	return []byte("stub_value"), nil
}

// Close
func (s *StoreImpl) Close() error {
	return nil
}
