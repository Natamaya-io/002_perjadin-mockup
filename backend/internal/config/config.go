package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Database  DatabaseConfig
	Signatory SignatoryConfig
	Redis     RedisConfig
	App       AppConfig
	Fonnte    FonnteConfig
	JWT       JWTConfig
}

type AppConfig struct {
	Port string
	Env  string
}

type FonnteConfig struct {
	Token string
}

type SignatoryConfig struct {
	PPKName       string
	PPKNIP        string
	BendaharaName string
	BendaharaNIP  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

type JWTConfig struct {
	Secret string
	Expiry int // Hours
}

func LoadConfig() *Config {
	// Defaults
	viper.SetDefault("APP_PORT", "8081")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("JWT_EXPIRY", 24)

	// Production Signatories (Defaults)
	viper.SetDefault("PPK_NAME", "Arief Hafidiyanto")
	viper.SetDefault("PPK_NIP", "19720827 200312 1 002")
	viper.SetDefault("BENDAHARA_NAME", "Liana Setyawati")
	viper.SetDefault("BENDAHARA_NIP", "19800512 200901 2 001")

	viper.AutomaticEnv()

	var cfg Config
	cfg.App.Port = viper.GetString("APP_PORT")
	cfg.App.Env = viper.GetString("APP_ENV")

	dbHost := viper.GetString("DB_HOST")
	cfg.Database.Host = dbHost
	cfg.Database.Port = viper.GetString("DB_PORT")
	cfg.Database.User = viper.GetString("DB_USER")
	cfg.Database.Password = viper.GetString("DB_PASSWORD")
	cfg.Database.Name = viper.GetString("DB_NAME")
	cfg.Database.SSLMode = viper.GetString("DB_SSLMODE")

	cfg.Redis.Host = viper.GetString("REDIS_HOST")
	cfg.Redis.Port = viper.GetString("REDIS_PORT")
	cfg.Redis.Password = viper.GetString("REDIS_PASSWORD")

	cfg.JWT.Secret = viper.GetString("JWT_SECRET")
	cfg.JWT.Expiry = viper.GetInt("JWT_EXPIRY")

	cfg.Fonnte.Token = viper.GetString("FONNTE_API")

	cfg.Signatory.PPKName = viper.GetString("PPK_NAME")
	cfg.Signatory.PPKNIP = viper.GetString("PPK_NIP")
	cfg.Signatory.BendaharaName = viper.GetString("BENDAHARA_NAME")
	cfg.Signatory.BendaharaNIP = viper.GetString("BENDAHARA_NIP")

	return &cfg
}
