package api

import (
	"encoding/json"
	"net/http"
	"time"

	"neraca/internal/database"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	App       string    `json:"app"`
	Timestamp time.Time `json:"timestamp"`
	Database  string    `json:"database"`
}

func handleHealth(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "connected"
		if err := db.Ping(); err != nil {
			dbStatus = "error: " + err.Error()
		}

		res := HealthResponse{
			Status:    "ok",
			App:       "neraca",
			Timestamp: time.Now(),
			Database:  dbStatus,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}
}
