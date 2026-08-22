package api

// Shared test-only string constants. These exist purely to avoid repeating
// the same literal across multiple _test.go files in this package (goconst).
const (
	// testAPIKey is the API key used by tests that exercise the
	// apiKeyMiddleware / ServerConfig.APIKey path.
	testAPIKey = "test-key"

	// testEncryptionKey is a 32-byte string suitable for AES-256 key
	// derivation in tests.
	testEncryptionKey = "12345678901234567890123456789012"

	// testAPIKeyDescription is a generic description used when creating
	// APIKey fixtures in tests.
	testAPIKeyDescription = "Test API key" //nolint:gosec // G101: this is a description string, not a credential

	// testKeyName is a generic key name used across handler tests.
	testKeyName = "testkey"

	// expectedPutSuccessBody is the JSON body returned by handlePut on
	// success.
	expectedPutSuccessBody = `{"success":true,"data":{"message":"Key-value pair stored successfully"}}`
)
