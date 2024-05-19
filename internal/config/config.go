package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Postgres    struct {
		Host     string
		User     string
		Password string
		Port     string
		Name     string `yaml:"databaseName"`
	} `yaml:"postgres"`
	Auth struct {
		PasswordSalt           string
		SigningKey             string
		AccessTokenTTL         time.Duration `yaml:"accessTokenTTL"`
		RefreshTokenTTL        time.Duration `yaml:"refreshTokenTTL"`
		VerificationCodeLength int           `yaml:"verificationCodeLength"`
	} `yaml:"auth"`
	HTTP struct {
		Host               string        `yaml:"host"`
		Port               string        `yaml:"port"`
		ReadTimeout        time.Duration `yaml:"readTimeout"`
		WriteTimeout       time.Duration `yaml:"writeTimeout"`
		MaxHeaderMegabytes int           `yaml:"maxHeaderBytes"`
	} `yaml:"http"`
}

func LoadConfigs(yamlFile, envFile string) (Config, error) {
	var cfg Config

	err := cleanenv.ReadConfig(yamlFile, &cfg)
	if err != nil {
		log.Fatalf("Error loading .yaml file: %v", err)
	}

	err = godotenv.Load(envFile)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	cfg.Environment = os.Getenv("APP_ENV")
	cfg.Postgres.User = os.Getenv("POSTGRES_USER")
	cfg.Postgres.Password = os.Getenv("POSTGRES_PASSWORD")
	cfg.Postgres.Port = os.Getenv("POSTGRES_PORT")
	cfg.Postgres.Host = os.Getenv("POSTGRES_HOST")
	cfg.Auth.PasswordSalt = os.Getenv("PASSWORD_SALT")
	cfg.Auth.SigningKey = os.Getenv("JWT_KEY")
	cfg.HTTP.Host = os.Getenv("HTTP_HOST")
	cfg.HTTP.Port = os.Getenv("HTTP_PORT")

	return cfg, nil
}
