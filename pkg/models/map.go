package models

import "time"

// DAG represents a directed acyclic graph of tasks.
type Map struct {
	ID               int
	Name             string
	ScheduleInterval string
	IsActive         bool
	StartDate        time.Time
	LastRun          time.Time
	Steps            []Step // Collection of steps
}

// NewDAG creates and returns a new DAG instance.
func NewMap(name string, scheduleInterval string, startDate time.Time, lastRun time.Time, steps []Step) *Map {
	return &Map{
		Name:             name,
		ScheduleInterval: scheduleInterval,
		IsActive:         true,
		StartDate:        startDate,
		LastRun:          lastRun,
		Steps:            steps,
	}
}
