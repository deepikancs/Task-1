package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	DemoMode          bool
	ServerPort        string
	CORSAllowedOrigin string
	DBHost            string
	DBPort            string
	DBName            string
	DBUser            string
	DBPassword        string
	DefaultDemoToday  string
}

// Load reads an optional .env file and populates Config from environment variables.
func Load() (*Config, error) {
	_ = loadDotEnv(".env")

	demoMode := strings.EqualFold(getEnv("DEMO_MODE", "true"), "true")

	cfg := &Config{
		DemoMode:          demoMode,
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		CORSAllowedOrigin: getEnv("CORS_ALLOWED_ORIGIN", "http://localhost:5173"),
		DBHost:            getEnv("DB_HOST", "127.0.0.1"),
		DBPort:            getEnv("DB_PORT", "3306"),
		DBName:            getEnv("DB_NAME", "Diviga"),
		DBUser:            os.Getenv("DB_USER"),
		DBPassword:        os.Getenv("DB_PASSWORD"),
		DefaultDemoToday:  getEnv("DEFAULT_DEMO_TODAY", "2026-10-11"),
	}

	// Only validate MySQL credentials when DEMO_MODE is false
	if !cfg.DemoMode {
		if strings.TrimSpace(cfg.DBUser) == "" {
			return nil, fmt.Errorf("missing required environment variable DB_USER (check your backend/.env file)")
		}
		if strings.TrimSpace(cfg.DBName) == "" {
			return nil, fmt.Errorf("missing required environment variable DB_NAME (check your backend/.env file)")
		}
		if cfg.DBPassword == "YOUR_CHOSEN_PASSWORD" || cfg.DBPassword == "YOUR_LOCAL_MYSQL_PASSWORD" || cfg.DBPassword == "your_mysql_password_placeholder" {
			return nil, fmt.Errorf("DB_PASSWORD in backend/.env is still set to a placeholder; replace it with your chosen MySQL password")
		}
	}

	return cfg, nil
}

// MySQLDSN builds a MySQL Data Source Name safely using mysql.Config so special
// characters in passwords, usernames, or database names are properly handled.
// Never log the returned DSN string.
func (c *Config) MySQLDSN() string {
	mc := mysql.NewConfig()
	mc.User = c.DBUser
	mc.Passwd = c.DBPassword
	mc.Net = "tcp"
	mc.Addr = fmt.Sprintf("%s:%s", c.DBHost, c.DBPort)
	mc.DBName = c.DBName
	mc.ParseTime = true
	mc.Loc = time.Local
	mc.AllowNativePasswords = true
	mc.Params = map[string]string{
		"charset": "utf8mb4,utf8",
	}
	return mc.FormatDSN()
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	return fallback
}

// loadDotEnv parses key=value pairs from a local .env file and applies them
// so edits to backend/.env immediately take effect.
func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, `"'`)
		_ = os.Setenv(key, val)
	}
	return scanner.Err()
}
