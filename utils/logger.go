package utils

import (
	"log"
	"os"
	"strings"
)

// NewLogger returns a logger using the provided tag (e.g., "[FIREBASE] ").
func NewLogger(tag string) *log.Logger {
	formattedTag := "[" + strings.ToUpper(strings.TrimSpace(tag)) + "] "
	return log.New(os.Stdout, formattedTag, 0)
}
