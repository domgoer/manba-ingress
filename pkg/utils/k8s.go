package utils

import (
	"fmt"
	"strings"
)

// ParseNameNS parses a string searching a namespace and name
func ParseNameNS(input string) (string, string, error) {
	nsName := strings.Split(input, "/")
	if len(nsName) != 2 {
		return "", "", fmt.Errorf("invalid format (namespace/name) found in '%v'", input)
	}

	return nsName[0], nsName[1], nil
}
