package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"neraca/internal/models"
)

func TestAnalytics_PaydayCycleCalculations(t *testing.T) {
	// 1. Tanggal 13 Sept 2026, gajian tanggal 25
	now := time.Date(2026, 9, 13, 10, 0, 0, 0, time.Local)
	cycleStart, nextPayday, daysRemain, daysElapsed := calculatePaydayCycle(now, 25)

	if daysRemain != 12 {
		t.Fatalf("expected 12 days remain, got %d", daysRemain)
	}
	if daysElapsed != 19 {
		t.Fatalf("expected 19 days elapsed, got %d", daysElapsed)
	}
	if nextPayday.Format("2006-01-02") != "2026-09-25" {
		t.Fatalf("expected next payday 2026-09-25, got %s", nextPayday.Format("2006-01-02"))
	}
	if cycleStart.Format("2006-01-02") != "2026-08-25" {
		t.Fatalf("expected cycle start 2026-08-25, got %s", cycleStart.Format("2006-01-02"))
	}

	// 2. Tepat hari gajian 25 Sept 2026
	nowPayday := time.Date(2026, 9, 25, 8, 0, 0, 0, time.Local)
	cycleStart2, nextPayday2, daysRemain2, daysElapsed2 := calculatePaydayCycle(nowPayday, 25)

	if nextPayday2.Format("2006-01-02") != "2026-10-25" {
		t.Fatalf("expected next payday 2026-10-25, got %s", nextPayday2.Format("2006-01-02"))
	}
	if cycleStart2.Format("2006-01-02") != "2026-09-25" {
		t.Fatalf("expected cycle start 2026-09-25, got %s", cycleStart2.Format("2006-01-02"))
	}
	if daysRemain2 != 30 {
		t.Fatalf("expected 30 days remain until Oct 25, got %d", daysRemain2)
	}
	if daysElapsed2 != 1 {
		t.Fatalf("expected 1 day elapsed on payday, got %d", daysElapsed2)
	}

	// 3. Tanggal gajian 31 di bulan Februari
	feb := time.Date(2026, 2, 10, 0, 0, 0, 0, time.Local)
	_, nextFebPayday, _, _ := calculatePaydayCycle(feb, 31)
	if nextFebPayday.Day() != 28 {
		t.Fatalf("expected clamped payday 28 in Feb 2026, got %d", nextFebPayday.Day())
	}
}

func TestAnalytics_BurnRateEndpoint(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	// 1. Buat akun BCA (operasional) = 4.800.000
	bca := createTestAccount(t, router, "BCA Utama", "operational", 4800000)
	// 2. Buat aset Reksadana (aset pasif) = 15.000.000
	createTestAccount(t, router, "Makmur Reksadana", "passive_asset", 15000000)

	// 3. Catat transaksi expense
	// Needs = 300.000
	expensePayload1 := fmt.Sprintf(`{
		"account_id": %d,
		"category_id": 1,
		"amount": 300000,
		"type": "expense",
		"description": "Belanja Bulanan Pokok",
		"transaction_date": "2026-09-01"
	}`, bca.ID)
	req := httptest.NewRequest(http.MethodPost, "/api/transactions", strings.NewReader(expensePayload1))
	req.Header.Set(headerContentType, mimeApplicationJSON)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Wants = 200.000
	expensePayload2 := fmt.Sprintf(`{
		"account_id": %d,
		"category_id": 7,
		"amount": 200000,
		"type": "expense",
		"description": "Nongkrong & Kopi",
		"transaction_date": "2026-09-05"
	}`, bca.ID)
	req = httptest.NewRequest(http.MethodPost, "/api/transactions", strings.NewReader(expensePayload2))
	req.Header.Set(headerContentType, mimeApplicationJSON)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// 4. Panggil /api/analytics/burn-rate?date=2026-09-13
	req = httptest.NewRequest(http.MethodGet, "/api/analytics/burn-rate?date=2026-09-13", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var res models.BurnRateAnalyticsResponse
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verifikasi Remain Funds: BCA mula-mula 4.800.000 - 500.000 = 4.300.000
	// Portofolio aset pasif 15.000.000 TIDAK boleh masuk ke remain funds!
	if res.RemainFunds != 4300000 {
		t.Fatalf("expected remain funds 4300000, got %f", res.RemainFunds)
	}

	// Verifikasi Next Payday: 12 hari
	if res.NextPaydayRemain != 12 {
		t.Fatalf("expected 12 days remain, got %d", res.NextPaydayRemain)
	}

	// Verifikasi Grand Total Expenses: 500.000
	if res.GrandTotalExpenses != 500000 {
		t.Fatalf("expected grand total expenses 500000, got %f", res.GrandTotalExpenses)
	}

	// Verifikasi Prospect Daily Limit: 4.300.000 / 12 = 358333.33
	expectedDailyLimit := 358333.33
	if res.ProspectDailyLimit != expectedDailyLimit {
		t.Fatalf("expected prospect daily limit %f, got %f", expectedDailyLimit, res.ProspectDailyLimit)
	}

	// Status burn rate harus safe
	if res.BurnRateStatus != "safe" {
		t.Fatalf("expected burn rate status 'safe', got %s", res.BurnRateStatus)
	}
}
