package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/oki-irawan/fem_project/internal/store"
	"net/http"
	"strconv"
)

type WorkoutHandler struct {
	workoutStore store.WorkoutStore
}

func NewWorkoutHandler(workoutStore store.WorkoutStore) *WorkoutHandler {
	return &WorkoutHandler{
		workoutStore: workoutStore,
	}
}

func (wh *WorkoutHandler) HandleGetWorkoutById(w http.ResponseWriter, r *http.Request) {

	paramsWorkoutID := chi.URLParam(r, "id")

	if paramsWorkoutID == "" {
		http.NotFound(w, r)
		return
	}

	workoutId, err := strconv.ParseInt(paramsWorkoutID, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	workout, err := wh.workoutStore.GetWorkoutByID(workoutId)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to fetch the workout", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(workout)
}

func (wh *WorkoutHandler) HandleCreateWorkout(w http.ResponseWriter, r *http.Request) {
	var workout store.Workout

	err := json.NewDecoder(r.Body).Decode(&workout)
	if err != nil {
		fmt.Println("error body request: ")
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	createdWorkout, err := wh.workoutStore.CreateWorkout(&workout)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdWorkout)

	fmt.Fprintf(w, "created Workout\n")
}

func (wh *WorkoutHandler) HandleUpdateWorkoutById(w http.ResponseWriter, r *http.Request) {
	paramsWorkoutID := chi.URLParam(r, "id")

	if paramsWorkoutID == "" {
		http.NotFound(w, r)
		return
	}

	workoutId, err := strconv.ParseInt(paramsWorkoutID, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	existingWorkout, err := wh.workoutStore.GetWorkoutByID(workoutId)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to fetch the workout", http.StatusNotFound)
		return
	}

	if existingWorkout == nil {
		http.NotFound(w, r)
		return
	}

	// At this stage, we assume that the existing workout exists

	var updatedWorkoutRequest struct {
		Title           *string                `json:"title"`
		Description     *string                `json:"description"`
		DurationMinutes *int                   `json:"duration_minutes"`
		CaloriesBurned  *int                   `json:"calories_burned"`
		Entries         []store.WorkoutEntries `json:"entries"`
	}

	err = json.NewDecoder(r.Body).Decode(&updatedWorkoutRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if updatedWorkoutRequest.Title != nil {
		existingWorkout.Title = *updatedWorkoutRequest.Title
	}
	if updatedWorkoutRequest.Description != nil {
		existingWorkout.Description = *updatedWorkoutRequest.Description
	}
	if updatedWorkoutRequest.DurationMinutes != nil {
		existingWorkout.DurationMinutes = *updatedWorkoutRequest.DurationMinutes
	}
	if updatedWorkoutRequest.CaloriesBurned != nil {
		existingWorkout.CaloriesBurned = *updatedWorkoutRequest.CaloriesBurned
	}
	if updatedWorkoutRequest.Entries != nil {
		existingWorkout.Entries = updatedWorkoutRequest.Entries
	}

	err = wh.workoutStore.UpdateWorkout(existingWorkout)
	if err != nil {
		fmt.Println("Updated workout error", err)
		http.Error(w, "Failed to update the workout", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingWorkout)

}

func (wh *WorkoutHandler) HandlerDeleteWorkoutById(w http.ResponseWriter, r *http.Request) {
	paramsWorkoutID := chi.URLParam(r, "id")

	if paramsWorkoutID == "" {
		http.NotFound(w, r)
		return
	}

	workoutId, err := strconv.ParseInt(paramsWorkoutID, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	err = wh.workoutStore.DeleteWorkout(workoutId)

	if errors.Is(err, sql.ErrNoRows) {
		fmt.Println(err)
		http.Error(w, "Workout not found", http.StatusNotFound)
		return
	}

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to delete the workout", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
