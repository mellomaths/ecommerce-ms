package env

import "os"

func GetString(key, fallback string) string {
	if str := os.Getenv(key); str != "" {
		return str
	}

	return fallback
}
