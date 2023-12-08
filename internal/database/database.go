package database

import (
	"database/sql"
	"log"

	"pilot/pkg/models"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	createTables(db)

	if err != nil {
		return nil, err
	}

	return &DB{conn: db}, nil
}

func createMapsTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS maps (
        id INT PRIMARY KEY,
        name VARCHAR(255),
        schedule_interval VARCHAR(255),
        is_active BOOLEAN,
        start_date TIMESTAMP
    );`
	_, err := db.Exec(query)
	return err
}

func createStepsTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS steps (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        map_id INT,
        name VARCHAR(255),
        state VARCHAR(255),
        start_date TIMESTAMP,
        end_date TIMESTAMP,
        dependencies TEXT, 
        FOREIGN KEY (map_id) REFERENCES maps(id)
    );`
	_, err := db.Exec(query)
	return err
}

func createTables(db *sql.DB) error {
	var createErr error

	if err := createMapsTable(db); err != nil {
		createErr = err
	}

	if err := createStepsTable(db); err != nil {
		createErr = err
	}

	return createErr
}

func (db *DB) AddMap(m models.Map) error {
	query := `INSERT INTO maps (name, schedule_interval, is_active, start_date) VALUES (?, ?, ?, ?)`
	_, err := db.conn.Exec(query, m.Name, m.ScheduleInterval, m.IsActive, m.StartDate)
	return err
}

func (db *DB) AddStep(task *models.Step) (int, error) {
	// INSERT query without RETURNING clause
	insertQuery := `INSERT INTO steps (name, map_id, state, start_date, end_date) VALUES (?, ?, ?, ?, ?)`
	result, err := db.conn.Exec(insertQuery, task.Name, task.MapID, task.State, task.StartDate, task.EndDate)
	if err != nil {
		log.Printf("Error adding step to database: %v", err)
		return 0, err
	}

	// Retrieve the last insert id
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		return 0, err
	}

	return int(id), nil
}

// GetActivemaps retrieves all active maps from the database
func (db *DB) GetActiveMaps() ([]models.Map, error) {
	var maps []models.Map
	query := `SELECT id, name, schedule_interval, is_active, start_date FROM maps WHERE is_active = true`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m models.Map
		if err := rows.Scan(&m.ID, &m.Name, &m.ScheduleInterval, &m.IsActive, &m.StartDate); err != nil {
			return nil, err
		}
		maps = append(maps, m)
	}

	return maps, nil
}

// Getmap retrieves a map by its ID
func GetMap(db *sql.DB, id int) (*models.Map, error) {
	m := &models.Map{}
	query := `SELECT id, name, schedule_interval, is_active, start_date FROM maps WHERE id = ?`
	err := db.QueryRow(query, id).Scan(&m.ID, &m.Name, &m.ScheduleInterval, &m.IsActive, &m.StartDate)
	return m, err
}

// GetStepsBymapID retrieves all steps for a given map
func (db *DB) GetStepsByMapID(id int) ([]models.Step, error) {
	var steps []models.Step
	query := `SELECT id, name, map_id, state, start_date, end_date FROM steps WHERE map_id = ?`
	rows, err := db.conn.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task models.Step
		if err := rows.Scan(&task.ID, &task.Name, &task.MapID, &task.State, &task.StartDate, &task.EndDate); err != nil {
			return nil, err
		}
		steps = append(steps, task)
	}

	return steps, nil
}

// GetStepByID retrieves a specific step by its ID
func (db *DB) GetStepByID(id int) (*models.Step, error) {
	var step models.Step
	query := `SELECT id, name, map_id, state, start_date, end_date FROM steps WHERE id = ?`
	row := db.conn.QueryRow(query, id)

	err := row.Scan(&step.ID, &step.Name, &step.MapID, &step.State, &step.StartDate, &step.EndDate)
	if err != nil {
		if err == sql.ErrNoRows {
			// No rows were returned, handle this case as needed
			return nil, err
		}
		return nil, err
	}

	return &step, nil
}

// Updatemap modifies an existing map
func (db *DB) UpdateMap(m models.Map) error {
	query := `UPDATE maps SET name = ?, schedule_interval = ?, is_active = ?, start_date = ? WHERE id = ?`
	_, err := db.conn.Exec(query, m.Name, m.ScheduleInterval, m.IsActive, m.StartDate, m.ID)
	return err
}

// UpdateStep modifies an existing task
func (db *DB) UpdateStep(step models.Step) error {
	query := `UPDATE steps SET name = ?, map_id = ?, state = ?, start_date = ?, end_date = ? WHERE id = ?`
	result, err := db.conn.Exec(query, step.Name, step.MapID, step.State, step.StartDate, step.EndDate, step.ID)
	if err != nil {
		// Detailed logging of the error
		log.Printf("Failed to update step: %v, error: %v\n", step, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to retrieve affected rows for step: %v, error: %v\n", step, err)
		return err
	}
	if rowsAffected == 0 {
		log.Printf("No rows affected for step: %v\n", step)
	}

	return err
}

// Deletemap removes a map from the database
func (db *DB) DeleteMap(id int) error {
	query := `DELETE FROM maps WHERE id = ?`
	_, err := db.conn.Exec(query, id)
	return err
}

// DeleteStep removes a task from the database
func (db *DB) DeleteStep(id int) error {
	query := `DELETE FROM steps WHERE id = ?`
	_, err := db.conn.Exec(query, id)
	return err
}

// DependenciesMet checks if all dependencies for a given task are completed
func (db *DB) DependenciesMet(task models.Step) (bool, error) {
	for _, depID := range task.Dependencies {
		var state string
		query := `SELECT state FROM steps WHERE id = ?`
		err := db.conn.QueryRow(query, depID).Scan(&state)
		if err != nil {
			return false, err
		}

		if state != "completed" {
			return false, nil
		}
	}
	return true, nil
}
