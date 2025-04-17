package utils

import "strings"

func GetFileNameWithoutExt(filepath string) string {
	// Get base filename from path
	base := filepath[strings.LastIndex(filepath, "/")+1:]

	// Remove extension
	if idx := strings.LastIndex(base, "."); idx != -1 {
		base = base[:idx]
	}

	return base
}
