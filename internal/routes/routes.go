package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/oki-irawan/fem_project/internal/app"
)

func SetupRoutes(app *app.Application) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/health", app.HealthCheck)

	r.Get("/workouts/{id}", app.WorkoutHandler.HandleGetWorkoutById)
	r.Post("/workouts", app.WorkoutHandler.HandleCreateWorkout)
	r.Put("/workouts/{id}", app.WorkoutHandler.HandleUpdateWorkoutById)
	r.Delete("/workouts/{id}", app.WorkoutHandler.HandlerDeleteWorkoutById)

	r.Post("/users", app.UserHandler.HandleCreateUser)

	return r
}
