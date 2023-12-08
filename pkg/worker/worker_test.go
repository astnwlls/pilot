package worker

import (
	"fmt"
	"log"
	"os"
	"pilot/pkg/models"
	"testing"
)

// Import statements...

func TestExecuteTask(t *testing.T) {
	os.Setenv("PROJECT_PATH", `C:\Users\AWills\Documents\pilot\maps`)

	mockStep := models.Step{
		ID:      1,
		Command: "main.py", // Use the mock step script
	}

	logger := log.New(os.Stdout, "test-logger: ", log.LstdFlags)
	worker := Worker{
		TaskQueue:      make(chan models.Step, 1),
		DatabaseClient: nil, // Provide a mock or test database if needed
		Logger:         logger,
	}

	worker.TaskQueue <- mockStep

	fmt.Printf("Starting task here")
	worker.ExecuteTask(mockStep)
}
