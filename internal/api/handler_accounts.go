package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"neraca/internal/database"
	"neraca/internal/models"
)

const (
	headerContentType   = "Content-Type"
	mimeApplicationJSON = "application/json"
	errInvalidPayload   = "Invalid request payload"
	errInvalidAccountID = "Invalid account ID"
	errAccountNotFound  = "Account not found"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set(headerContentType, mimeApplicationJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func handleGetAccounts(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := `
		SELECT id, name, account_group, type, balance, currency, institution, is_active, created_at, updated_at
		FROM accounts
		WHERE is_active = 1
		ORDER BY account_group ASC, name ASC;
		`
		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		res := models.AccountsSummaryResponse{
			OperationalAccounts: []models.Account{},
			PassiveAccounts:     []models.Account{},
			EmergencyAccounts:   []models.Account{},
		}

		for rows.Next() {
			var a models.Account
			if err := rows.Scan(
				&a.ID, &a.Name, &a.AccountGroup, &a.Type, &a.Balance,
				&a.Currency, &a.Institution, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
			); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			switch a.AccountGroup {
			case "passive_asset":
				res.PassiveAccounts = append(res.PassiveAccounts, a)
				res.TotalPassiveAssets += a.Balance
			case "emergency":
				res.EmergencyAccounts = append(res.EmergencyAccounts, a)
				res.TotalEmergencyBalance += a.Balance
			default:
				res.OperationalAccounts = append(res.OperationalAccounts, a)
				res.TotalOperationalBalance += a.Balance
			}
		}

		res.TotalNetWorth = res.TotalOperationalBalance + res.TotalPassiveAssets + res.TotalEmergencyBalance

		writeJSON(w, http.StatusOK, res)
	}
}

func validateAndNormalizeAccount(req *models.CreateAccountRequest) error {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return errors.New("Nama akun wajib diisi")
	}
	if req.AccountGroup == "" {
		req.AccountGroup = "operational"
	}
	if req.Type == "" {
		req.Type = "bank"
	}
	if req.Currency == "" {
		req.Currency = "IDR"
	}
	return nil
}

func insertAccountWithInitialValuation(db *database.DB, req models.CreateAccountRequest) (models.Account, error) {
	tx, err := db.Begin()
	if err != nil {
		return models.Account{}, err
	}
	defer tx.Rollback()

	insertQuery := `
	INSERT INTO accounts (name, account_group, type, balance, currency, institution)
	VALUES (?, ?, ?, ?, ?, ?)
	RETURNING id, created_at, updated_at;
	`
	a := models.Account{
		Name:         req.Name,
		AccountGroup: req.AccountGroup,
		Type:         req.Type,
		Balance:      req.Balance,
		Currency:     req.Currency,
		Institution:  req.Institution,
		IsActive:     true,
	}

	err = tx.QueryRow(insertQuery, a.Name, a.AccountGroup, a.Type, a.Balance, a.Currency, a.Institution).
		Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return models.Account{}, err
	}

	if a.AccountGroup == "passive_asset" && a.Balance > 0 {
		valQuery := `
		INSERT INTO asset_valuations (account_id, previous_balance, new_balance, difference, notes)
		VALUES (?, ?, ?, ?, ?);
		`
		if _, err = tx.Exec(valQuery, a.ID, 0, a.Balance, a.Balance, "Saldo awal pendaftaran aset"); err != nil {
			return models.Account{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return models.Account{}, err
	}

	return a, nil
}

func handleCreateAccount(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, errInvalidPayload, http.StatusBadRequest)
			return
		}

		if err := validateAndNormalizeAccount(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		a, err := insertAccountWithInitialValuation(db, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusCreated, a)
	}
}

func handleUpdateAccount(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, errInvalidAccountID, http.StatusBadRequest)
			return
		}

		var req models.UpdateAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, errInvalidPayload, http.StatusBadRequest)
			return
		}

		query := `
		UPDATE accounts
		SET name = ?, type = ?, institution = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND is_active = 1
		RETURNING id, name, account_group, type, balance, currency, institution, is_active, created_at, updated_at;
		`
		var a models.Account
		err = db.QueryRow(query, req.Name, req.Type, req.Institution, id).
			Scan(&a.ID, &a.Name, &a.AccountGroup, &a.Type, &a.Balance, &a.Currency, &a.Institution, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, errAccountNotFound, http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, a)
	}
}

func handleDeleteAccount(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, errInvalidAccountID, http.StatusBadRequest)
			return
		}

		// Soft delete
		res, err := db.Exec(`UPDATE accounts SET is_active = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rowsAff, _ := res.RowsAffected()
		if rowsAff == 0 {
			http.Error(w, errAccountNotFound, http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// Revalue: Penyesuaian nilai pasar untuk Aset Pasif (Investasi, Reksadana, Emas, dll)
func handleRevalueAccount(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, errInvalidAccountID, http.StatusBadRequest)
			return
		}

		var req models.RevalueAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, errInvalidPayload, http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		var currentBalance float64
		var accountGroup string
		err = tx.QueryRow(`SELECT balance, account_group FROM accounts WHERE id = ? AND is_active = 1`, id).
			Scan(&currentBalance, &accountGroup)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, errAccountNotFound, http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		diff := req.NewBalance - currentBalance

		// Update saldo akun
		var updated models.Account
		updateQuery := `
		UPDATE accounts
		SET balance = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING id, name, account_group, type, balance, currency, institution, is_active, created_at, updated_at;
		`
		err = tx.QueryRow(updateQuery, req.NewBalance, id).
			Scan(&updated.ID, &updated.Name, &updated.AccountGroup, &updated.Type, &updated.Balance,
				&updated.Currency, &updated.Institution, &updated.IsActive, &updated.CreatedAt, &updated.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Catat riwayat valuasi
		valQuery := `
		INSERT INTO asset_valuations (account_id, previous_balance, new_balance, difference, notes)
		VALUES (?, ?, ?, ?, ?);
		`
		if _, err = tx.Exec(valQuery, id, currentBalance, req.NewBalance, diff, req.Notes); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, updated)
	}
}
