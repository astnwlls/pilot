package scheduler

import (
	"log"
	"pilot/internal/database"
	"pilot/pkg/models"
	"time"

	"github.com/robfig/cron/v3"
	// other imports
)

type Scheduler struct {
	db        *database.DB
	TaskQueue chan models.Step
	nowFunc   func() time.Time
	// other fields
}

func NewScheduler(db *database.DB, taskQueueSize int) *Scheduler {
	return &Scheduler{
		db:        db,
		TaskQueue: make(chan models.Step, taskQueueSize),
		nowFunc:   time.Now,
		// initialize other fields
	}
}

func (s *Scheduler) SetNowFunc(f func() time.Time) {
	s.nowFunc = f
}

func (s *Scheduler) Start() {
	for {
		// 1. Fetch all active DAGs from the database
		maps, err := s.db.GetActiveMaps()
		if err != nil {
			// handle error
		}

		// 2. Iterate through each Map and determine if any tasks are ready to run
		for _, m := range maps {
			if s.IsTimeToRun(m) {
				tasks, err := s.db.GetStepsByMapID(m.ID)
				if err != nil {
					// handle error
				}

				for _, task := range tasks {
					depsMet, err := s.db.DependenciesMet(task)
					if err != nil {
						// handle error
					}

					if depsMet {
						s.queueTask(task)
					}
				}
			}
		}

		// Add a sleep interval to avoid constant database querying
		time.Sleep(1 * time.Minute)
	}
}

func (s *Scheduler) ScheduleMap(m models.Map) {
	if s.IsTimeToRun(m) {
		for _, step := range m.Steps {
			if s.dependenciesMet(step) {
				s.TaskQueue <- step
			} else {
				log.Printf("Dependencies not met for step: %+v", step)
			}
		}
	} else {
		log.Println("Map is not scheduled to run at this time.")
	}
}

func (s *Scheduler) dependenciesMet(step models.Step) bool {
	for _, depID := range step.Dependencies {
		if !s.isDependencyMet(depID) {
			return false
		}
	}
	return true
}

func (s *Scheduler) isDependencyMet(depID int) bool {
	depStep, err := s.db.GetStepByID(depID)
	if err != nil {
		log.Println("Error getting step:", err)
		return false
	}
	return depStep.State == "completed"
}

func (s *Scheduler) IsTimeToRun(m models.Map) bool {
	now := s.nowFunc()
	schedule, err := cron.ParseStandard(m.ScheduleInterval)
	if err != nil {
		log.Fatalf("Failed to parse cron schedule: %v", err)
	}

	// Get the next scheduled run time
	nextRun := schedule.Next(m.LastRun)

	// Check if the next run time is now or in the past
	return now.After(nextRun) || now.Equal(nextRun)
}

func (s *Scheduler) queueTask(step models.Step) {
	s.TaskQueue <- step
}
