package classifier

import (
	"log-classifier/internal/models"
	"regexp"
)

type regexRule struct {
	pattern *regexp.Regexp
	labelID string
	label   string
}

var regexRules = []regexRule{
	{
		pattern: regexp.MustCompile(`(?i)User \w+ logged (in|out)`),
		labelID: "USER_ACTION",
		label:   "User Action",
	},
	{
		pattern: regexp.MustCompile(`(?i)Account with ID .+ created by .+`),
		labelID: "USER_ACTION",
		label:   "User Action",
	},
	{
		pattern: regexp.MustCompile(`(?i)Backup (started|ended) at .+`),
		labelID: "SYSTEM_NOTIFICATION",
		label:   "System Notification",
	},
	{
		pattern: regexp.MustCompile(`(?i)Backup completed successfully`),
		labelID: "SYSTEM_NOTIFICATION",
		label:   "System Notification",
	},
	{
		pattern: regexp.MustCompile(`(?i)System updated to version .+`),
		labelID: "SYSTEM_NOTIFICATION",
		label:   "System Notification",
	},
	{
		pattern: regexp.MustCompile(`(?i)File .+ uploaded successfully by user .+`),
		labelID: "SYSTEM_NOTIFICATION",
		label:   "System Notification",
	},
	{
		pattern: regexp.MustCompile(`(?i)Disk cleanup completed successfully`),
		labelID: "SYSTEM_NOTIFICATION",
		label:   "System Notification",
	},
	{
		pattern: regexp.MustCompile(`(?i)System reboot initiated by user .+`),
		labelID: "SYSTEM_NOTIFICATION",
		label:   "System Notification",
	},
}

func ClassifyWithRegex(msg string) *models.ClassificationResult {
	for _, rule := range regexRules {
		if rule.pattern.MatchString(msg) {
			return &models.ClassificationResult{
				LabelID:    rule.labelID,
				Label:      rule.label,
				Source:     "regex",
				Confidence: 0.95,
			}
		}
	}
	return nil
}
