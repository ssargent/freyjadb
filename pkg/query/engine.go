package query

import (
	"context"
	"fmt"

	"github.com/ssargent/freyjadb/pkg/index"
	"github.com/ssargent/freyjadb/pkg/store"
)

// SimpleQueryEngine implements basic field-based queries using secondary indexes
type SimpleQueryEngine struct {
	indexManager *index.IndexManager
	kvStore      *store.KVStore
}

// NewSimpleQueryEngine creates a new query engine
func NewSimpleQueryEngine(indexManager *index.IndexManager, kvStore *store.KVStore) *SimpleQueryEngine {
	return &SimpleQueryEngine{
		indexManager: indexManager,
		kvStore:      kvStore,
	}
}

// ExecuteQuery executes a single field query
func (qe *SimpleQueryEngine) ExecuteQuery(ctx context.Context, partitionKey string, query FieldQuery, extractor FieldExtractor) (QueryIterator, error) {
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	// Get the secondary index for this field
	idx := qe.indexManager.GetOrCreateIndex(query.Field)

	// For now, implement basic equality search
	// TODO: Add support for range queries and other operators
	switch query.Operator {
	case "=":
		return qe.executeEqualityQuery(ctx, idx, query.Value, extractor)
	case ">", ">=", "<", "<=":
		return qe.executeRangeQuery(ctx, idx, query, extractor)
	default:
		return nil, fmt.Errorf("unsupported operator: %s", query.Operator)
	}
}

// ExecuteRangeQuery executes a range query between two field conditions
func (qe *SimpleQueryEngine) ExecuteRangeQuery(ctx context.Context, partitionKey string, startQuery, endQuery FieldQuery, extractor FieldExtractor) (QueryIterator, error) {
	if err := startQuery.Validate(); err != nil {
		return nil, fmt.Errorf("invalid start query: %w", err)
	}
	if err := endQuery.Validate(); err != nil {
		return nil, fmt.Errorf("invalid end query: %w", err)
	}

	// Ensure both queries are for the same field
	if startQuery.Field != endQuery.Field {
		return nil, fmt.Errorf("range query fields must match: %s != %s", startQuery.Field, endQuery.Field)
	}

	idx := qe.indexManager.GetOrCreateIndex(startQuery.Field)
	return qe.executeRangeQueryBetween(ctx, idx, startQuery, endQuery, extractor)
}

// executeEqualityQuery handles exact field value matches
func (qe *SimpleQueryEngine) executeEqualityQuery(ctx context.Context, idx *index.SecondaryIndex, value interface{}, extractor FieldExtractor) (QueryIterator, error) {
	// Search the index for matching records
	primaryKeys, err := idx.Search(value)
	if err != nil {
		return nil, fmt.Errorf("index search failed: %w", err)
	}

	// Fetch actual records from KV store
	results := make([]QueryResult, 0, len(primaryKeys))
	for _, key := range primaryKeys {
		if qe.kvStore != nil {
			// Fetch the actual record from KV store
			value, err := qe.kvStore.Get(key)
			if err != nil {
				// Skip records that can't be fetched (might be deleted)
				continue
			}
			results = append(results, QueryResult{
				Key:   key,
				Value: value,
			})
		} else {
			// Fallback for testing: return key with empty value
			results = append(results, QueryResult{
				Key:   key,
				Value: []byte{},
			})
		}
	}

	return &simpleIterator{results: results}, nil
}

// executeRangeQuery handles single-field range queries
func (qe *SimpleQueryEngine) executeRangeQuery(ctx context.Context, idx *index.SecondaryIndex, query FieldQuery, extractor FieldExtractor) (QueryIterator, error) {
	var startValue, endValue interface{}

	switch query.Operator {
	case ">":
		startValue = query.Value
		endValue = nil // No upper bound
	case ">=":
		startValue = query.Value
		endValue = nil // No upper bound
	case "<":
		startValue = nil // No lower bound
		endValue = query.Value
	case "<=":
		startValue = nil // No lower bound
		endValue = query.Value
	default:
		return nil, fmt.Errorf("unsupported range operator: %s", query.Operator)
	}

	primaryKeys, err := idx.SearchRange(startValue, endValue)
	if err != nil {
		return nil, fmt.Errorf("range search failed: %w", err)
	}

	// Fetch actual records from KV store
	results := make([]QueryResult, 0, len(primaryKeys))
	for _, key := range primaryKeys {
		if qe.kvStore != nil {
			value, err := qe.kvStore.Get(key)
			if err != nil {
				continue // Skip records that can't be fetched
			}
			results = append(results, QueryResult{
				Key:   key,
				Value: value,
			})
		} else {
			results = append(results, QueryResult{
				Key:   key,
				Value: []byte{},
			})
		}
	}

	return &simpleIterator{results: results}, nil
}

// executeRangeQueryBetween handles range queries between two values
func (qe *SimpleQueryEngine) executeRangeQueryBetween(ctx context.Context, idx *index.SecondaryIndex, startQuery, endQuery FieldQuery, extractor FieldExtractor) (QueryIterator, error) {
	primaryKeys, err := idx.SearchRange(startQuery.Value, endQuery.Value)
	if err != nil {
		return nil, fmt.Errorf("range search failed: %w", err)
	}

	// Fetch actual records from KV store
	results := make([]QueryResult, 0, len(primaryKeys))
	for _, key := range primaryKeys {
		if qe.kvStore != nil {
			value, err := qe.kvStore.Get(key)
			if err != nil {
				continue // Skip records that can't be fetched
			}
			results = append(results, QueryResult{
				Key:   key,
				Value: value,
			})
		} else {
			results = append(results, QueryResult{
				Key:   key,
				Value: []byte{},
			})
		}
	}

	return &simpleIterator{results: results}, nil
}

// simpleIterator implements QueryIterator for basic result streaming
type simpleIterator struct {
	results []QueryResult
	index   int
}

func (it *simpleIterator) Next() bool {
	if it.index < len(it.results) {
		it.index++
		return true
	}
	return false
}

func (it *simpleIterator) Result() QueryResult {
	if it.index > 0 && it.index <= len(it.results) {
		return it.results[it.index-1]
	}
	return QueryResult{}
}

func (it *simpleIterator) Close() error {
	// Cleanup resources if needed
	return nil
}
