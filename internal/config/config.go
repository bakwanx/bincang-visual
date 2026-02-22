package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Google   GoogleConfig
	TurnStun TurnStunConfig
	Storage  StorageConfig
}

type ServerConfig struct {
	Port        int
	Host        string
	BaseURL     string
	Environment string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	Secret     string
	Expiration int // in hours
}

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type TurnStunConfig struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

type StorageConfig struct {
	Type      string // "local" or "s3"
	LocalPath string
	S3Bucket  string
	S3Region  string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found or failed to load")
	}
	var turnUrls []string
	json.Unmarshal([]byte(getEnv("TURN_URLS", "[example.com:your-port,example.com:your-port]")), &turnUrls)

	config := &Config{
		Server: ServerConfig{
			Port:        getEnvAsInt("SERVER_PORT", 8080),
			Host:        getEnv("SERVER_HOST", "0.0.0.0"),
			BaseURL:     getEnv("BASE_URL", "http://192.168.1.68:8080"),
			Environment: getEnv("ENV", "development"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			Expiration: getEnvAsInt("JWT_EXPIRATION", 24),
		},
		Google: GoogleConfig{
			ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),
		},
		TurnStun: TurnStunConfig{
			URLs:       turnUrls,
			Username:   getEnv("TURN_USERNAME", "your-turn-username"),
			Credential: getEnv("TURN_SECRET", "your-turn-password"),
		},
		Storage: StorageConfig{
			Type:      getEnv("STORAGE_TYPE", "local"),
			LocalPath: getEnv("STORAGE_LOCAL_PATH", "./storage"),
			S3Bucket:  getEnv("STORAGE_S3_BUCKET", ""),
			S3Region:  getEnv("STORAGE_S3_REGION", "us-east-1"),
		},
	}

	return config, nil
}

func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.JWT.Secret == "your-secret-key-change-in-production" && c.Server.Environment == "production" {
		return fmt.Errorf("JWT secret must be changed in production")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}
