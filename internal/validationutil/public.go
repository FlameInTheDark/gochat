package validationutil

import (
	"sort"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func PublicMessage(err error) string {
	if err == nil {
		return ""
	}

	switch typed := err.(type) {
	case validation.Errors:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if message := PublicMessage(typed[key]); message != "" {
				return message
			}
		}
		return ""
	default:
		return strings.TrimSpace(err.Error())
	}
}
