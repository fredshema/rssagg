package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fredshema/rssagg/internal/database"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

func (apiCfg *apiConfig) handlerCreateFeedFollow(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		FeedID uuid.UUID `json:"feed_id"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	feedFollow, err := apiCfg.DB.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    params.FeedID,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot create feed")
		return
	}

	respondWithJSON(w, http.StatusCreated, databaseFeedFollowToFeedFollow(feedFollow))
}

func (apiCfg *apiConfig) handlerGetFeedFollows(w http.ResponseWriter, r *http.Request, user database.User) {
	feedFollows, err := apiCfg.DB.GetFeedFollows(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot get feed follows")
		return
	}

	respondWithJSON(w, http.StatusOK, databaseFeedFollowsToFeedFollows(feedFollows))
}

func (apiCfg *apiConfig) handlerDeleteFeedFollow(w http.ResponseWriter, r *http.Request, user database.User) {
	feedFollowID := chi.URLParam(r, "feedFollowID")
	if feedFollowID == "" {
		respondWithError(w, http.StatusBadRequest, "feedFollowID is required")
		return
	}

	err := apiCfg.DB.DeleteFeedFollow(r.Context(), database.DeleteFeedFollowParams{
		ID:     uuid.MustParse(feedFollowID),
		UserID: user.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Cannot delete feed follow")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
