package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"neraca/internal/database"
	"neraca/internal/models"
)

func handleGetSettings(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := models.SettingsResponse{
			PaydayDate:          25,
			MonthlyIncomeBudget: 6000000.0,
		}

		rows, err := db.Query(`SELECT key, value FROM settings WHERE key IN ('payday_date', 'monthly_income_budget');`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var key, val string
				if err := rows.Scan(&key, &val); err == nil {
					switch key {
					case "payday_date":
						if d, err := strconv.Atoi(val); err == nil {
							res.PaydayDate = d
						}
					case "monthly_income_budget":
						if b, err := strconv.ParseFloat(val, 64); err == nil {
							res.MonthlyIncomeBudget = b
						}
					}
				}
			}
		}

		writeJSON(w, http.StatusOK, res)
	}
}

func handleUpdateSettings(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.UpdateSettingsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, errInvalidPayload, http.StatusBadRequest)
			return
		}

		if req.PaydayDate < 1 || req.PaydayDate > 31 {
			http.Error(w, "Tanggal gajian harus antara 1 sampai 31", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		query := `INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?);`

		if _, err := tx.Exec(query, "payday_date", strconv.Itoa(req.PaydayDate)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := tx.Exec(query, "monthly_income_budget", strconv.FormatFloat(req.MonthlyIncomeBudget, 'f', -1, 64)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, req)
	}
}

func handleResetFinancialData(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.ResetFinancialData(); err != nil {
			http.Error(w, "Gagal mereset data keuangan: "+err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "Data keuangan berhasil di-reset ke kondisi awal",
		})
	}
}
