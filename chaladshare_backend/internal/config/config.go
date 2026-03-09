package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	AppPort          string
	DatabaseHost     string
	DatabasePort     int
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	DatabaseSSLMode  string
	JWTSecret        string

	TokenTTLMinutes int
	CookieName      string
	AllowOrigin     string
}

func LoadConfig() (Config, error) {

	_ = godotenv.Load()

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set default values
	viper.SetDefault("POSTGRES.HOST", "localhost")
	viper.SetDefault("POSTGRES.PORT", 5432)
	viper.SetDefault("POSTGRES.USER", "postgres")
	viper.SetDefault("POSTGRES.PASSWORD", "")
	viper.SetDefault("POSTGRES.DBNAME", "chaladshare")
	viper.SetDefault("POSTGRES.SSLMODE", "disable")
	viper.SetDefault("JWT.SECRET", "changeme")
	viper.SetDefault("APP.PORT", "8080")

	// ADD THIS PART
	viper.SetDefault("JWT.TTL_MINUTES", 30)
	viper.SetDefault("COOKIE.NAME", "access_token")
	viper.SetDefault("ALLOW.ORIGIN", "http://localhost:3000")

	// Set config values
	config := Config{
		AppPort:          viper.GetString("APP.PORT"),
		DatabaseHost:     viper.GetString("POSTGRES.HOST"),
		DatabasePort:     viper.GetInt("POSTGRES.PORT"),
		DatabaseUser:     viper.GetString("POSTGRES.USER"),
		DatabasePassword: viper.GetString("POSTGRES.PASSWORD"),
		DatabaseName:     viper.GetString("POSTGRES.DBNAME"),
		DatabaseSSLMode:  viper.GetString("POSTGRES.SSLMODE"),
		JWTSecret:        viper.GetString("JWT.SECRET"),

		// ADD THIS PATH
		TokenTTLMinutes: viper.GetInt("JWT.TTL_MINUTES"),
		CookieName:      viper.GetString("COOKIE.NAME"),
		AllowOrigin:     viper.GetString("ALLOW.ORIGIN"),
	}

	return config, nil
}

func (c *Config) GetConnectionString() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DatabaseHost,
		c.DatabasePort,
		c.DatabaseUser,
		c.DatabasePassword,
		c.DatabaseName,
		c.DatabaseSSLMode)
}
