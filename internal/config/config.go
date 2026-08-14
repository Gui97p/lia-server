package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DatabaseURL   string
	JWTSecret     []byte
	EncryptionKey []byte
}

func parseBase64(key string) ([]byte, error) {
	raw := os.Getenv(key)

	if raw == "" {
		return nil, fmt.Errorf("%s is required", key)
	}

	value, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("%s is not valid base64: %w", key, err)
	}
	if len(value) != 32 {
		return nil, fmt.Errorf("%s must decode to 32 bytes, got %d", key, len(value))
	}

	return value, nil
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	cfg := &Config{
		Port:        os.Getenv("PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	jwtSecret, err := parseBase64("JWT_SECRET")
	if err != nil {
		return nil, err
	}
	cfg.JWTSecret = jwtSecret

	var encryptionKey []byte
	encryptionKey, err = parseBase64("ENCRYPTION_KEY")
	if err != nil {
		return nil, err
	}
	cfg.EncryptionKey = encryptionKey

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg, nil
}
