package api

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// RelationshipRequest represents a relationship creation/deletion request
type RelationshipRequest struct {
	FromKey  string `json:"from_key"`
	ToKey    string `json:"to_key"`
	Relation string `json:"relation"`
}

// ServerConfig holds configuration for the API server
type ServerConfig struct {
	Port    int
	APIKey  string
	DataDir string
}
