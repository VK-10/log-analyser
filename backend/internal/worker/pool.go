package worker

import (
	"log-classifier/internal/classifier"
	"log-classifier/internal/metrics"
	"log-classifier/internal/models"
	"sync"
)

type job struct {
	index int
	entry models.LogEntry
}

type result struct {
	index int
	value *models.ClassificationResult
}

func ProcessLogs(logs []models.LogEntry, workers int) []*models.ClassificationResult {
	jobs := make(chan job, len(logs))
	results := make(chan result, len(logs))

	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			metrics.ActiveWorkers.Inc()
			defer metrics.ActiveWorkers.Dec()

			for j := range jobs {
				r := classifier.Classify(j.entry)
				results <- result{index: j.index, value: r}
			}
		}(w)
	}

	for i, log := range logs {
		jobs <- job{index: i, entry: log}
	}
	close(jobs)

	// Close results channel after all workers finish
	go func() {
		wg.Wait()
		close(results)
	}()

	output := make([]*models.ClassificationResult, len(logs))
	for r := range results {
		output[r.index] = r.value
	}

	return output

}
