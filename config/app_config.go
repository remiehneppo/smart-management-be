package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type AppConfig struct {
	Port    string       `mapstructure:"port"`
	Logger  LoggerConfig `mapstructure:"logger"`
	MongoDB struct {
		URI      string `mapstructure:"URI"`
		Database string `mapstructure:"DATABASE"`
	} `mapstructure:"MONGODB"`
	JWT struct {
		Secret string `mapstructure:"SECRET"`
		Issuer string `mapstructure:"ISSUER"`
		Expire int64  `mapstructure:"EXPIRE"`
	} `mapstructure:"JWT"`
	Environment string `mapstructure:"ENVIRONMENT"`
}

// Config holds configuration for the logger
type LoggerConfig struct {
	LogLevel        string        `mapstructure:"log_level"`
	EnableConsole   bool          `mapstructure:"enable_console"`
	EnableFile      bool          `mapstructure:"enable_file"`
	FilePath        string        `mapstructure:"file_path"`
	FileNamePattern string        `mapstructure:"file_name_pattern"`
	MaxAge          time.Duration `mapstructure:"max_age"`
	RotationTime    time.Duration `mapstructure:"rotation_time"`
}

// LoadConfig loads configuration from environment variables and config files
func LoadConfig(path string) (*AppConfig, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	godotenv.Load()
	// Try to read the config file
	viper.ReadInConfig()

	// Configure Viper to use environment variables
	viper.AutomaticEnv()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.BindEnv("port", "PORT")
	viper.BindEnv("MONGODB.URI", "MONGODB_URI")
	viper.BindEnv("MONGODB.DATABASE", "MONGODB_DATABASE")
	viper.BindEnv("JWT.SECRET", "JWT_SECRET")
	viper.BindEnv("JWT.ISSUER", "JWT_ISSUER")
	viper.BindEnv("JWT.EXPIRE", "JWT_EXPIRE")
	viper.BindEnv("ENVIRONMENT", "ENVIRONMENT")

	var config AppConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return &config, nil
}
