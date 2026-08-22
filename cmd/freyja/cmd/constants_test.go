package cmd

// Shared test fixtures used across init_test.go, service_test.go, and
// up_test.go to avoid duplicated string literals (goconst).
const (
	testSystemKey    = "test-system-key-1234567890123456"     // 32 bytes for AES-256
	testSystemAPIKey = "test-system-api-key-1234567890123456" // 32 bytes for AES-256
	testClientAPIKey = "test-client-api-key-1234567890123456" // 32 bytes for AES-256

	testBindAddressAll = "0.0.0.0"

	testLogLevelInfo = "info"
)
