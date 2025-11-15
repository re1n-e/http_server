package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	type param struct {
		Body string `json:"body"`
	}
	var params param
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errResp{
			Err: fmt.Sprintf("Recived no token from client: %v", err),
		})
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errResp{
			Err: fmt.Sprintf("Unauthorized login: %v", err),
		})
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		errString := fmt.Sprintf("failed to decode params: %v", err)
		writeJSON(w, 500, errResp{Err: errString})
		return
	}
	body, err := validateChirp(params.Body)
	if err != nil {
		writeJSON(w, 500, errResp{Err: err.Error()})
	}
	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		UserID: userId,
		Body:   params.Body,
	})
	if err != nil {
		errString := fmt.Sprintf("failed to create chirp: %v", err)
		writeJSON(w, 500, errResp{Err: errString})
		return
	}
	type Chirp struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}
	writeJSON(w, 201, Chirp{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      body,
		UserId:    userId,
	})
}

func validateChirp(body string) (string, error) {
	if len(body) > 140 {
		return "", fmt.Errorf("chirp is too long")
	}
	return cleanseBody(body), nil
}

func cleanseBody(body string) string {
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	for _, word := range profaneWords {
		re := regexp.MustCompile(`(?i)\b` + word + `\b`)
		body = re.ReplaceAllString(body, "****")
	}
	return body
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Query().Get("author_id")
	st := r.URL.Query().Get("sort")
	var chirps []database.Chirp
	if s != "" {
		userId, err := uuid.Parse(s)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errResp{
				err.Error(),
			})
			return
		}
		chirps, err = cfg.db.GetChirpsByOwnerId(r.Context(), userId)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errResp{
				err.Error(),
			})
			return
		}
	} else {
		var err error
		chirps, err = cfg.db.GetChirps(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errResp{
				Err: fmt.Sprintf("failed to get chirps: %v", err),
			})
			return
		}
	}

	switch st {
	case "asc":
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
		})

	case "desc":
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})

	default:
	}

	type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	var response []Chirp
	for _, c := range chirps {
		response = append(response, Chirp{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserID:    c.UserID,
		})
	}

	writeJSON(w, http.StatusOK, response)
}

func (cfg *apiConfig) getChirpById(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	id := strings.TrimPrefix(path, "/api/chirps/")
	chirpId, err := uuid.Parse(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp{
			Err: fmt.Sprintf("failed to parse chirp id: %v", err),
		})
		return
	}
	chirp, err := cfg.db.GetChirpById(r.Context(), chirpId)
	if err != nil {
		writeJSON(w, 404, errResp{
			Err: fmt.Sprintf("failed to retrive chirp by id: %v", err),
		})
		return
	}

	type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}
	writeJSON(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeJSON(w, 401, errResp{
			Err: err.Error(),
		})
		return
	}

	userId, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		writeJSON(w, 401, errResp{
			Err: err.Error(),
		})
		return
	}

	path := r.URL.Path
	chirpId := strings.TrimPrefix(path, "/api/chirps/")
	chirpPostId, err := uuid.Parse(chirpId)
	if err != nil {
		writeJSON(w, 401, errResp{
			Err: err.Error(),
		})
		return
	}

	chirpOwnerId, err := cfg.db.GetChirpOwnerId(r.Context(), chirpPostId)
	if err != nil {
		writeJSON(w, 404, errResp{
			Err: err.Error(),
		})
		return
	}

	if chirpOwnerId != userId {
		writeJSON(w, 403, errResp{
			Err: "Unauthorized access",
		})
		return
	}

	if err := cfg.db.DeleteChirp(r.Context(), chirpPostId); err != nil {
		writeJSON(w, 403, errResp{
			Err: "Chirp not deleted",
		})
		return
	}

	writeJSON(w, 204, nil)
}
