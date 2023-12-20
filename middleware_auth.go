package main

import (
	"net/http"

	"github.com/fredshema/rssagg/internal/auth"
	"github.com/fredshema/rssagg/internal/database"
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (cfg *apiConfig) authMiddleware(next authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := auth.GetAPIKey(r.Header)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}

		user, err := cfg.DB.GetUserByAPIKey(r.Context(), apiKey)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Cannot get user")
			return
		}

		next(w, r, user)
	}
}
