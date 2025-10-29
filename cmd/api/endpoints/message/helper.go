package message

import (
	"regexp"
	"strconv"
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

var (
	urlRegex         = regexp.MustCompile(`(?i)\bhttps?://[^\s]+`)
	userMentionRegex = regexp.MustCompile(`<@(\d+)>`)
	roleMentionRegex = regexp.MustCompile(`<@&(\d+)>`)
)

// HasURL returns true if the input string contains at least one URL.
func HasURL(text string) bool {
	return urlRegex.MatchString(text)
}

// MentionsExtractor extract mentions from the text message
// <@2226021950625415168> - extracts all as user ids
// <@&2226021950625415168> - extracts all as role ids
//
//	@here		- if present in content returns true for here mentions
//	@everyone	- if present in content returns true for everyone mentions
//
// Example: "Hello <@2226021950625415168> and <@2229920912390488064>" returns slice if users [2226021950625415168, 2229920912390488064]
func MentionsExtractor(content string) (users, roles []int64, everyone, here bool) {
	// Check for special mentions
	everyone = strings.Contains(content, "@everyone")
	here = strings.Contains(content, "@here")

	// Extract user mentions: <@123456>
	userMatches := userMentionRegex.FindAllStringSubmatch(content, -1)
	if len(userMatches) > 0 {
		users = make([]int64, 0, len(userMatches))
		seen := make(map[int64]struct{}) // Deduplicate mentions

		for _, match := range userMatches {
			if len(match) > 1 {
				if id, err := strconv.ParseInt(match[1], 10, 64); err == nil {
					if _, exists := seen[id]; !exists {
						seen[id] = struct{}{}
						users = append(users, id)
					}
				}
			}
		}
	}

	// Extract role mentions: <@&123456>
	roleMatches := roleMentionRegex.FindAllStringSubmatch(content, -1)
	if len(roleMatches) > 0 {
		roles = make([]int64, 0, len(roleMatches))
		seen := make(map[int64]struct{}) // Deduplicate mentions

		for _, match := range roleMatches {
			if len(match) > 1 {
				if id, err := strconv.ParseInt(match[1], 10, 64); err == nil {
					if _, exists := seen[id]; !exists {
						seen[id] = struct{}{}
						roles = append(roles, id)
					}
				}
			}
		}
	}

	return users, roles, everyone, here
}
