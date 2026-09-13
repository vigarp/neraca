package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSettings_ResetFinancialData_Success(t *testing.T) {
	router, db := setupTestRouter(t)

	// Seed dummy account and transaction
	res, err := db.Exec(
		"INSERT INTO accounts (name, type, balance, account_group) VALUES (?, ?, ?, ?)",
		"Test Bank", "bank", 500000, "operational",
	)
	if err != nil {
		t.Fatalf("failed to insert dummy account: %v", err)
	}
	accID, _ := res.LastInsertId()

	_, err = db.Exec(
		"INSERT INTO transactions (account_id, amount, type, description, transaction_date) VALUES (?, ?, ?, ?, '2026-09-13')",
		accID, 50000, "expense", "Makan Siang",
	)
	if err != nil {
		t.Fatalf("failed to insert dummy transaction: %v", err)
	}

	// Call POST /api/settings/reset-data
	req := httptest.NewRequest(http.MethodPost, "/api/settings/reset-data", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Verify accounts and transactions are cleared
	var accCount, txCount, catCount, userCount, sessionCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM accounts").Scan(&accCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&txCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&catCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&sessionCount)

	if accCount != 0 {
		t.Errorf("expected 0 accounts, got %d", accCount)
	}
	if txCount != 0 {
		t.Errorf("expected 0 transactions, got %d", txCount)
	}
	if catCount == 0 {
		t.Errorf("expected default categories to be re-seeded, got 0")
	}
	if userCount != 1 {
		t.Errorf("expected 1 user preserved, got %d", userCount)
	}
	if sessionCount != 1 {
		t.Errorf("expected 1 session preserved, got %d", sessionCount)
	}
}

func TestSettings_ResetFinancialData_Unauthorized(t *testing.T) {
	router, _ := setupCleanTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/settings/reset-data", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized, got %d", rec.Code)
	}
}

func TestSettings_DownloadBackup_Success(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/settings/backup", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/x-sqlite3" {
		t.Errorf("expected Content-Type application/x-sqlite3, got %s", contentType)
	}

	contentDisposition := rec.Header().Get("Content-Disposition")
	if !strings.Contains(contentDisposition, "attachment; filename=\"neraca_backup_") {
		t.Errorf("unexpected Content-Disposition: %s", contentDisposition)
	}

	body := rec.Body.Bytes()
	if len(body) == 0 {
		t.Errorf("expected non-empty backup database file")
	}

	sqliteHeader := []byte("SQLite format 3\x00")
	if !bytes.HasPrefix(body, sqliteHeader) {
		t.Errorf("expected file to have SQLite header")
	}
}

func TestSettings_DownloadBackup_Unauthorized(t *testing.T) {
	router, _ := setupCleanTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/settings/backup", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized, got %d", rec.Code)
	}
}
