package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func GetEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func SplitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// ConvertDSN converts mysql://user:pass@host:port/db?params to user:pass@tcp(host:port)/db?params
func ConvertDSN(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("parse dsn uri: %w", err)
	}
	if u.Scheme != "mysql" {
		return "", fmt.Errorf("unsupported scheme %q", u.Scheme)
	}

	user := u.User.Username()
	pass, _ := u.User.Password()
	host := u.Host
	dbName := strings.TrimPrefix(u.Path, "/")
	query := u.RawQuery

	var dsn string
	if pass != "" {
		dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s", user, pass, host, dbName)
	} else {
		dsn = fmt.Sprintf("%s@tcp(%s)/%s", user, host, dbName)
	}
	if query != "" {
		dsn += "?" + query
	}

	return dsn, nil
}
