package worker

import (
	"log-classifier/internal/classifier"
	"log-classifier/internal/metrics"
	"log-classifier/internal/models"
	"sync"
)

func ProcessLogs(logs []models.LogEntry, workers int) []models.ClassificationResult {
	jobs := make(chan models.LogEntry, len(logs))
	results := make(chan models.ClassificationResult, len(logs))

	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(WorkerID int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {

				}
			}()

			metrics.ActiveWorkers.Inc()
			defer metrics.ActiveWorkers.Dec()

			for job := range jobs {
				label := classifier.Classify(job)
				results <- models.ClassificationResult{
					Source: job.Source,
					Label:  label,
				}
			}
		}(w)
	}

	for _, log := range logs {
		jobs <- log
	}
	close(jobs)

	// Close results channel after all workers finish
	go func() {
		wg.Wait()
		close(results)
	}()

	var output []models.ClassificationResult
	for result := range results {
		output = append(output, result)
	}

	return output

}
