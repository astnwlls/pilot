package main

import (
	"fmt"
	"log"
	"pilot/internal/database"
	"pilot/pkg/models"
	"pilot/pkg/scheduler"
	// other imports
)

// TopologicalSort performs a topological sort on the steps.
func TopologicalSort(steps []models.Step) ([]models.Step, error) {
	var result []models.Step
	visited := make(map[int]bool)
	temp := make(map[int]bool)
	var visitAll func([]models.Step) error
	var visit func(models.Step) error

	visitAll = func(s []models.Step) error {
		for _, step := range s {
			if err := visit(step); err != nil {
				return err
			}
		}
		return nil
	}

	visit = func(step models.Step) error {
		if temp[step.ID] {
			return fmt.Errorf("cycle detected")
		}
		if !visited[step.ID] {
			temp[step.ID] = true
			// Recursively visit all the step's dependencies
			for _, depID := range step.Dependencies {
				for _, depStep := range steps {
					if depStep.ID == depID {
						if err := visit(depStep); err != nil {
							return err
						}
						break
					}
				}
			}
			visited[step.ID] = true
			temp[step.ID] = false
			result = append(result, step) // Append step instead of prepending
		}
		return nil
	}

	if err := visitAll(steps); err != nil {
		return nil, err
	}

	return result, nil
}

func main() {
	// Initialize the database
	db, err := database.NewDB("meta.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	// Define a task queue size
	const taskQueueSize = 20

	// Initialize the scheduler with the database
	scheduler := scheduler.NewScheduler(db, taskQueueSize)

	// Start the scheduler
	scheduler.Start()
	// Other application logic...

	// Example usage top sort
	steps := []models.Step{
		{ID: 1, Dependencies: []int{2}, Name: "Step1", MapID: 1}, // Step 1 depends on Step 2
		{ID: 2, Dependencies: []int{}, Name: "Step2", MapID: 1},  // Step 2 has no dependencies
	}
	sortedSteps, err := TopologicalSort(steps)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	for _, step := range sortedSteps {
		fmt.Println(step.ID)
	}
}
