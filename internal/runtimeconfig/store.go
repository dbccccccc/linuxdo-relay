package runtimeconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

var (
	ErrNotFound = errors.New("runtime config not found")
)

type Store struct {
	Path string
	mu   sync.RWMutex
}

type Data struct {
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
}

type DatabaseConfig struct {
	DSN string `json:"dsn"`
}

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
}

func DefaultPath() string {
	if v := os.Getenv("APP_RUNTIME_CONFIG_PATH"); v != "" {
		return v
	}
	return filepath.Join("runtimeconfig", "config.json")
}

func NewStore(path string) *Store {
	if path == "" {
		path = DefaultPath()
	}
	return &Store{Path: path}
}

func (s *Store) Load() (*Data, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, err := os.ReadFile(s.Path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("read runtime config: %w", err)
	}

	var data Data
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, fmt.Errorf("parse runtime config: %w", err)
	}
	return &data, nil
}

func (s *Store) Save(data *Data) error {
	if data == nil {
		return errors.New("runtime config data is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Dir(s.Path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create runtime config dir: %w", err)
	}

	payload, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("encode runtime config: %w", err)
	}

	tmpPath := s.Path + ".tmp"
	if err := os.WriteFile(tmpPath, payload, 0o600); err != nil {
		return fmt.Errorf("write runtime config tmp: %w", err)
	}

	if err := os.Rename(tmpPath, s.Path); err != nil {
		return fmt.Errorf("persist runtime config: %w", err)
	}

	return nil
}
