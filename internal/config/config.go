package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configDir      = ".gosqlit"
	configFileName = "config.encrypted"
	currentVersion = 1
)

// Manager handles config file operations
type Manager struct {
	path     string
	password string
	config   *Config
}

// NewManager creates config manager
func NewManager(password string) (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}

	configPath := filepath.Join(homeDir, configDir)
	if err := os.MkdirAll(configPath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config dir: %w", err)
	}

	return &Manager{
		path:     filepath.Join(configPath, configFileName),
		password: password,
	}, nil
}

// Exists checks if config file exists
func (m *Manager) Exists() bool {
	_, err := os.Stat(m.path)
	return err == nil
}

// Load loads and decrypts config
func (m *Manager) Load() (*Config, error) {
	if !m.Exists() {
		return &Config{
			Version:     currentVersion,
			Connections: []SavedConnection{},
		}, nil
	}

	data, err := os.ReadFile(m.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var encrypted EncryptedFile
	if err := json.Unmarshal(data, &encrypted); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	salt, err := DecodeBase64(encrypted.Salt)
	if err != nil {
		return nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	key, err := DeriveKey(m.password, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	encryptedData, err := DecodeBase64(encrypted.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode data: %w", err)
	}

	plaintext, err := Decrypt(encryptedData, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt (wrong password?): %w", err)
	}

	var config Config
	if err := json.Unmarshal(plaintext, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config data: %w", err)
	}

	m.config = &config
	return &config, nil
}

// Save encrypts and saves config
func (m *Manager) Save(config *Config) error {
	config.Version = currentVersion

	plaintext, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	salt, err := GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	key, err := DeriveKey(m.password, salt)
	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}

	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		return fmt.Errorf("failed to encrypt: %w", err)
	}

	encryptedFile := EncryptedFile{
		Version: currentVersion,
		Salt:    EncodeBase64(salt),
		Data:    EncodeBase64(encrypted),
	}

	data, err := json.MarshalIndent(encryptedFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal encrypted file: %w", err)
	}

	if err := os.WriteFile(m.path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	m.config = config
	return nil
}

// AddConnection adds connection to config
func (m *Manager) AddConnection(conn SavedConnection) error {
	if m.config == nil {
		return fmt.Errorf("config not loaded")
	}

	m.config.Connections = append(m.config.Connections, conn)
	return m.Save(m.config)
}

// DeleteConnection removes connection by ID
func (m *Manager) DeleteConnection(id string) error {
	if m.config == nil {
		return fmt.Errorf("config not loaded")
	}

	filtered := make([]SavedConnection, 0)
	for _, conn := range m.config.Connections {
		if conn.ID != id {
			filtered = append(filtered, conn)
		}
	}

	m.config.Connections = filtered
	return m.Save(m.config)
}

// UpdateConnection updates existing connection
func (m *Manager) UpdateConnection(conn SavedConnection) error {
	if m.config == nil {
		return fmt.Errorf("config not loaded")
	}

	for i, c := range m.config.Connections {
		if c.ID == conn.ID {
			m.config.Connections[i] = conn
			return m.Save(m.config)
		}
	}

	return fmt.Errorf("connection not found: %s", conn.ID)
}

// GetConnections returns all saved connections
func (m *Manager) GetConnections() []SavedConnection {
	if m.config == nil {
		return []SavedConnection{}
	}
	return m.config.Connections
}
