package worker

import (
	"log-classifier/internal/classifier"
	"log-classifier/internal/models"
)

func ProcessLogs(logs []models.LogEntry, workers int) []models.ClassificationResult {
	jobs := make(chan models.LogEntry, len(logs))
	results := make(chan models.ClassificationResult, len(logs))

	for w := 0; w < workers; w++ {
		go func() {
			for job := range jobs {
				label := classifier.Classify(job)
				results <- models.ClassificationResult{
					Source: job.Source,
					Label:  label,
				}
			}
		}()
	}

	for _, log := range logs {
		jobs <- log
	}

	close(jobs)

	var output []models.ClassificationResult
	for i := 0; i < len(logs); i++ {
		output = append(output, <-results)
	}

	return output

}
