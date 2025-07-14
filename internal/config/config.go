package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	JWTSecret            string
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Флаги командной строки
	flag.StringVar(&cfg.RunAddress, "a", ":8080", "server address")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "database URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "accrual system address")
	flag.Parse()

	// Переменные окружения имеют приоритет
	if addr := os.Getenv("RUN_ADDRESS"); addr != "" {
		cfg.RunAddress = addr
	}

	if dbURI := os.Getenv("DATABASE_URI"); dbURI != "" {
		cfg.DatabaseURI = dbURI
	}

	if accrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); accrualAddr != "" {
		cfg.AccrualSystemAddress = accrualAddr
	}

	// JWT секрет (можно генерировать или брать из переменных)
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.JWTSecret = secret
	} else {
		cfg.JWTSecret = "your-secret-key"
	}

	return cfg, nil
}
