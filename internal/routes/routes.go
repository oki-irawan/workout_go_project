package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/oki-irawan/fem_project/internal/app"
)

func SetupRoutes(app *app.Application) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/health", app.HealthCheck)

	r.Get("/workouts/{id}", app.WorkoutHandler.GetWorkoutById)
	r.Post("/workouts", app.WorkoutHandler.CreateWorkout)

	return r
}
