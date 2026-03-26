package config

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() Config {
	_ = loadDotEnv(".env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = databaseURLFromParts()
	}
	return Config{
		Port:        port,
		DatabaseURL: databaseURL,
	}
}

func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		value := strings.TrimSpace(parts[1])
		if len(value) >= 2 {
			first := value[0]
			last := value[len(value)-1]
			if (first == '\'' && last == '\'') || (first == '"' && last == '"') {
				value = value[1 : len(value)-1]
			}
		}

		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set env %s: %w", key, err)
		}
	}

	return scanner.Err()
}

func databaseURLFromParts() string {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	if host == "" || user == "" || name == "" {
		return ""
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	return (&url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   host + ":" + port,
		Path:   name,
		RawQuery: url.Values{
			"sslmode": []string{sslmode},
		}.Encode(),
	}).String()
}
