package config

// Config holds all application configuration
type Config struct {
	Version     int                `json:"version"`
	Connections []SavedConnection  `json:"connections"`
}

// SavedConnection holds connection details
type SavedConnection struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	Timeout  int    `json:"timeout"` // seconds, 0 = default
}

// EncryptedFile format on disk
type EncryptedFile struct {
	Version int    `json:"version"`
	Salt    string `json:"salt"` // base64
	Data    string `json:"data"` // base64 AES-GCM encrypted
}
