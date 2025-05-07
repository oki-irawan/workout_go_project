package app

import (
	"database/sql"
	"fmt"
	"github.com/oki-irawan/fem_project/internal/api"
	"github.com/oki-irawan/fem_project/internal/store"
	"github.com/oki-irawan/fem_project/migrations"
	"log"
	"net/http"
	"os"
)

type Application struct {
	Logger         *log.Logger
	WorkoutHandler *api.WorkoutHandler
	DB             *sql.DB
}

func NewApplication() (*Application, error) {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	pgDB, err := store.Open()
	if err != nil {
		return nil, err
	}

	err = store.MigrateFS(pgDB, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	//store
	workoutStore := store.NewPostgresWorkoutStore(pgDB)

	//api
	workoutHandler := api.NewWorkoutHandler(workoutStore)

	app := &Application{
		Logger:         logger,
		WorkoutHandler: workoutHandler,
		DB:             pgDB,
	}

	return app, nil
}

func (a *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Status is availaible\n")
}
