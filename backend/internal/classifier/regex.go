package classifier

import "regexp"

type ClassificationResult struct {
	LabelID    string  `json:"label_id"`
	Label      string  `json:"label"`
	Source     string  `json:"source"`
	Confidence float64 `json:"confidence"`
}

type regexRule struct {
	pattern *regexp.Regexp
	labelID string
	label   string
}

var regexRules = []regexRule{
	{
		pattern: regexp.MustCompile(`(?i)User \w+ logged (in|out)\.`),
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
		pattern: regexp.MustCompile(`(?i)Backup completed successfully\.`),
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
		pattern: regexp.MustCompile(`(?i)Disk cleanup completed successfully\.`),
		labelID: "SYSTEM_NOTIFICATION",
		label:   "System Notification",
	},
	{
		pattern: regexp.MustCompile(`(?i)System reboot initiated by user .+`),
		labelID: "SYSTEM_NOTIFICATION",
		label:   "System Notification",
	},
}

func ClassifyWithRegex(msg string) *ClassificationResult {
	for _, rule := range regexRules {
		if rule.pattern.MatchString(msg) {
			return &ClassificationResult{
				LabelID:    rule.labelID,
				Label:      rule.label,
				Source:     "regex",
				Confidence: 0.95,
			}
		}
	}
	return nil
}
