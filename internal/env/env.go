package env

import (
	"fmt"
	"os"
	"strings"
)

func RequiredEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return v, nil
}

func GetEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func RedactMongoURI(uri string) string {
	at := strings.LastIndex(uri, "@")
	if at == -1 {
		return uri
	}
	schemeEnd := strings.Index(uri, "://")
	if schemeEnd == -1 {
		return uri
	}
	creds := uri[schemeEnd+3 : at]
	if !strings.Contains(creds, ":") {
		return uri
	}
	username, _, _ := strings.Cut(creds, ":")
	return uri[:schemeEnd+3] + username + ":***@" + uri[at+1:]
}
