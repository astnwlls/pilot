package main // or the name of the package where your topological sort is

import (
	"fmt"
	"log"
	"pilot/internal/database"
	"pilot/pkg/models" // import your models package
	"pilot/pkg/scheduler"
	"reflect"
	"testing"
	"time"
)

func TestTopologicalSort(t *testing.T) {
	steps := []models.Step{
		{ID: 1, Dependencies: []int{2}, Name: "Step1", MapID: 1}, // Step 1 depends on Step 2
		{ID: 2, Dependencies: []int{}, Name: "Step2", MapID: 1},  // Step 2 has no dependencies
	}

	expected := []models.Step{
		{ID: 2, Dependencies: []int{}, Name: "Step2", MapID: 1},
		{ID: 1, Dependencies: []int{2}, Name: "Step1", MapID: 1},
	}

	sortedSteps, err := TopologicalSort(steps)
	if err != nil {
		t.Errorf("TopologicalSort failed with error: %v", err)
	}

	if !reflect.DeepEqual(sortedSteps, expected) {
		t.Errorf("TopologicalSort =\n\n %v,\n\n want %v", sortedSteps, expected)
	}
}

func TestSchedulerInitialization(t *testing.T) {
	db, err := database.NewDB("meta.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	scheduler := scheduler.NewScheduler(db, 10)

	if scheduler == nil {
		t.Error("Failed to initialize Scheduler")
	}
}

func TestSchedulerTimeBasedTriggering(t *testing.T) {
	mockTime := time.Date(2021, time.January, 10, 10, 0, 0, 0, time.UTC)
	db, err := database.NewDB("meta.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	scheduler := scheduler.NewScheduler(db, 10)
	scheduler.SetNowFunc(func() time.Time {
		return mockTime
	})

	// Mock Map with a schedule that should trigger at 10 AM
	mockMap := models.Map{
		ScheduleInterval: "0 10 * * *", // Cron expression for every day at 10 AM
		// ... other necessary initializations ...
	}

	if scheduler.IsTimeToRun(mockMap) {
		t.Log("Scheduler correctly identified map to run")
	} else {
		t.Error("Scheduler failed to identify map to run")
	}

	// Add more assertions as necessary
}

func TestSchedulerDependencyManagement(t *testing.T) {
	// Mock steps with dependencies
	stepA := models.Step{Dependencies: []int{}}
	stepB := models.Step{Dependencies: []int{1}} // Depends on step A

	db, err := database.NewDB("meta.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	err = db.CreateTables()
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Insert mock steps into the database and set their IDs
	stepAID, err := db.AddStep(&stepA)
	if err != nil {
		log.Fatalf("Failed to add step A: %v", err)
	}
	stepA.ID = stepAID
	fmt.Printf("Added step A with ID: %d\n", stepAID)

	stepBID, err := db.AddStep(&stepB)
	if err != nil {
		log.Fatalf("Failed to add step B: %v", err)
	}
	stepB.ID = stepBID
	fmt.Printf("Added step B with ID: %d\n", stepBID)

	mockMap := models.Map{
		Steps:            []models.Step{stepA, stepB},
		ScheduleInterval: "0 10 * * *",
	}

	scheduler := scheduler.NewScheduler(db, 10)

	// Simulate scheduling
	scheduler.ScheduleMap(mockMap)

	// Verify the task queue order
	firstTask := <-scheduler.TaskQueue
	secondTask := <-scheduler.TaskQueue

	if firstTask.ID != stepA.ID || secondTask.ID != stepB.ID {
		t.Errorf("Tasks were not queued in the correct order.")
	}
}
