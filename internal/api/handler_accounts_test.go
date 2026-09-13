package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"neraca/internal/config"
	"neraca/internal/database"
	"neraca/internal/models"
)

type authTestWrapper struct {
	handler http.Handler
	cookie  *http.Cookie
}

func (w *authTestWrapper) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if _, err := req.Cookie(cookieSessionName); err != nil {
		req.AddCookie(w.cookie)
	}
	w.handler.ServeHTTP(rw, req)
}

func setupCleanTestRouter(t *testing.T) (http.Handler, *database.DB) {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_neraca.db")

	db, err := database.Connect(dbPath)
	if err != nil {
		t.Fatalf("failed to connect test db: %v", err)
	}

	cfg := &config.Config{
		Port:   "8088",
		DBPath: dbPath,
		Env:    "test",
	}

	router := NewRouter(cfg, db, nil)
	return router, db
}

func setupTestRouter(t *testing.T) (http.Handler, *database.DB) {
	t.Helper()
	router, db := setupCleanTestRouter(t)

	res, err := db.Exec(
		"INSERT INTO users (username, password_hash) VALUES (?, ?);",
		"testadmin", "$2a$10$7EqJtq98hPqEX7fNZaFWoOZhB4J8Q3qNnJ3eL4w5g6h7i8j9k0l1m",
	)
	if err != nil {
		t.Fatalf("failed to seed test user: %v", err)
	}
	userID, _ := res.LastInsertId()

	rawToken, tokenHash, _ := generateSecureToken()
	expiresAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	_, err = db.Exec(
		"INSERT INTO sessions (user_id, token_hash, user_agent, ip_address, expires_at) VALUES (?, ?, ?, ?, ?);",
		userID, tokenHash, "TestAgent", "127.0.0.1", expiresAt,
	)
	if err != nil {
		t.Fatalf("failed to seed test session: %v", err)
	}

	cookie := &http.Cookie{
		Name:  cookieSessionName,
		Value: rawToken,
		Path:  "/",
	}

	return &authTestWrapper{handler: router, cookie: cookie}, db
}

func TestAccounts_CRUDAndRevaluation(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	// 1. Tambah Akun Kas Operasional (BCA)
	bcaPayload := `{"name":"BCA Utama","account_group":"operational","type":"bank","balance":4800000,"institution":"BCA"}`
	req := httptest.NewRequest(http.MethodPost, "/api/accounts", bytes.NewBufferString(bcaPayload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// 2. Tambah Akun Kas Operasional (GoPay)
	gopayPayload := `{"name":"GoPay","account_group":"operational","type":"ewallet","balance":200000,"institution":"GoTo"}`
	req = httptest.NewRequest(http.MethodPost, "/api/accounts", bytes.NewBufferString(gopayPayload))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// 3. Tambah Aset Pasif (Reksadana Makmur)
	makmurPayload := `{"name":"Reksadana Pasar Uang","account_group":"passive_asset","type":"mutual_fund","balance":15000000,"institution":"Makmur"}`
	req = httptest.NewRequest(http.MethodPost, "/api/accounts", bytes.NewBufferString(makmurPayload))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var makmur models.Account
	if err := json.NewDecoder(rr.Body).Decode(&makmur); err != nil {
		t.Fatalf("failed to decode created account: %v", err)
	}

	// 4. Cek Ringkasan Akun (GET /api/accounts)
	req = httptest.NewRequest(http.MethodGet, "/api/accounts", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var summary models.AccountsSummaryResponse
	if err := json.NewDecoder(rr.Body).Decode(&summary); err != nil {
		t.Fatalf("failed to decode accounts summary: %v", err)
	}

	if len(summary.OperationalAccounts) != 2 {
		t.Errorf("expected 2 operational accounts, got %d", len(summary.OperationalAccounts))
	}
	if len(summary.PassiveAccounts) != 1 {
		t.Errorf("expected 1 passive account, got %d", len(summary.PassiveAccounts))
	}
	if summary.TotalOperationalBalance != 5000000.0 {
		t.Errorf("expected operational balance 5,000,000, got %f", summary.TotalOperationalBalance)
	}
	if summary.TotalPassiveAssets != 15000000.0 {
		t.Errorf("expected passive assets 15,000,000, got %f", summary.TotalPassiveAssets)
	}
	if summary.TotalNetWorth != 20000000.0 {
		t.Errorf("expected total net worth 20,000,000, got %f", summary.TotalNetWorth)
	}

	// 5. Uji Revaluasi Aset Pasif (Naik Rp 500.000 menjadi Rp 15.500.000)
	revalURL := fmt.Sprintf("/api/accounts/%d/revalue", makmur.ID)
	revalPayload := `{"new_balance":15500000,"notes":"Kenaikan return pasar akhir bulan"}`
	req = httptest.NewRequest(http.MethodPost, revalURL, bytes.NewBufferString(revalPayload))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 on revaluation, got %d: %s", rr.Code, rr.Body.String())
	}

	// 6. Cek kembali Ringkasan Akun: Pastikan Kas Operasional TETAP STERIL (5jt), Net Worth naik ke 20.5jt
	req = httptest.NewRequest(http.MethodGet, "/api/accounts", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var summaryAfterReval models.AccountsSummaryResponse
	_ = json.NewDecoder(rr.Body).Decode(&summaryAfterReval)

	if summaryAfterReval.TotalOperationalBalance != 5000000.0 {
		t.Errorf("operational balance was contaminated! expected 5,000,000, got %f", summaryAfterReval.TotalOperationalBalance)
	}
	if summaryAfterReval.TotalPassiveAssets != 15500000.0 {
		t.Errorf("expected passive assets 15,500,000, got %f", summaryAfterReval.TotalPassiveAssets)
	}
	if summaryAfterReval.TotalNetWorth != 20500000.0 {
		t.Errorf("expected total net worth 20,500,000, got %f", summaryAfterReval.TotalNetWorth)
	}
}

func TestSettings_GetAndUpdate(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	// 1. Ambil setting default
	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var s models.SettingsResponse
	if err := json.NewDecoder(rr.Body).Decode(&s); err != nil {
		t.Fatalf("failed to decode settings: %v", err)
	}
	if s.PaydayDate != 25 {
		t.Errorf("expected default payday date 25, got %d", s.PaydayDate)
	}

	// 2. Ubah tanggal gajian jadi 28 dan budget 7jt
	updatePayload := `{"payday_date":28,"monthly_income_budget":7000000}`
	req = httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewBufferString(updatePayload))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 on update, got %d", rr.Code)
	}

	// 3. Ambil lagi untuk verifikasi tersimpan
	req = httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var sUpdated models.SettingsResponse
	_ = json.NewDecoder(rr.Body).Decode(&sUpdated)

	if sUpdated.PaydayDate != 28 {
		t.Errorf("expected updated payday date 28, got %d", sUpdated.PaydayDate)
	}
	if sUpdated.MonthlyIncomeBudget != 7000000.0 {
		t.Errorf("expected updated income budget 7,000,000, got %f", sUpdated.MonthlyIncomeBudget)
	}
}
