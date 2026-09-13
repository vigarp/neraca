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

func TestAnalytics_DetermineEffectiveDaysElapsed(t *testing.T) {
	cycleStart := time.Date(2026, 9, 5, 0, 0, 0, 0, time.Local)

	// 1. Kasus Pengguna Baru: Mulai catat di tengah siklus (13 Sept), gajian tgl 5
	today1 := time.Date(2026, 9, 13, 0, 0, 0, 0, time.Local)
	elapsed1 := determineEffectiveDaysElapsed(today1, cycleStart, "2026-09-13")
	if elapsed1 != 1 {
		t.Fatalf("expected 1 day elapsed for new user starting today, got %d", elapsed1)
	}

	// 2. Kasus Hari ke-2: Belanja mulai kemarin (13 Sept), hari ini 14 Sept
	today2 := time.Date(2026, 9, 14, 0, 0, 0, 0, time.Local)
	elapsed2 := determineEffectiveDaysElapsed(today2, cycleStart, "2026-09-13")
	if elapsed2 != 2 {
		t.Fatalf("expected 2 days elapsed, got %d", elapsed2)
	}

	// 3. Kasus Pengguna Lama: Sudah catat sejak awal gajian (5 Sept)
	elapsed3 := determineEffectiveDaysElapsed(today1, cycleStart, "2026-09-05")
	if elapsed3 != 8 {
		t.Fatalf("expected 8 days elapsed for full cycle tracking, got %d", elapsed3)
	}

	// 4. Kasus Belum Ada Transaksi Pengeluaran
	elapsed4 := determineEffectiveDaysElapsed(today1, cycleStart, "")
	if elapsed4 != 8 {
		t.Fatalf("expected fallback to 8 calendar days when empty, got %d", elapsed4)
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

	// Income dalam siklus = 5.000.000
	incomePayload := fmt.Sprintf(`{
		"account_id": %d,
		"category_id": 16,
		"amount": 5000000,
		"type": "income",
		"description": "Gaji Bulanan",
		"transaction_date": "2026-08-26"
	}`, bca.ID)
	req = httptest.NewRequest(http.MethodPost, "/api/transactions", strings.NewReader(incomePayload))
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

	// Verifikasi Remain Funds: BCA mula-mula 4.800.000 - 500.000 + 5.000.000 = 9.300.000
	// Portofolio aset pasif 15.000.000 TIDAK boleh masuk ke remain funds!
	if res.RemainFunds != 9300000 {
		t.Fatalf("expected remain funds 9300000, got %f", res.RemainFunds)
	}

	// Verifikasi Cycle Income: 5.000.000
	if res.CycleIncome != 5000000 {
		t.Fatalf("expected cycle income 5000000, got %f", res.CycleIncome)
	}

	// Verifikasi Next Payday: 12 hari
	if res.NextPaydayRemain != 12 {
		t.Fatalf("expected 12 days remain, got %d", res.NextPaydayRemain)
	}

	// Verifikasi Grand Total Expenses: 500.000
	if res.GrandTotalExpenses != 500000 {
		t.Fatalf("expected grand total expenses 500000, got %f", res.GrandTotalExpenses)
	}

	// Verifikasi Prospect Daily Limit: 9.300.000 / 12 = 775000
	expectedDailyLimit := 775000.0
	if res.ProspectDailyLimit != expectedDailyLimit {
		t.Fatalf("expected prospect daily limit %f, got %f", expectedDailyLimit, res.ProspectDailyLimit)
	}

	// Verifikasi Days Elapsed (Smart Tracked): 13 hari (dari pengeluaran terlama 2026-09-01 s.d. 2026-09-13)
	if res.DaysElapsed != 13 {
		t.Fatalf("expected 13 days elapsed, got %d", res.DaysElapsed)
	}

	// Verifikasi Average Daily Expense: 500.000 / 13 = 38461.54
	expectedAvgExpense := 38461.54
	if res.AverageDailyExpense != expectedAvgExpense {
		t.Fatalf("expected avg daily expense %f, got %f", expectedAvgExpense, res.AverageDailyExpense)
	}

	// Status burn rate harus safe
	if res.BurnRateStatus != "safe" {
		t.Fatalf("expected burn rate status 'safe', got %s", res.BurnRateStatus)
	}

	// Verifikasi transaction count pada category breakdown
	if len(res.CategoryBreakdown) == 0 {
		t.Fatalf("expected category breakdown not empty")
	}
	for _, cat := range res.CategoryBreakdown {
		if cat.TransactionCount <= 0 {
			t.Fatalf("expected transaction count > 0 for category %s, got %d", cat.Name, cat.TransactionCount)
		}
	}
}
