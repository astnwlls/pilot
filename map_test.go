package main // or the name of the package where your topological sort is

import (
	"fmt"
	"log"
	"os"
	"pilot/internal/database"
	"pilot/pkg/models" // import your models package
	"pilot/pkg/scheduler"
	"pilot/pkg/worker"
	"reflect"
	"testing"
	"time"
)

func TestMapExecutionFlow(t *testing.T) {
	os.Setenv("PROJECT_PATH", os.Getenv("TEST_PROJECT_PATH"))

	db, err := database.NewDB("meta.db")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	scheduler := scheduler.NewScheduler(db, 10)

	startDate := time.Now()
	endDate := startDate.Add(24 * time.Hour)

	// Define a mock map with steps
	step1 := models.Step{
		Name:         "Step 1",
		MapID:        1,
		State:        "pending",
		Command:      "xerox/step1.py",
		StartDate:    startDate,
		EndDate:      endDate,
		Dependencies: []int{},
	}
	step1ID, err := db.AddStep(&step1)
	fmt.Printf("Added step A with ID: %d\n", step1ID)

	step2 := models.Step{
		Name:         "Step 2",
		MapID:        1,
		State:        "pending",
		Command:      "xerox/step2.py",
		StartDate:    startDate,
		EndDate:      endDate,
		Dependencies: []int{step1ID},
	}
	step2ID, err := db.AddStep(&step2)

	step3 := models.Step{
		Name:         "Step 3",
		MapID:        1,
		State:        "pending",
		Command:      "xerox/step3.py",
		StartDate:    startDate,
		EndDate:      endDate,
		Dependencies: []int{step1ID, step2ID},
	}
	step3ID, err := db.AddStep(&step3)

	step1.ID = step1ID
	step2.ID = step2ID
	step3.ID = step3ID

	mockMap := models.Map{
		Name:             "Sample Map",
		ScheduleInterval: "0 10 * * *",
		StartDate:        startDate,
		LastRun:          time.Time{},
		Steps:            []models.Step{step1, step2, step3},
	}

	done := make(chan bool)
	taskOrder := make([]int, 0)

	// Override QueueTaskFunc for testing
	scheduler.QueueTaskFunc = func(step models.Step) {
		fmt.Printf("Queueing task: %+v\n", step)
		scheduler.TaskQueue <- step

		taskOrder = append(taskOrder, step.ID)
		if len(taskOrder) == len(mockMap.Steps) {
			done <- true
		}
	}

	logger := log.New(os.Stdout, "test-logger: ", log.LstdFlags)
	worker := worker.Worker{
		TaskQueue:      scheduler.TaskQueue,
		DatabaseClient: db,
		Scheduler:      scheduler,
		Logger:         logger,
	}

	go scheduler.Start()
	go worker.StartWorker(worker.TaskQueue, db, scheduler, logger)

	fmt.Printf("started worker\n")
	scheduler.ScheduleMap(mockMap)

	<-done

	expectedOrder := []int{step1ID, step2ID, step3ID}

	if !reflect.DeepEqual(taskOrder, expectedOrder) {
		t.Errorf("Steps were not queued in the correct order. Got %v, want %v", taskOrder, expectedOrder)
	}

	close(done)

}
