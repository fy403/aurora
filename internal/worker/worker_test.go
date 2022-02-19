package worker_test

import (
	"aurora/internal/worker"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactURL(t *testing.T) {
	t.Parallel()

	broker := "amqp://guest:guest@localhost:5672"
	redactedURL := worker.RedactURL(broker)
	assert.Equal(t, "amqp://localhost:5672", redactedURL)
}

func TestPreConsumeHandler(t *testing.T) {
	t.Parallel()

	worker := &worker.Worker{}

	worker.SetPreConsumeHandler(SamplePreConsumeHandler)
	assert.True(t, worker.PreConsumeHandler())
}

func SamplePreConsumeHandler(w *worker.Worker) bool {
	return true
}

// func TestNewWorker(t *testing.T) {
// 	t.Parallel()

// 	server := getTestServer(t)

// 	worker.NewWorker("test_worker", 1)
// 	assert.NoError(t, nil)
// }

// func TestNewCustomQueueWorker(t *testing.T) {
// 	t.Parallel()

// 	server := getTestServer(t)

// 	worker.NewCustomQueueWorker("test_customqueueworker", 1, "test_queue")
// 	assert.NoError(t, nil)
// }
