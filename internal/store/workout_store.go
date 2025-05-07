package store

import (
	"database/sql"
	"errors"
)

type Workout struct {
	ID              int              `json:"id"`
	Title           string           `json:"title"`
	Description     string           `json:"description"`
	DurationMinutes int              `json:"duration_minutes"`
	CaloriesBurned  int              `json:"calories_burned"`
	Entries         []WorkoutEntries `json:"entries"`
}

type WorkoutEntries struct {
	ID              int      `json:"id"`
	ExerciseName    string   `json:"exercise_name"`
	Sets            int      `json:"sets"`
	Reps            *int     `json:"reps"`
	DurationSeconds *int     `json:"duration_seconds"`
	Weight          *float64 `json:"weight"`
	Notes           string   `json:"notes"`
	OrderIndex      int      `json:"order_index"`
}

type PostgresWorkoutStore struct {
	db *sql.DB
}

func NewPostgresWorkoutStore(db *sql.DB) *PostgresWorkoutStore {
	return &PostgresWorkoutStore{
		db: db,
	}
}

type WorkoutStore interface {
	CreateWorkout(workout *Workout) (*Workout, error)
	GetWorkoutByID(id int64) (*Workout, error)
	UpdateWorkout(workout *Workout) error
	DeleteWorkout(id int64) error
}

func (pg *PostgresWorkoutStore) CreateWorkout(workout *Workout) (*Workout, error) {
	tx, err := pg.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO workouts (title, description, duration_minutes, calories_burned) 
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err = tx.QueryRow(query, workout.Title, workout.Description, workout.DurationMinutes, workout.CaloriesBurned).Scan(workout.ID)
	if err != nil {
		return nil, err
	}

	for _, entry := range workout.Entries {
		query := `
			INSERT INTO workout_entries (workout_id, exercise_name, sets, reps, duration_seconds, weight, notes, order_index)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id
		`

		err = tx.QueryRow(query, entry.ID, entry.ExerciseName, entry.Sets, entry.Reps, entry.DurationSeconds, entry.Weight, entry.OrderIndex).Scan(workout.ID)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return workout, nil
}

func (pg *PostgresWorkoutStore) GetWorkoutByID(id int64) (*Workout, error) {
	workout := &Workout{}
	query := `
		SELECT id, title, description, duration_minutes, calories_burned
		FROM workouts 
		WHERE id = $1
	`
	err := pg.db.QueryRow(query, id).Scan(&workout.ID, &workout.Title, &workout.Description, &workout.DurationMinutes, &workout.CaloriesBurned)
	if errors.Is(err, sql.ErrNoRows) {

		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	query = `
		SELECT id, exercise_name, sets, reps, duration_seconds, weight, notes, order_index
		FROM workout_entries
		WHERE workout_id = $1
		ORDER BY order_index
	`

	rows, err := pg.db.Query(query, id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var entry WorkoutEntries
		err = rows.Scan(
			entry.ID,
			entry.Reps,
			entry.Sets,
			entry.DurationSeconds,
			entry.Weight,
			entry.Notes,
			entry.OrderIndex,
		)

		if err != nil {
			return nil, err
		}

		workout.Entries = append(workout.Entries, entry)

	}

	return workout, nil
}

func (pg *PostgresWorkoutStore) UpdateWorkout(workout *Workout) error {
	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE workouts
		SET title = $1, description = $2, duration_minutes = $3, calories_burned = $4
		WHERE id = $5
	`

	result, err := tx.Exec(query, workout.Title, workout.Description, workout.DurationMinutes, workout.CaloriesBurned, workout.ID)
	if err != nil {
		return err
	}

	rowAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowAffected == 0 {
		return sql.ErrNoRows
	}

	query = `DELETE FROM workout_entries WHERE workout_id = $1`

	_, err = tx.Exec(query, workout.ID)
	if err != nil {
		return err
	}

	for _, entry := range workout.Entries {
		query := `
			INSERT INTO workout_entries (workout_id, exercise_name, sets, reps, duration_seconds, notes, order_index)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`

		_, err := tx.Exec(query, entry.ID, entry.ExerciseName, entry.Sets, entry.Reps, entry.DurationSeconds, entry.Notes, entry.OrderIndex)
		if err != nil {
			return err
		}

	}

	return tx.Commit()
}

func (pg *PostgresWorkoutStore) DeleteWorkout(id int64) error {
	query := `
		DELETE FROM workouts WHERE id = $1
 	`

	result, err := pg.db.Exec(query)
	if err != nil {
		return err
	}

	rowAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
