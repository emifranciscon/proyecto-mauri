package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

type Config struct {
	Port        string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	JWTSecret   string
	CORSOrigins []string
}

func Load() (*Config, error) {
	port := getenv("PORT", "8080")
	host := getenv("DB_HOST", "127.0.0.1")
	dbPort := getenv("DB_PORT", "3306")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := getenv("DB_NAME", "asientos")
	secret := os.Getenv("JWT_SECRET")
	if user == "" || pass == "" || secret == "" {
		return nil, fmt.Errorf("DB_USER, DB_PASSWORD, and JWT_SECRET are required")
	}
	cors := getenv("CORS_ORIGINS", "http://localhost:5173,http://localhost:3000")
	origins := splitComma(cors)
	return &Config{
		Port:        port,
		DBHost:      host,
		DBPort:      dbPort,
		DBUser:      user,
		DBPassword:  pass,
		DBName:      name,
		JWTSecret:   secret,
		CORSOrigins: origins,
	}, nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func splitComma(s string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			seg := trimSpace(s[start:i])
			if seg != "" {
				out = append(out, seg)
			}
			start = i + 1
		}
	}
	return out
}

func trimSpace(s string) string {
	i, j := 0, len(s)
	for i < j && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\n' || s[j-1] == '\r') {
		j--
	}
	return s[i:j]
}

func (c *Config) DSN() string {
	q := url.Values{}
	q.Set("parseTime", "true")
	q.Set("charset", "utf8mb4")
	q.Set("loc", "UTC")
	if tls := os.Getenv("DB_TLS"); tls != "" {
		// e.g. skip-verify or true — required by many managed MySQL providers (Aiven, etc.)
		q.Set("tls", tls)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, q.Encode())
}

func (c *Config) DBMaxOpen() int {
	if v := os.Getenv("DB_MAX_OPEN_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 25
}

func (c *Config) DBMaxIdle() int {
	if v := os.Getenv("DB_MAX_IDLE_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return n
		}
	}
	return 5
}
