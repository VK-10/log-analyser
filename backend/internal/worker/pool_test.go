package worker

import (
	"fmt"
	"log-classifier/internal/models"
	"testing"
)

func TestProcessLogs_OrderIsPreserved(t *testing.T) {
	// build 20 logs with identifiable messages
	logs := make([]models.LogEntry, 20)
	for i := range logs {
		logs[i] = models.LogEntry{
			Source:     fmt.Sprintf("source-%d", i),
			LogMessage: fmt.Sprintf("User user%d logged in", i), // hits regex, fast
		}
	}

	results := ProcessLogs(logs, 4)

	if len(results) != len(logs) {
		t.Fatalf("expected %d results, got %d", len(logs), len(results))
	}

	for i, r := range results {
		if r == nil {
			t.Fatalf("result[%d] is nil", i)
		}
		// since regex result carries log source back, verify it matches
		if r.LogSource != fmt.Sprintf("source-%d", i) {
			t.Errorf("position %d: expected LogSource source-%d, got %s", i, i, r.LogSource)
		}
	}
}

func TestProcessLogs_OrderIsPreserved_Stress(t *testing.T) {
	for run := 0; run < 100; run++ {
		logs := make([]models.LogEntry, 20)
		for i := range logs {
			logs[i] = models.LogEntry{
				Source:     fmt.Sprintf("source-%d", i),
				LogMessage: fmt.Sprintf("User user%d logged in", i),
			}
		}

		results := ProcessLogs(logs, 4)

		for i, r := range results {
			if r.LogSource != fmt.Sprintf("source-%d", i) {
				t.Fatalf("run %d: position %d got LogSource %s", run, i, r.LogSource)
			}
		}
	}
}
