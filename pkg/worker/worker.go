package worker

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"pilot/internal/database"
	"pilot/pkg/models"
	"pilot/pkg/scheduler"
	// Other necessary imports
)

type Worker struct {
	TaskQueue      chan models.Step
	DatabaseClient *database.DB
	Scheduler      *scheduler.Scheduler
	Logger         *log.Logger
	// Other fields as needed
}

func (w *Worker) ExecuteTask(step models.Step) {
	// Log task start
	w.Logger.Printf("Starting task: %v\n", step.ID)
	fmt.Printf("Starting task: %v\n", step.ID)

	err := w.performTaskAction(step)
	if err != nil {
		// Handle error, log it, and possibly update task state to 'failed'
		w.Logger.Printf("Error executing task %v: %v\n", step.ID, err)
		return
	}

	step.State = "completed"
	err = w.DatabaseClient.UpdateStep(step)
	if err != nil {
		w.Logger.Printf("Error updating Step: %v\n", err)
	}
	w.Scheduler.TaskCompleted <- step.ID
	w.Logger.Printf("Completed task: %v\n", step.ID)
}

func (w *Worker) performTaskAction(step models.Step) error {
	// Retrieve the base path for scripts from an environment variable
	basePath := os.Getenv("PROJECT_PATH")
	if basePath == "" {
		log.Println("PROJECT_PATH environment variable is not set.")
		return errors.New("script base path not configured")
	}

	// Construct the full script path
	scriptPath := filepath.Join(basePath, step.Command)

	// Execute the script
	cmd := exec.Command("python", scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error executing script: %s, Output: %s\n", err, output)
		return err
	}

	log.Printf("Output: %s\n", output)
	return nil
}

func (w *Worker) StartWorker(taskQueue chan models.Step, dbClient *database.DB, scheduler *scheduler.Scheduler, logger *log.Logger) {
	worker := Worker{
		TaskQueue:      taskQueue,
		DatabaseClient: dbClient,
		Scheduler:      scheduler,
		Logger:         logger,
	}

	for task := range worker.TaskQueue {
		worker.ExecuteTask(task)
	}
}
