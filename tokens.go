package main

import (
	"chirpy/internal/auth"
	"net/http"
	"time"
)

func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	refToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errResp{
			Err: err.Error(),
		})
		return
	}
	token, err := cfg.db.GetRefreshTokenByRefToken(r.Context(), refToken)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errResp{
			Err: err.Error(),
		})
		return
	}
	if token.RevokedAt.Valid {
		writeJSON(w, http.StatusUnauthorized, errResp{
			Err: "token revoked",
		})
	}

	type Resp struct {
		Token string `json:"token"`
	}

	newToken, err := auth.MakeJWT(token.UserID, cfg.jwtSecret, time.Hour)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errResp{
			Err: err.Error(),
		})
		return
	}
	writeJSON(w, 200, Resp{
		Token: newToken,
	})
}

func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	refToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errResp{
			Err: err.Error(),
		})
		return
	}
	if err := cfg.db.RevokeRefreshToken(r.Context(), refToken); err != nil {
		writeJSON(w, http.StatusUnauthorized, errResp{
			Err: err.Error(),
		})
		return
	}
	writeJSON(w, 204, struct{}{})
}
