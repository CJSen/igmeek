package cmd

import "strings"

func parseTagList(values ...string) []string {
	var result []string
	for _, value := range values {
		normalized := strings.ReplaceAll(value, "，", ",")
		for _, part := range strings.Split(normalized, ",") {
			tag := strings.TrimSpace(part)
			if tag != "" {
				result = append(result, tag)
			}
		}
	}
	return result
}
