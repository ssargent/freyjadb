package query

// Shared test fixtures used across engine_test.go and types_test.go
// (extracted to satisfy goconst for strings repeated across the package's tests).
const (
	testFieldAge  = "age"
	testFieldName = "name"

	testJSONNameAge = `{"name":"John","age":25}`
)
