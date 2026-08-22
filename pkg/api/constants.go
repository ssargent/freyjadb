package api

// Shared field/label/MIME-type string constants.
//
// These exist to avoid repeating the same string literal across the package
// (which golangci-lint's goconst check flags once a literal is repeated a
// handful of times) - callers should reference the named constant instead of
// restating the raw string.
const (
	// fieldMessage/fieldStatus are common JSON response field names.
	fieldMessage = "message"
	fieldStatus  = "status"

	// labelMethod/labelEndpoint/labelOperation are Prometheus metric label names.
	labelMethod    = "method"
	labelEndpoint  = "endpoint"
	labelOperation = "operation"

	// ContentTypeJSONHeader and ContentTypeOctetStreamHeader are the HTTP
	// Content-Type header values used when serving JSON or raw byte payloads.
	ContentTypeJSONHeader        = "application/json"
	ContentTypeOctetStreamHeader = "application/octet-stream"
)
