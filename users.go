package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type Params struct {
		Passwd string `json:"password"`
		Email  string `json:"email"`
	}

	type respError struct {
		Err string `json:"error"`
	}

	var param Params
	if err := json.NewDecoder(r.Body).Decode(&param); err != nil {
		errString := fmt.Sprintf("Unable to decode email string: %v", err)
		writeJSON(w, http.StatusInternalServerError, respError{Err: errString})
		return
	}
	hashedPasswd, err := auth.HashPassword(param.Passwd)
	if err != nil {
		errString := fmt.Sprintf("Failed to hash passwd: %v", err)
		writeJSON(w, http.StatusInternalServerError, respError{Err: errString})
		return
	}
	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          param.Email,
		HashedPassword: hashedPasswd,
	})
	if err != nil {
		errString := fmt.Sprintf("Failed to create a new user: %v", err)
		writeJSON(w, http.StatusInternalServerError, respError{Err: errString})
		return
	}

	writeJSON(w, 201, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed.Bool,
	})
}

func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errResp{Err: fmt.Sprintf("No bearer token given: %v", err)})
		return
	}

	userId, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errResp{Err: fmt.Sprintf(": %v", err)})
		return
	}

	type Params struct {
		Passwd string `json:"password"`
		Email  string `json:"email"`
	}

	type respError struct {
		Err string `json:"error"`
	}

	var param Params
	if err := json.NewDecoder(r.Body).Decode(&param); err != nil {
		errString := fmt.Sprintf("Unable to decode email string: %v", err)
		writeJSON(w, http.StatusInternalServerError, respError{Err: errString})
		return
	}
	hashedPasswd, err := auth.HashPassword(param.Passwd)
	if err != nil {
		errString := fmt.Sprintf("Failed to hash passwd: %v", err)
		writeJSON(w, http.StatusInternalServerError, respError{Err: errString})
		return
	}

	user, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          param.Email,
		HashedPassword: hashedPasswd,
		ID:             userId,
	})
	if err != nil {
		errString := fmt.Sprintf("Failed update users: %v", err)
		writeJSON(w, http.StatusInternalServerError, respError{Err: errString})
		return
	}

	writeJSON(w, http.StatusOK, User{
		ID:          userId,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed.Bool,
	})
}

func (cfg *apiConfig) resetUsers(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		errString := "403 Forbidden"
		writeJSON(w, 500, errResp{Err: errString})
		return
	}
	if err := cfg.db.ResetUsers(r.Context()); err != nil {
		errString := fmt.Sprintf("Failed to reset the user: %v", err)
		writeJSON(w, 500, errResp{Err: errString})
		return
	}
}

func (cfg *apiConfig) loginUser(w http.ResponseWriter, r *http.Request) {
	type Param struct {
		Passwd string `json:"password"`
		Email  string `json:"email"`
	}
	var param Param
	if err := json.NewDecoder(r.Body).Decode(&param); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp{
			Err: fmt.Sprintf("failed to parse the parameters: %v", err),
		})
		return
	}
	expires := time.Hour
	user, err := cfg.db.GetUserByMail(r.Context(), param.Email)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errResp{
			Err: fmt.Sprintf("Incorrect email or password: %v", err),
		})
		return
	}
	isOwner, err := auth.CheckPasswordHash(param.Passwd, user.HashedPassword)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errResp{
			Err: fmt.Sprintf("failed to check passwd hash: %v", err),
		})
		return
	}
	if !isOwner {
		writeJSON(w, http.StatusUnauthorized, errResp{
			Err: "Incorrect email or password",
		})
		return
	}

	authToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, expires)
	if err != nil {
		writeJSON(w, 401, errResp{
			Err: fmt.Sprintf("failed to create token: %v", err),
		})
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp{
			Err: fmt.Sprintf("Failed to create refresh token: %v", err),
		})
		return
	}

	if err := cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
		RevokedAt: sql.NullTime{},
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp{
			Err: fmt.Sprintf("Failed to store refresh token: %v", err),
		})
		return
	}

	writeJSON(w, http.StatusOK, User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		AuthToken:    authToken,
		RefreshToken: refreshToken,
		IsChirpyRed:  user.IsChirpyRed.Bool,
	})
}

func (cfg *apiConfig) HandleUpdateChirpyRed(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetApiKey(r.Header)
	if err != nil {
		writeJSON(w, 401, errResp{
			Err: err.Error(),
		})
	}

	if apiKey != cfg.polkaKey {
		writeJSON(w, 401, errResp{
			Err: "Unauthorized access",
		})
	}

	type Param struct {
		Event string `json:"event"`
		Data  struct {
			UserId string `json:"user_id"`
		} `json:"data"`
	}

	var param Param
	if err := json.NewDecoder(r.Body).Decode(&param); err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp{
			Err: fmt.Sprintf("failed to decode params: %v", err),
		})
		return
	}

	if param.Event != "user.upgraded" {
		writeJSON(w, 204, nil)
		return
	}

	userId, err := uuid.Parse(param.Data.UserId)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp{
			Err: fmt.Sprintf("failed to parse uuid: %v", err),
		})
		return
	}

	if err := cfg.db.UpdateChirpyRed(r.Context(), userId); err != nil {
		writeJSON(w, 404, errResp{
			Err: err.Error(),
		})
		return
	}

	writeJSON(w, 204, nil)
}
