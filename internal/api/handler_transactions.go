package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"neraca/internal/database"
	"neraca/internal/models"
)

const (
	errInvalidTransactionID = "Invalid transaction ID"
	errTransactionNotFound  = "Transaction not found"
	queryAccountDeduct      = "UPDATE accounts SET balance = balance - ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	queryAccountAdd         = "UPDATE accounts SET balance = balance + ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
)

func buildTransactionsFilter(r *http.Request) (string, []any) {
	conditions := []string{"1=1"}
	args := []any{}

	if month := r.URL.Query().Get("month"); month != "" {
		conditions = append(conditions, "strftime('%Y-%m', t.transaction_date) = ?")
		args = append(args, month)
	}
	if accID := r.URL.Query().Get("account_id"); accID != "" {
		conditions = append(conditions, "(t.account_id = ? OR t.to_account_id = ?)")
		args = append(args, accID, accID)
	}
	if catID := r.URL.Query().Get("category_id"); catID != "" {
		conditions = append(conditions, "t.category_id = ?")
		args = append(args, catID)
	}
	if tType := r.URL.Query().Get("type"); tType != "" {
		conditions = append(conditions, "t.type = ?")
		args = append(args, tType)
	}

	return strings.Join(conditions, " AND "), args
}

func handleGetTransactions(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filterClause, args := buildTransactionsFilter(r)

		query := fmt.Sprintf(`
		SELECT 
			t.id, t.account_id, a.name AS account_name,
			t.to_account_id, COALESCE(ta.name, '') AS to_account_name,
			t.category_id, COALESCE(c.name, '') AS category_name, COALESCE(c.pillar, '') AS pillar,
			t.amount, t.type, COALESCE(t.description, '') AS description,
			t.transaction_date, t.created_at
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		LEFT JOIN accounts ta ON t.to_account_id = ta.id
		LEFT JOIN categories c ON t.category_id = c.id
		WHERE %s
		ORDER BY t.transaction_date DESC, t.id DESC;
		`, filterClause)

		rows, err := db.Query(query, args...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		res := models.TransactionsResponse{
			Transactions: []models.Transaction{},
		}

		for rows.Next() {
			var t models.Transaction
			if err := rows.Scan(
				&t.ID, &t.AccountID, &t.AccountName,
				&t.ToAccountID, &t.ToAccountName,
				&t.CategoryID, &t.CategoryName, &t.Pillar,
				&t.Amount, &t.Type, &t.Description,
				&t.TransactionDate, &t.CreatedAt,
			); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if t.Type == "expense" {
				res.TotalExpense += t.Amount
			} else if t.Type == "income" {
				res.TotalIncome += t.Amount
			}

			res.Transactions = append(res.Transactions, t)
		}

		writeJSON(w, http.StatusOK, res)
	}
}

func applyTransactionBalance(tx *sql.Tx, accountID int64, amount float64, tType string) error {
	var query string
	if tType == "expense" {
		query = queryAccountDeduct
	} else if tType == "income" {
		query = queryAccountAdd
	} else {
		return errors.New("invalid transaction type for balance update")
	}

	res, err := tx.Exec(query, amount, accountID)
	if err != nil {
		return err
	}
	rowsAff, _ := res.RowsAffected()
	if rowsAff == 0 {
		return errors.New(errAccountNotFound)
	}
	return nil
}

func handleCreateTransaction(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateTransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, errInvalidPayload, http.StatusBadRequest)
			return
		}

		if req.AccountID <= 0 || req.Amount <= 0 {
			http.Error(w, "Akun dan jumlah nominal harus valid (> 0)", http.StatusBadRequest)
			return
		}
		if req.Type != "expense" && req.Type != "income" {
			http.Error(w, "Tipe transaksi harus 'expense' atau 'income'", http.StatusBadRequest)
			return
		}
		if req.TransactionDate == "" {
			req.TransactionDate = time.Now().Format(time.DateOnly)
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		if err := applyTransactionBalance(tx, req.AccountID, req.Amount, req.Type); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		insertQuery := `
		INSERT INTO transactions (account_id, category_id, amount, type, description, transaction_date)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id, created_at;
		`
		var t models.Transaction
		t.AccountID = req.AccountID
		t.CategoryID = req.CategoryID
		t.Amount = req.Amount
		t.Type = req.Type
		t.Description = strings.TrimSpace(req.Description)
		t.TransactionDate = req.TransactionDate

		err = tx.QueryRow(insertQuery, t.AccountID, t.CategoryID, t.Amount, t.Type, t.Description, t.TransactionDate).
			Scan(&t.ID, &t.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusCreated, t)
	}
}

func handleCreateTransfer(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateTransferRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, errInvalidPayload, http.StatusBadRequest)
			return
		}

		if req.FromAccountID <= 0 || req.ToAccountID <= 0 || req.Amount <= 0 {
			http.Error(w, "Akun asal, akun tujuan, dan nominal harus valid", http.StatusBadRequest)
			return
		}
		if req.FromAccountID == req.ToAccountID {
			http.Error(w, "Akun asal dan akun tujuan tidak boleh sama", http.StatusBadRequest)
			return
		}
		if req.TransactionDate == "" {
			req.TransactionDate = time.Now().Format(time.DateOnly)
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// Potong saldo akun asal
		if err := applyTransactionBalance(tx, req.FromAccountID, req.Amount, "expense"); err != nil {
			http.Error(w, "Gagal memotong saldo akun asal: "+err.Error(), http.StatusBadRequest)
			return
		}
		// Tambah saldo akun tujuan
		if err := applyTransactionBalance(tx, req.ToAccountID, req.Amount, "income"); err != nil {
			http.Error(w, "Gagal menambah saldo akun tujuan: "+err.Error(), http.StatusBadRequest)
			return
		}

		insertQuery := `
		INSERT INTO transactions (account_id, to_account_id, category_id, amount, type, description, transaction_date)
		VALUES (?, ?, ?, ?, 'transfer', ?, ?)
		RETURNING id, created_at;
		`
		var t models.Transaction
		t.AccountID = req.FromAccountID
		t.ToAccountID = &req.ToAccountID
		t.CategoryID = req.CategoryID
		t.Amount = req.Amount
		t.Type = "transfer"
		t.Description = strings.TrimSpace(req.Description)
		t.TransactionDate = req.TransactionDate

		err = tx.QueryRow(insertQuery, t.AccountID, t.ToAccountID, t.CategoryID, t.Amount, t.Description, t.TransactionDate).
			Scan(&t.ID, &t.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusCreated, t)
	}
}

func revertTransactionBalance(tx *sql.Tx, t models.Transaction) error {
	switch t.Type {
	case "expense":
		_, err := tx.Exec(queryAccountAdd, t.Amount, t.AccountID)
		return err
	case "income":
		_, err := tx.Exec(queryAccountDeduct, t.Amount, t.AccountID)
		return err
	case "transfer":
		if t.ToAccountID != nil {
			if _, err := tx.Exec(queryAccountAdd, t.Amount, t.AccountID); err != nil {
				return err
			}
			_, err := tx.Exec(queryAccountDeduct, t.Amount, *t.ToAccountID)
			return err
		}
	}
	return nil
}

func handleDeleteTransaction(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, errInvalidTransactionID, http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		var t models.Transaction
		query := `SELECT id, account_id, to_account_id, amount, type FROM transactions WHERE id = ?`
		err = tx.QueryRow(query, id).Scan(&t.ID, &t.AccountID, &t.ToAccountID, &t.Amount, &t.Type)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, errTransactionNotFound, http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := revertTransactionBalance(tx, t); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := tx.Exec("DELETE FROM transactions WHERE id = ?", id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
