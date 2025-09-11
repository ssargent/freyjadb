package query

import (
	"context"
	"encoding/json"
	"fmt"
)

// FieldExtractor defines how to extract field values from record data
type FieldExtractor interface {
	Extract(value []byte, field string) (interface{}, error)
}

// JSONFieldExtractor extracts fields from JSON-encoded values
type JSONFieldExtractor struct{}

// Extract implements FieldExtractor for JSON data
func (e *JSONFieldExtractor) Extract(value []byte, field string) (interface{}, error) {
	if len(value) == 0 {
		return nil, fmt.Errorf("empty value")
	}

	// Parse JSON
	var data map[string]interface{}
	if err := json.Unmarshal(value, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract field value
	fieldValue, exists := data[field]
	if !exists {
		return nil, fmt.Errorf("field '%s' not found in JSON", field)
	}

	return fieldValue, nil
}

// FieldQuery represents a single field-based query condition
type FieldQuery struct {
	Field    string      // Field name to query (e.g., "age", "name")
	Operator string      // Comparison operator: "=", ">", "<", ">=", "<="
	Value    interface{} // Value to compare against
}

// Validate checks if the query is properly formed
func (q *FieldQuery) Validate() error {
	if q.Field == "" {
		return fmt.Errorf("field name cannot be empty")
	}
	if q.Operator == "" {
		return fmt.Errorf("operator cannot be empty")
	}
	validOps := map[string]bool{
		"=": true, ">": true, "<": true, ">=": true, "<=": true,
	}
	if !validOps[q.Operator] {
		return fmt.Errorf("invalid operator: %s", q.Operator)
	}
	return nil
}

// QueryResult represents a single query result
type QueryResult struct {
	Key   []byte // The record key
	Value []byte // The record value
}

// QueryIterator provides streaming access to query results
type QueryIterator interface {
	Next() bool
	Result() QueryResult
	Close() error
}

// QueryEngine handles query execution
type QueryEngine interface {
	ExecuteQuery(ctx context.Context, partitionKey string, query FieldQuery, extractor FieldExtractor) (QueryIterator, error)
	ExecuteRangeQuery(ctx context.Context, partitionKey string, startQuery, endQuery FieldQuery, extractor FieldExtractor) (QueryIterator, error)
}
