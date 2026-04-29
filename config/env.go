package config

import (
	"os"
	"strconv"
)

func envString(key string, fallback *string) {
	if val := os.Getenv(key); val != "" {
		*fallback = val
	}
}

func envInt(key string, fallback *int) {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			*fallback = n
		}
	}
}

func envBool(key string, fallback *bool) {
	if val := os.Getenv(key); val != "" {
		*fallback = val == "true" || val == "1"
	}
}

func (c *Config) loadEnv() {
	envString("SERVER_HOST", &c.Server.Host)
	envInt("SERVER_PORT", &c.Server.Port)
	envString("DB_PATH", &c.Database.Path)
	envBool("DEBUG", &c.Debug)
}
