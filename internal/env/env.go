package env

import (
	"os"
	"strings"
)

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
