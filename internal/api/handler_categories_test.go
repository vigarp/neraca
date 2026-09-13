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

func TestCategories_SeedAndCRUD(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()

	// 1. GET /api/categories - Harus sudah ter-seed default categories
	req := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var res models.CategoriesSummaryResponse
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(res.AllCategories) == 0 {
		t.Fatalf("expected seeded categories, got 0")
	}
	if len(res.NeedsCategories) == 0 {
		t.Fatalf("expected needs categories, got 0")
	}
	if len(res.WantsCategories) == 0 {
		t.Fatalf("expected wants categories, got 0")
	}
	if len(res.SavingsCategories) == 0 {
		t.Fatalf("expected savings categories, got 0")
	}

	// 2. POST /api/categories - Buat kategori baru
	newCat := `{"name":"Kursus Online","type":"expense","pillar":"needs","icon":"GraduationCap","color":"#3B82F6"}`
	req = httptest.NewRequest(http.MethodPost, "/api/categories", bytes.NewBufferString(newCat))
	req.Header.Set(headerContentType, mimeApplicationJSON)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var createdCat models.Category
	_ = json.NewDecoder(rr.Body).Decode(&createdCat)
	if createdCat.ID == 0 {
		t.Fatalf("expected created category ID > 0")
	}

	// 3. PUT /api/categories/{id} - Update nama/pilar
	updatePayload := `{"name":"Kursus & Buku","pillar":"wants","icon":"Book","color":"#A855F7"}`
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/categories/%d", createdCat.ID), bytes.NewBufferString(updatePayload))
	req.Header.Set(headerContentType, mimeApplicationJSON)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// 4. DELETE /api/categories/{id}
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/categories/%d", createdCat.ID), nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}
