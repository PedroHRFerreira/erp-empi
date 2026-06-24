package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Env         string         `mapstructure:"env"`
	APIPort     string         `mapstructure:"apiPort"`
	FrontendURL string         `mapstructure:"frontendURL"`
	JWT         JWTConfig      `mapstructure:"jwt"`
	Database    DatabaseConfig `mapstructure:"database"`
	Admin       AdminConfig    `mapstructure:"admin"`
}

type JWTConfig struct {
	AccessSecret     string `mapstructure:"accessSecret"`
	AccessTTLMinutes int    `mapstructure:"accessTTLMinutes"`
}

type DatabaseConfig struct {
	PostgresDSN string `mapstructure:"postgresDSN"`
}

type AdminConfig struct {
	Name                  string  `mapstructure:"name"`
	CPF                   string  `mapstructure:"cpf"`
	Password              string  `mapstructure:"password"`
	Email                 string  `mapstructure:"email"`
	Phone                 string  `mapstructure:"phone"`
	MarkupPercent         float64 `mapstructure:"markupPercent"`
	MachineFeePercent     float64 `mapstructure:"machineFeePercent"`
	InstallmentFeePercent float64 `mapstructure:"installmentFeePercent"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.SetDefault("env", "local")
	v.SetDefault("apiPort", "8080")
	v.SetDefault("frontendURL", "http://localhost:3000")
	v.SetDefault("jwt.accessTTLMinutes", 15)
	v.SetDefault("admin.markupPercent", 10)
	v.SetDefault("admin.machineFeePercent", 0)
	v.SetDefault("admin.installmentFeePercent", 0)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	for _, key := range v.AllKeys() {
		switch value := v.Get(key).(type) {
		case string:
			v.Set(key, os.ExpandEnv(value))
		}
	}

	cfg := new(Config)
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}
	normalizeConfig(cfg)
	return cfg, nil
}

func normalizeConfig(cfg *Config) {
	if cfg.APIPort == "" {
		cfg.APIPort = "8080"
	}
	if cfg.JWT.AccessTTLMinutes == 0 {
		cfg.JWT.AccessTTLMinutes = int(parseDurationEnv("JWT_ACCESS_TTL_MINUTES", 15*time.Minute).Minutes())
	}
}

func parseDurationEnv(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	if key == "JWT_ACCESS_TTL_MINUTES" {
		return time.Duration(value) * time.Minute
	}
	return time.Duration(value) * time.Hour
}
