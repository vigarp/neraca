package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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

func handleDownloadBackup(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tempFile, err := os.CreateTemp("", "neraca-backup-*.db")
		if err != nil {
			http.Error(w, "Gagal membuat berkas sementara: "+err.Error(), http.StatusInternalServerError)
			return
		}
		tempPath := tempFile.Name()
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		defer os.Remove(tempPath)

		sanitizedPath := strings.ReplaceAll(tempPath, "'", "''")
		if _, err := db.Exec(fmt.Sprintf("VACUUM INTO '%s';", sanitizedPath)); err != nil {
			http.Error(w, "Gagal melakukan backup database: "+err.Error(), http.StatusInternalServerError)
			return
		}

		backupFile, err := os.Open(tempPath)
		if err != nil {
			http.Error(w, "Gagal membaca berkas backup: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer backupFile.Close()

		fileInfo, err := backupFile.Stat()
		if err != nil {
			http.Error(w, "Gagal membaca info berkas backup: "+err.Error(), http.StatusInternalServerError)
			return
		}

		filename := fmt.Sprintf("neraca_backup_%s.db", time.Now().Format("20060102_150405"))
		w.Header().Set("Content-Type", "application/x-sqlite3")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

		http.ServeContent(w, r, filename, fileInfo.ModTime(), backupFile)
	}
}
