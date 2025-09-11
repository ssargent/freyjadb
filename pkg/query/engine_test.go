package query

import (
	"testing"

	"github.com/ssargent/freyjadb/pkg/index"
)

func TestFieldQuery_Validate(t *testing.T) {
	tests := []struct {
		name    string
		query   FieldQuery
		wantErr bool
	}{
		{
			name: "valid equality query",
			query: FieldQuery{
				Field:    "age",
				Operator: "=",
				Value:    25,
			},
			wantErr: false,
		},
		{
			name: "valid range query",
			query: FieldQuery{
				Field:    "age",
				Operator: ">",
				Value:    18,
			},
			wantErr: false,
		},
		{
			name: "empty field",
			query: FieldQuery{
				Field:    "",
				Operator: "=",
				Value:    25,
			},
			wantErr: true,
		},
		{
			name: "invalid operator",
			query: FieldQuery{
				Field:    "age",
				Operator: "invalid",
				Value:    25,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.query.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("FieldQuery.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJSONFieldExtractor_Extract(t *testing.T) {
	extractor := &JSONFieldExtractor{}

	tests := []struct {
		name     string
		jsonData string
		field    string
		want     interface{}
		wantErr  bool
	}{
		{
			name:     "extract string field",
			jsonData: `{"name":"John","age":25}`,
			field:    "name",
			want:     "John",
			wantErr:  false,
		},
		{
			name:     "extract number field",
			jsonData: `{"name":"John","age":25}`,
			field:    "age",
			want:     float64(25), // JSON unmarshals numbers as float64
			wantErr:  false,
		},
		{
			name:     "field not found",
			jsonData: `{"name":"John","age":25}`,
			field:    "email",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "invalid JSON",
			jsonData: `{"name":"John","age":`,
			field:    "name",
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractor.Extract([]byte(tt.jsonData), tt.field)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONFieldExtractor.Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("JSONFieldExtractor.Extract() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimpleQueryEngine_ExecuteQuery(t *testing.T) {
	// Create index manager
	indexManager := index.NewIndexManager(4)

	// Create query engine
	engine := NewSimpleQueryEngine(indexManager)

	// Create extractor
	extractor := &JSONFieldExtractor{}

	// Test query execution (will return empty results since index is empty)
	query := FieldQuery{
		Field:    "age",
		Operator: "=",
		Value:    25,
	}

	iterator, err := engine.ExecuteQuery(nil, "test-partition", query, extractor)
	if err != nil {
		t.Fatalf("ExecuteQuery failed: %v", err)
	}

	// Should have no results
	if iterator.Next() {
		t.Error("Expected no results, but got some")
	}

	iterator.Close()
}
