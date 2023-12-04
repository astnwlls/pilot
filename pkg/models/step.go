package models

import "time"

// Task represents an individual task in a DAG.
type Step struct {
	ID           int
	Name         string
	MapID        int
	State        string
	Command      string
	StartDate    time.Time
	EndDate      time.Time
	Dependencies []int // IDs of dependent tasks
}

// NewTask creates and returns a new Task instance.
func NewStep(name string, mapID int) *Step {
	return &Step{
		Name:  name,
		MapID: mapID,
		State: "pending",
	}
}
