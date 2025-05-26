package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/oki-irawan/fem_project/internal/middleware"
	"github.com/oki-irawan/fem_project/internal/store"
	"github.com/oki-irawan/fem_project/internal/utils"
	"log"
	"net/http"
)

type WorkoutHandler struct {
	workoutStore store.WorkoutStore
	logger       *log.Logger
}

func NewWorkoutHandler(workoutStore store.WorkoutStore, logger *log.Logger) *WorkoutHandler {
	return &WorkoutHandler{
		workoutStore: workoutStore,
		logger:       logger,
	}
}

func (wh *WorkoutHandler) HandleGetWorkoutById(w http.ResponseWriter, r *http.Request) {
	workoutId, err := utils.ReadIdParameter(r)
	if err != nil {
		wh.logger.Println("ERROR: read id parameter %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid workout id"})
		return
	}

	workout, err := wh.workoutStore.GetWorkoutByID(workoutId)
	if err != nil {
		wh.logger.Println("ERROR: Failed to fetch the workout %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"workout": workout})
}

func (wh *WorkoutHandler) HandleCreateWorkout(w http.ResponseWriter, r *http.Request) {
	var workout store.Workout

	err := json.NewDecoder(r.Body).Decode(&workout)
	if err != nil {
		wh.logger.Println("ERROR: Decoding Create Workout: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid request body"})
		return
	}

	currentUser := middleware.GetUser(r)
	if currentUser == nil || currentUser == store.AnonymousUser {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "You must be logged in to create a workout"})
		return
	}

	workout.UserID = currentUser.ID

	createdWorkout, err := wh.workoutStore.CreateWorkout(&workout)
	if err != nil {
		wh.logger.Println("ERROR: Creating Workout: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to create workout"})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"workout": createdWorkout})

}

func (wh *WorkoutHandler) HandleUpdateWorkoutById(w http.ResponseWriter, r *http.Request) {
	workoutId, err := utils.ReadIdParameter(r)
	if err != nil {
		wh.logger.Println("ERROR: read id parameter %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid workout id"})
		return
	}

	existingWorkout, err := wh.workoutStore.GetWorkoutByID(workoutId)
	if err != nil {
		wh.logger.Println("ERROR: getWorkoutByID: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
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
		wh.logger.Println("ERROR: Decoding Create Workout: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid request body"})
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

	currentUser := middleware.GetUser(r)
	if currentUser == nil || currentUser == store.AnonymousUser {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "You must be logged in to update a workout"})
		return
	}

	workoutOwner, err := wh.workoutStore.GetWorkoutOwner(workoutId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{"error": "Workout does not exist"})
			return
		}

		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Internal server error"})
		return
	}

	if workoutOwner != currentUser.ID {
		utils.WriteJSON(w, http.StatusForbidden, utils.Envelope{"error": "You are not allowed to update this workout"})
		return
	}

	err = wh.workoutStore.UpdateWorkout(existingWorkout)
	if err != nil {
		wh.logger.Println("ERROR: UpdateWorkout: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to update workout"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"workout": existingWorkout})
}

func (wh *WorkoutHandler) HandlerDeleteWorkoutById(w http.ResponseWriter, r *http.Request) {
	workoutId, err := utils.ReadIdParameter(r)
	if err != nil {
		wh.logger.Println("ERROR: read id parameter %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid workout id"})
		return
	}

	currentUser := middleware.GetUser(r)
	if currentUser == nil || currentUser == store.AnonymousUser {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "You must be logged in to update a workout"})
		return
	}

	workoutOwner, err := wh.workoutStore.GetWorkoutOwner(workoutId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{"error": "Workout does not exist"})
			return
		}

		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Internal server error"})
		return
	}

	if workoutOwner != currentUser.ID {
		utils.WriteJSON(w, http.StatusForbidden, utils.Envelope{"error": "You are not allowed to delete this workout"})
		return
	}

	err = wh.workoutStore.DeleteWorkout(workoutId)

	if errors.Is(err, sql.ErrNoRows) {
		fmt.Println(err)
		http.Error(w, "Workout not found", http.StatusNotFound)
		return
	}

	if err != nil {
		wh.logger.Println("ERROR: DeleteWorkout: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to delete the workout"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
