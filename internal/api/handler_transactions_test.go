package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"neraca/internal/models"
)

func TestTransactions_FlowAndReversals(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	// 1. Buat dua akun: BCA (1.000.000) dan GoPay (200.000)
	bcaRes := createTestAccount(t, router, "BCA", "operational", 1000000)
	gopayRes := createTestAccount(t, router, "GoPay", "operational", 200000)

	// 2. Catat Pengeluaran: Rp 50.000 dari BCA
	expensePayload := fmt.Sprintf(`{
		"account_id": %d,
		"amount": 50000,
		"type": "expense",
		"description": "Makan Siang Nasi Padang",
		"transaction_date": "2026-09-13"
	}`, bcaRes.ID)

	req := httptest.NewRequest(http.MethodPost, "/api/transactions", bytes.NewBufferString(expensePayload))
	req.Header.Set(headerContentType, mimeApplicationJSON)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 created, got %d: %s", rr.Code, rr.Body.String())
	}

	var expTx models.Transaction
	_ = json.NewDecoder(rr.Body).Decode(&expTx)

	// Cek saldo BCA: 1.000.000 - 50.000 = 950.000
	acc := getTestAccount(t, router, bcaRes.ID)
	if acc.Balance != 950000 {
		t.Fatalf("expected BCA balance 950000, got %f", acc.Balance)
	}

	// 3. Catat Pemasukan: Rp 100.000 ke GoPay
	incomePayload := fmt.Sprintf(`{
		"account_id": %d,
		"amount": 100000,
		"type": "income",
		"description": "Cashback GoPay",
		"transaction_date": "2026-09-13"
	}`, gopayRes.ID)

	req = httptest.NewRequest(http.MethodPost, "/api/transactions", bytes.NewBufferString(incomePayload))
	req.Header.Set(headerContentType, mimeApplicationJSON)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 created, got %d: %s", rr.Code, rr.Body.String())
	}

	// Cek saldo GoPay: 200.000 + 100.000 = 300.000
	acc = getTestAccount(t, router, gopayRes.ID)
	if acc.Balance != 300000 {
		t.Fatalf("expected GoPay balance 300000, got %f", acc.Balance)
	}

	// 4. Catat Transfer: Rp 150.000 dari BCA ke GoPay
	transferPayload := fmt.Sprintf(`{
		"from_account_id": %d,
		"to_account_id": %d,
		"amount": 150000,
		"description": "Top Up GoPay dari BCA",
		"transaction_date": "2026-09-13"
	}`, bcaRes.ID, gopayRes.ID)

	req = httptest.NewRequest(http.MethodPost, "/api/transactions/transfer", bytes.NewBufferString(transferPayload))
	req.Header.Set(headerContentType, mimeApplicationJSON)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 created for transfer, got %d: %s", rr.Code, rr.Body.String())
	}

	var transTx models.Transaction
	_ = json.NewDecoder(rr.Body).Decode(&transTx)

	// Cek saldo BCA: 950.000 - 150.000 = 800.000
	// Cek saldo GoPay: 300.000 + 150.000 = 450.000
	accBCA := getTestAccount(t, router, bcaRes.ID)
	accGoPay := getTestAccount(t, router, gopayRes.ID)
	if accBCA.Balance != 800000 {
		t.Fatalf("expected BCA balance 800000, got %f", accBCA.Balance)
	}
	if accGoPay.Balance != 450000 {
		t.Fatalf("expected GoPay balance 450000, got %f", accGoPay.Balance)
	}

	// 5. GET /api/transactions
	req = httptest.NewRequest(http.MethodGet, "/api/transactions?month=2026-09", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var list models.TransactionsResponse
	_ = json.NewDecoder(rr.Body).Decode(&list)
	if len(list.Transactions) != 3 {
		t.Fatalf("expected 3 transactions, got %d", len(list.Transactions))
	}
	if list.TotalExpense != 50000 {
		t.Fatalf("expected total expense 50000, got %f", list.TotalExpense)
	}
	if list.TotalIncome != 100000 {
		t.Fatalf("expected total income 100000, got %f", list.TotalIncome)
	}

	// 6. Hapus transaksi transfer dan cek saldo ter-revert
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/transactions/%d", transTx.ID), nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204 no content, got %d: %s", rr.Code, rr.Body.String())
	}

	// Saldo BCA kembali ke 950.000, GoPay kembali ke 300.000
	accBCA = getTestAccount(t, router, bcaRes.ID)
	accGoPay = getTestAccount(t, router, gopayRes.ID)
	if accBCA.Balance != 950000 {
		t.Fatalf("expected reverted BCA balance 950000, got %f", accBCA.Balance)
	}
	if accGoPay.Balance != 300000 {
		t.Fatalf("expected reverted GoPay balance 300000, got %f", accGoPay.Balance)
	}

	// 7. Hapus transaksi expense dan cek saldo BCA kembali ke 1.000.000
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/transactions/%d", expTx.ID), nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	accBCA = getTestAccount(t, router, bcaRes.ID)
	if accBCA.Balance != 1000000 {
		t.Fatalf("expected reverted BCA balance 1000000, got %f", accBCA.Balance)
	}
}

func createTestAccount(t *testing.T, router http.Handler, name, group string, balance float64) models.Account {
	t.Helper()
	payload := fmt.Sprintf(`{"name":"%s","account_group":"%s","type":"bank","balance":%f}`, name, group, balance)
	req := httptest.NewRequest(http.MethodPost, "/api/accounts", bytes.NewBufferString(payload))
	req.Header.Set(headerContentType, mimeApplicationJSON)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("failed to create test account: %s", rr.Body.String())
	}

	var a models.Account
	_ = json.NewDecoder(rr.Body).Decode(&a)
	return a
}

func getTestAccount(t *testing.T, router http.Handler, id int64) models.Account {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/accounts", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var res models.AccountsSummaryResponse
	_ = json.NewDecoder(rr.Body).Decode(&res)

	for _, a := range res.OperationalAccounts {
		if a.ID == id {
			return a
		}
	}
	for _, a := range res.PassiveAccounts {
		if a.ID == id {
			return a
		}
	}
	t.Fatalf("account %d not found in summary", id)
	return models.Account{}
}
