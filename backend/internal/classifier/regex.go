package classifier

import (
	"regexp"
	"strings"
)

type regexRule struct {
	pattern *regexp.Regexp
	label   string
}

var regexRules = []regexRule{
	{regexp.MustCompile(`(?i)User \w+ logged (in|out)\.`), "User Action"},
	{regexp.MustCompile(`(?i)Backup (started|ended) at .+`), "System Notification"},
	{regexp.MustCompile(`(?i)Backup completed successfully\.`), "System Notification"},
	{regexp.MustCompile(`(?i)System updated to version .+`), "System Notification"},
	{regexp.MustCompile(`(?i)File .+ uploaded successfully by user .+`), "System Notification"},
	{regexp.MustCompile(`(?i)Disk cleanup completed successfully\.`), "System Notification"},
	{regexp.MustCompile(`(?i)System reboot initiated by user .+`), "System Notification"},
	{regexp.MustCompile(`(?i)Account with ID .+ created by .+`), "User Action"},
}

func ClassifyWithRegex(msg string) string {
	_ = strings.ToLower

	for _, rule := range regexRules {
		if rule.pattern.MatchString(msg) {
			return rule.label
		}
	}
	return ""
}
