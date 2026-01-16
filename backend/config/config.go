package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	Ollama    OllamaConfig
	Synthesia SynthesiaConfig
}

type ServerConfig struct {
	Port         string
	Environment  string
	AllowOrigins []string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	Secret           string
	AccessExpiresIn  time.Duration
	RefreshExpiresIn time.Duration
	Issuer           string
}

type OllamaConfig struct {
	BaseURL string
	Model   string
	Timeout time.Duration
}

type SynthesiaConfig struct {
	APIKey     string
	BaseURL    string
	WebhookURL string
	AvatarID   string
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("server.allowOrigins", []string{"http://localhost:4200"})

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5433)
	viper.SetDefault("database.user", "secusense")
	viper.SetDefault("database.password", "secusense")
	viper.SetDefault("database.dbname", "secusense")
	viper.SetDefault("database.sslmode", "disable")

	viper.SetDefault("jwt.accessExpiresIn", "15m")
	viper.SetDefault("jwt.refreshExpiresIn", "7d")
	viper.SetDefault("jwt.issuer", "secusense")

	viper.SetDefault("ollama.baseUrl", "http://localhost:11434")
	viper.SetDefault("ollama.model", "llama3.2")
	viper.SetDefault("ollama.timeout", "5m")

	viper.SetDefault("synthesia.baseUrl", "https://api.synthesia.io/v2")
	viper.SetDefault("synthesia.avatarId", "anna_costume1_cameraA")

	// Environment variable bindings
	viper.AutomaticEnv()
	viper.SetEnvPrefix("SECUSENSE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Explicit bindings for nested config
	viper.BindEnv("database.host", "SECUSENSE_DATABASE_HOST")
	viper.BindEnv("database.port", "SECUSENSE_DATABASE_PORT")
	viper.BindEnv("database.user", "SECUSENSE_DATABASE_USER")
	viper.BindEnv("database.password", "SECUSENSE_DATABASE_PASSWORD")
	viper.BindEnv("database.dbname", "SECUSENSE_DATABASE_DBNAME")
	viper.BindEnv("jwt.secret", "SECUSENSE_JWT_SECRET")
	viper.BindEnv("ollama.baseUrl", "SECUSENSE_OLLAMA_BASEURL")
	viper.BindEnv("server.allowOrigins", "SECUSENSE_SERVER_ALLOWORIGINS")

	// Read config file if exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	accessExpiry, _ := time.ParseDuration(viper.GetString("jwt.accessExpiresIn"))
	refreshExpiry, _ := time.ParseDuration(viper.GetString("jwt.refreshExpiresIn"))
	ollamaTimeout, _ := time.ParseDuration(viper.GetString("ollama.timeout"))

	return &Config{
		Server: ServerConfig{
			Port:         viper.GetString("server.port"),
			Environment:  viper.GetString("server.environment"),
			AllowOrigins: viper.GetStringSlice("server.allowOrigins"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("database.host"),
			Port:     viper.GetInt("database.port"),
			User:     viper.GetString("database.user"),
			Password: viper.GetString("database.password"),
			DBName:   viper.GetString("database.dbname"),
			SSLMode:  viper.GetString("database.sslmode"),
		},
		JWT: JWTConfig{
			Secret:           viper.GetString("jwt.secret"),
			AccessExpiresIn:  accessExpiry,
			RefreshExpiresIn: refreshExpiry,
			Issuer:           viper.GetString("jwt.issuer"),
		},
		Ollama: OllamaConfig{
			BaseURL: viper.GetString("ollama.baseUrl"),
			Model:   viper.GetString("ollama.model"),
			Timeout: ollamaTimeout,
		},
		Synthesia: SynthesiaConfig{
			APIKey:     viper.GetString("synthesia.apiKey"),
			BaseURL:    viper.GetString("synthesia.baseUrl"),
			WebhookURL: viper.GetString("synthesia.webhookUrl"),
			AvatarID:   viper.GetString("synthesia.avatarId"),
		},
	}, nil
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
}
