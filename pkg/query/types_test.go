package query

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryResult_Struct(t *testing.T) {
	key := []byte("test_key")
	value := []byte("test_value")

	result := QueryResult{
		Key:   key,
		Value: value,
	}

	assert.Equal(t, key, result.Key)
	assert.Equal(t, value, result.Value)
}

func TestQueryIterator_Interface(t *testing.T) {
	// Test that QueryIterator is an interface
	var iterator QueryIterator
	assert.Nil(t, iterator)
}

func TestQueryEngine_Interface(t *testing.T) {
	// Test that QueryEngine is an interface
	var engine QueryEngine
	assert.Nil(t, engine)

	// Test interface methods exist
	ctx := context.Background()
	var extractor FieldExtractor = &JSONFieldExtractor{}
	query := FieldQuery{Field: "name", Operator: "=", Value: "Alice"}

	// These would normally be implemented by a concrete type
	// We're just testing that the interface is properly defined
	_ = ctx
	_ = extractor
	_ = query
}

func TestFieldExtractor_Interface(t *testing.T) {
	// Test that FieldExtractor is an interface
	var extractor FieldExtractor
	assert.Nil(t, extractor)

	// Test with concrete implementation
	extractor = &JSONFieldExtractor{}
	assert.NotNil(t, extractor)
}

func TestJSONFieldExtractor_NullValues(t *testing.T) {
	extractor := &JSONFieldExtractor{}

	// Test with null value in JSON
	jsonData := []byte(`{"name": null, "age": 25}`)
	result, err := extractor.Extract(jsonData, "name")
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestJSONFieldExtractor_ArrayValues(t *testing.T) {
	extractor := &JSONFieldExtractor{}

	jsonData := []byte(`{"tags": ["red", "blue", "green"], "count": 3}`)
	result, err := extractor.Extract(jsonData, "tags")
	assert.NoError(t, err)
	assert.IsType(t, []interface{}{}, result)

	tags := result.([]interface{})
	assert.Len(t, tags, 3)
	assert.Equal(t, "red", tags[0])
	assert.Equal(t, "blue", tags[1])
	assert.Equal(t, "green", tags[2])
}

func TestJSONFieldExtractor_NumberTypes(t *testing.T) {
	extractor := &JSONFieldExtractor{}

	jsonData := []byte(`{"int": 42, "float": 3.14, "zero": 0}`)
	result, err := extractor.Extract(jsonData, "int")
	assert.NoError(t, err)
	assert.Equal(t, float64(42), result)

	result, err = extractor.Extract(jsonData, "float")
	assert.NoError(t, err)
	assert.Equal(t, 3.14, result)

	result, err = extractor.Extract(jsonData, "zero")
	assert.NoError(t, err)
	assert.Equal(t, float64(0), result)
}

func BenchmarkJSONFieldExtractor_Extract(b *testing.B) {
	extractor := &JSONFieldExtractor{}

	testData := map[string]interface{}{
		"name":     "Alice",
		"age":      25,
		"city":     "New York",
		"active":   true,
		"score":    95.5,
		"tags":     []string{"developer", "golang"},
		"metadata": map[string]interface{}{"version": "1.0", "env": "prod"},
	}

	jsonData, _ := json.Marshal(testData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extractor.Extract(jsonData, "name")
	}
}

func BenchmarkFieldQuery_Validate(b *testing.B) {
	query := FieldQuery{
		Field:    "age",
		Operator: "=",
		Value:    25,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = query.Validate()
	}
}
