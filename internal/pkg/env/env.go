package env

import "os"

func GetOrDefault(kety string, defaultValue string) string {
	value := os.Getenv(kety)

	if value == "" {
		return defaultValue
	}

	return value
}
