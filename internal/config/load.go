package config

import "os"

func Load() *Config {

	return &Config{
		HTTPPort: getEnv("HTTP_PORT", "8080"),
	}
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultVal
}
