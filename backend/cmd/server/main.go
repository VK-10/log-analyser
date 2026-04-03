package main

import (
	"log"
	"log-classifier/internal/api"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	classifyTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "log_classifier_requests_total",
			Help: "Total classification requests",
		},
		[]string{"status"},
	)
	classifyDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: "log_classifier_duration_seconds",
			Help: "Classification request duration",
		},
	)
)

func init() {
	prometheus.MustRegister(classifyTotal, classifyDuration)
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/classify", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		api.ClassifyHandler(w, r)

		classifyDuration.Observe(time.Since(start).Seconds())
		classifyTotal.WithLabelValues("success").Inc()
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy"}`))
	})

	mux.Handle("/metrics", promhttp.Handler())

	handler := loggingMiddleware(enableCORS(mux))

	log.Println("Server running on :8080")
	log.Println("Metrics available at :8080/metrics")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
