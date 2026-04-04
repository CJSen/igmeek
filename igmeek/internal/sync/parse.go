package sync

import (
	"fmt"
	"strings"
)

func ParseOwnerRepo(fullName string) (string, string, error) {
	parts := strings.Split(fullName, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repository format: %s (expected owner/repo)", fullName)
	}
	return parts[0], parts[1], nil
}
