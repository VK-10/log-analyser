package classifier

import "strings"

func ClassifyWithRegex(msg string) string {
	msg = strings.ToLower(msg)

	if strings.Contains(msg, "error") {
		return "error"
	}

	if strings.Contains(msg, "timeout") {
		return "timeout"
	}

	if strings.Contains(msg, "failed") {
		return "failure"
	}

	return ""
}
