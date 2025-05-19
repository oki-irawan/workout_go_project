package store

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("pgx", "host=localhost user=postgres password=postgres dbname=postgres port=5433 sslmode=disable")
	if err != nil {
		t.Fatalf("open test db error happend : %v", err)
	}

	err = Migrate(db, "../../migrations/")
	if err != nil {
		t.Fatalf("migrate error happend : %v", err)
	}

	_, err = db.Exec(`TRUNCATE  workouts, workout_entries CASCADE`)
	if err != nil {
		t.Fatalf("truncate error happend : %v", err)
	}

	return db
}

func TestCreateWorkout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewPostgresWorkoutStore(db)

	test := []struct {
		name    string
		workout *Workout
		wantErr bool
	}{
		{
			name: "Valid Create Workout",
			workout: &Workout{
				Title:           "Push Day",
				Description:     "Upper Body Day",
				DurationMinutes: 60,
				CaloriesBurned:  200,
				Entries: []WorkoutEntries{
					{
						ExerciseName: "Bench Press",
						Sets:         3,
						Reps:         intPtr(10),
						Weight:       floatPtr(100.34),
						Notes:        "Warm up properly",
						OrderIndex:   1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid Create Workout",
			workout: &Workout{
				Title:           "Full Body",
				Description:     "Complete workout day",
				DurationMinutes: 90,
				CaloriesBurned:  500,
				Entries: []WorkoutEntries{
					{
						ExerciseName: "Bench Press",
						Sets:         6,
						Reps:         intPtr(7),
						Weight:       floatPtr(123.34),
						Notes:        "Warm up properly",
						OrderIndex:   1,
					},
					{
						ExerciseName: "Plank",
						Sets:         6,
						Reps:         intPtr(7),
						Notes:        "Keep Form",
						OrderIndex:   2,
					},
					{
						ExerciseName:    "Squads",
						Sets:            6,
						Reps:            intPtr(7),
						DurationSeconds: intPtr(120),
						Weight:          floatPtr(53.34),
						Notes:           "Full depth",
						OrderIndex:      3,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			createdWorkout, err := store.CreateWorkout(tt.workout)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.workout.Title, createdWorkout.Title)
			assert.Equal(t, tt.workout.Description, createdWorkout.Description)
			assert.Equal(t, tt.workout.DurationMinutes, createdWorkout.DurationMinutes)
			assert.Equal(t, tt.workout.CaloriesBurned, createdWorkout.CaloriesBurned)

			retrieve, err := store.GetWorkoutByID(int64(createdWorkout.ID))
			require.NoError(t, err)

			assert.Equal(t, createdWorkout.ID, retrieve.ID)
			assert.Equal(t, len(tt.workout.Entries), len(retrieve.Entries))

			for i, entry := range tt.workout.Entries {
				assert.Equal(t, entry.ExerciseName, retrieve.Entries[i].ExerciseName)
				assert.Equal(t, entry.Sets, retrieve.Entries[i].Sets)
				assert.Equal(t, entry.Reps, retrieve.Entries[i].Reps)
				assert.Equal(t, entry.Weight, retrieve.Entries[i].Weight)
				assert.Equal(t, entry.Notes, retrieve.Entries[i].Notes)
				assert.Equal(t, entry.OrderIndex, retrieve.Entries[i].OrderIndex)
			}

		})
	}

}

func intPtr(i int) *int {
	return &i
}

func floatPtr(i float64) *float64 {
	return &i
}
