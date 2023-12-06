package worker

import (
	"log"
	"os"
	"pilot/pkg/models"
	"testing"
)

// Import statements...

func TestExecuteTask(t *testing.T) {
	// Set the path to your mock scripts
	os.Setenv("STEP_SCRIPT_PATH", "/path/to/mock/scripts")

	// Create a mock step
	mockStep := models.Step{
		ID:      1,
		Command: "mock_step.py", // Use the mock step script
		// Other fields if necessary
	}

	// Set up your logger and worker instance
	logger := log.New(os.Stdout, "test-logger: ", log.LstdFlags)
	worker := Worker{
		TaskQueue:      make(chan models.Step, 1),
		DatabaseClient: nil, // Provide a mock or test database if needed
		Logger:         logger,
	}

	// Enqueue the mock step
	worker.TaskQueue <- mockStep

	// Execute the task
	go worker.ExecuteTask(mockStep)

	// Implement any additional assertions or checks as needed
}
