package message

import (
	"regexp"
	"strings"
)

func GetAttachmentType(contentType string) string {
	if idx := strings.Index(contentType, "/"); idx != -1 {
		prefix := strings.ToLower(contentType[:idx])
		switch prefix {
		case "image":
			return "image"
		case "video":
			return "video"
		case "audio":
			return "audio"
		}
	}
	return "file"
}

func UniqueAttachmentTypes(types []string) []string {
	seen := make(map[string]struct{})
	var unique []string

	for _, t := range types {
		t = strings.ToLower(t)
		if _, ok := seen[t]; !ok {
			seen[t] = struct{}{}
			unique = append(unique, t)
		}
	}

	return unique
}

var urlRegex = regexp.MustCompile(`(?i)\bhttps?://[^\s]+`)

// HasURL returns true if the input string contains at least one URL.
func HasURL(text string) bool {
	return urlRegex.MatchString(text)
}
