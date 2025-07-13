package utils

import (
	"encoding/base64"
	"net/http"
)

// Helper functions
func ParseStringArray(v interface{}) []string {
	if v == nil {
		return nil
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, len(arr))
	for i, item := range arr {
		result[i] = item.(string)
	}
	return result
}

func ConvertToBase64URL(data []byte) string {
	mimeType := http.DetectContentType(data)
	encoded := base64.StdEncoding.EncodeToString(data)
	return "data:" + mimeType + ";base64," + encoded
}
