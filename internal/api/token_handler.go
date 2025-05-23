package api

import (
	"encoding/json"
	"github.com/oki-irawan/fem_project/internal/store"
	"github.com/oki-irawan/fem_project/internal/tokens"
	"github.com/oki-irawan/fem_project/internal/utils"
	"log"
	"net/http"
	"time"
)

type TokenHandler struct {
	tokenStore store.TokenStore
	userStore  store.UserStore
	logger     *log.Logger
}

type createTokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewTokenHandler(tokenStore store.TokenStore, userStore store.UserStore, logger *log.Logger) *TokenHandler {
	return &TokenHandler{
		tokenStore: tokenStore,
		userStore:  userStore,
		logger:     logger,
	}
}

func (t *TokenHandler) HandlerCreateToken(w http.ResponseWriter, r *http.Request) {
	var req createTokenRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		t.logger.Printf("ERROR: Decoding Create Token: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid request body"})
		return
	}

	// get user from username
	user, err := t.userStore.GetUserByUsername(req.Username)
	if err != nil {
		t.logger.Printf("ERROR: GetUserByUsername: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Internal server error"})
		return
	}

	// user doesn't exist
	if user == nil {
		t.logger.Printf("ERROR: GetUserByUsername: user not found %v", err)
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Invalid username and password"})
		return
	}

	// compare password
	passwordDoMatch, err := user.PasswordHash.Matches(req.Password)
	if err != nil {
		t.logger.Printf("ERROR: PasswordHash.Matches: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Internal server error"})
		return
	}

	// password doesn't match
	if !passwordDoMatch {
		t.logger.Printf("ERROR: PasswordHash.Matches: password not match %v", err)
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Invalid username and password"})
		return
	}

	token, err := t.tokenStore.CreateNewToken(int64(user.ID), 24*time.Hour, tokens.ScopeAuth)
	if err != nil {
		t.logger.Printf("ERROR: CreateNewToken: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Internal server error"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"token": token})

}
