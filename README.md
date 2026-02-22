# Log Classifier

A distributed log classification system that automatically categorizes log messages using a three-stage pipeline: **Regex → BERT → LLM**. The backend is written in Go and exposes an HTTP API, while the classification logic is powered by Python microservices.

---

## Architecture Overview

```
Client
  │
  ▼
Go HTTP Server (:8080)
  │
  ▼
Worker Pool (concurrent goroutines)
  │
  ▼
Classification Pipeline
  ├── 1. Regex Classifier     → fast, rule-based, high confidence
  ├── 2. BERT Classifier      → ML model via Python service (:5000)
  └── 3. LLM Classifier       → Groq LLaMA via Python service (:5001)
```

The pipeline is **short-circuit**: if a stage produces a confident result, downstream stages are skipped. Each ML service call is protected by a **circuit breaker** and **retry logic**.

---

## Classification Labels

| Label ID             | Label                 | Primary Source |
|----------------------|-----------------------|----------------|
| `USER_ACTION`        | User Action           | Regex          |
| `SYSTEM_NOTIFICATION`| System Notification   | Regex          |
| `AUTH_ERROR`         | Authentication Error  | BERT           |
| `DB_ERROR`           | Database Error        | BERT           |
| `BACKUP`             | Backup Event          | BERT           |
| `INFO`               | Informational Log     | BERT           |
| `WORKFLOW_ERROR`     | Workflow Error        | LLM            |
| `DEPRECATION_WARNING`| Deprecation Warning   | LLM            |
| `UNCLASSIFIED`       | Unclassified          | Fallback       |

---

## Project Structure

```
.
├── backend/
│   ├── cmd/server/main.go          # HTTP server entrypoint
│   └── internal/
│       ├── api/handler.go          # /classify endpoint handler
│       ├── classifier/
│       │   ├── pipeline.go         # Orchestrates the 3-stage pipeline
│       │   ├── regex.go            # Regex-based classifier
│       │   ├── bert.go             # BERT service client
│       │   ├── llm.go              # LLM service client
│       │   ├── circuit.go          # Circuit breaker implementation
│       │   ├── circuit_test.go     # Circuit breaker unit tests
│       │   └── retry.go            # Retry with backoff logic
│       ├── models/log.go           # Shared data models
│       ├── metrics/metrics.go      # Prometheus metrics
│       └── worker/pool.go          # Concurrent worker pool
├── processor/
│   ├── processor_regex.py          # Regex classification (Python)
│   ├── processor_bert.py           # BERT/SentenceTransformer classification
│   └── processor_llm.py            # LLM classification via Groq
├── server.py                       # Flask servers for BERT (:5000) & LLM (:5001)
├── labels.py                       # Label definitions
└── models/
    └── log_classifier_model.joblib # Trained sklearn classifier
```

---

## Getting Started

### Prerequisites

- Go 1.24+
- Python 3.9+
- A [Groq API key](https://console.groq.com/) (for LLM classification)

### 1. Start the Python Services

```bash
# Install dependencies
pip install flask sentence-transformers scikit-learn joblib groq python-dotenv

# Set your Groq API key
echo "GROQ_API_KEY=your_key_here" > .env

# Start BERT (:5000) and LLM (:5001) services
python server.py
```

### 2. Start the Go Server

```bash
cd backend
go mod download
go run cmd/server/main.go
```

The API will be available at `http://localhost:8080`.

---

## API Reference

### `POST /classify`

Classifies a batch of log entries.

**Request Body**
```json
[
  { "source": "app-server", "log_message": "User admin123 logged in." },
  { "source": "db-server",  "log_message": "Connection pool exhausted after 30s" }
]
```

**Response**
```json
[
  {
    "label_id": "USER_ACTION",
    "label": "User Action",
    "source": "regex",
    "confidence": 0.95
  },
  {
    "label_id": "WORKFLOW_ERROR",
    "label": "Workflow Error",
    "source": "llm",
    "confidence": 0.85
  }
]
```

### `GET /health`

Returns server health status.

```json
{ "status": "healthy" }
```

### `GET /metrics`

Exposes Prometheus metrics for scraping.

---

## Observability

Prometheus metrics are available at `/metrics` and include:

| Metric | Type | Description |
|--------|------|-------------|
| `log_classifications_total` | Counter | Total classifications by classifier and label |
| `log_classification_duration_seconds` | Histogram | Classification latency |
| `log_classification_errors_total` | Counter | Errors by classifier and type |
| `log_classifier_active_workers` | Gauge | Number of active worker goroutines |
| `log_classifier_circuit_breaker_open` | Gauge | Circuit breaker state (1=open, 0=closed) |
| `log_classifier_http_requests_total` | Counter | HTTP requests by endpoint, method, status |
| `log_classifier_http_request_duration_seconds` | Histogram | HTTP request latency |

---

## Resilience Features

**Circuit Breaker** — Each downstream service (BERT, LLM) is wrapped in an independent circuit breaker:

| Service | Max Failures | Reset Timeout |
|---------|-------------|---------------|
| LLM     | 3           | 5 seconds     |
| BERT    | 5           | 10 seconds    |

States: `Closed → Open → Half-Open → Closed`

**Retry with Backoff** — Transient failures are retried up to 2 times with exponential backoff (100ms, 200ms). Permanent errors (e.g. circuit open) skip retries immediately.

**Context Timeouts** — BERT calls time out after 4 seconds; LLM calls after 2 seconds.

**Worker Pool** — Log entries are processed concurrently using a configurable pool (default: 4 workers).

---

## Running Tests

```bash
cd backend
go test ./internal/classifier/...
```

The circuit breaker has full unit test coverage including state transitions, half-open probing, and concurrent request rejection.

---

## Configuration

| Parameter | Location | Default |
|-----------|----------|---------|
| BERT service URL | `classifier/bert.go` | `http://127.0.0.1:5000/classify` |
| LLM service URL | `classifier/llm.go` | `http://127.0.0.1:5001/classify` |
| Worker count | `api/handler.go` | `4` |
| Server port | `cmd/server/main.go` | `:8080` |
| BERT confidence threshold | `classifier/pipeline.go` | `0.20` |
| BERT classifier threshold | `processor/processor_bert.py` | `0.50` |
