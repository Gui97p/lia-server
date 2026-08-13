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
	JWTSecret     string
	EncryptionKey []byte
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	cfg := &Config{
		Port:        os.Getenv("PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}

	encryptionKeyRaw := os.Getenv("ENCRYPTION_KEY")
	if encryptionKeyRaw == "" {
		return nil, errors.New("ENCRYPTION_KEY is required")
	}

	encryptionKey, err := base64.StdEncoding.DecodeString(encryptionKeyRaw)
	if err != nil {
		return nil, fmt.Errorf("ENCRYPTION_KEY is not valid base64: %w", err)
	}
	if len(encryptionKey) != 32 {
		return nil, fmt.Errorf("ENCRYPTION_KEY must decode to 32 bytes, got %d", len(encryptionKey))
	}

	cfg.EncryptionKey = encryptionKey

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg, nil
}
