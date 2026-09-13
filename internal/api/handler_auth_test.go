package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"neraca/internal/models"
)

func TestAuth_SetupAndDuplicateCheck(t *testing.T) {
	router, db := setupCleanTestRouter(t)
	defer db.Close()

	// 1. Awalnya status harus belum diinisialisasi
	req := httptest.NewRequest(http.MethodGet, "/api/auth/status", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	var statusRes models.AuthStatusResponse
	if err := json.NewDecoder(rr.Body).Decode(&statusRes); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if statusRes.Initialized || statusRes.Authenticated {
		t.Fatalf("expected initialized=false, authenticated=false, got init=%v auth=%v",
			statusRes.Initialized, statusRes.Authenticated)
	}

	// 2. Lakukan setup akun pertama
	setupPayload := `{"username": "vigarp", "password": "supersecretpassword"}`
	req = httptest.NewRequest(http.MethodPost, "/api/auth/setup", strings.NewReader(setupPayload))
	req.Header.Set(headerContentType, mimeApplicationJSON)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on setup, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verifikasi cookie sesi terpasang
	cookies := rr.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == cookieSessionName {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Fatalf("expected session cookie set after setup")
	}

	// 3. Coba lakukan setup ulang (harus gagal 400 Bad Request)
	duplicateSetup := `{"username": "hacker", "password": "hackerpassword"}`
	req = httptest.NewRequest(http.MethodPost, "/api/auth/setup", strings.NewReader(duplicateSetup))
	req.Header.Set(headerContentType, mimeApplicationJSON)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request on duplicate setup, got %d", rr.Code)
	}

	// 4. Periksa status dengan cookie
	req = httptest.NewRequest(http.MethodGet, "/api/auth/status", nil)
	req.AddCookie(sessionCookie)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var authStatus models.AuthStatusResponse
	_ = json.NewDecoder(rr.Body).Decode(&authStatus)
	if !authStatus.Initialized || !authStatus.Authenticated || authStatus.Username != "vigarp" {
		t.Fatalf("expected initialized=true, authenticated=true, username=vigarp, got %+v", authStatus)
	}
}

func TestAuth_LoginLogoutFlow(t *testing.T) {
	router, db := setupCleanTestRouter(t)
	defer db.Close()

	// 1. Setup akun
	setupPayload := `{"username": "vigarp", "password": "supersecretpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/setup", strings.NewReader(setupPayload))
	req.Header.Set(headerContentType, mimeApplicationJSON)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("setup failed: %d", rr.Code)
	}

	// 2. Login dengan password salah
	wrongLogin := `{"username": "vigarp", "password": "wrongpassword"}`
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(wrongLogin))
	req.Header.Set(headerContentType, mimeApplicationJSON)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized on wrong password, got %d", rr.Code)
	}

	// 3. Login dengan kredensial benar + remember me
	correctLogin := `{"username": "vigarp", "password": "supersecretpassword", "remember_me": true}`
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(correctLogin))
	req.Header.Set(headerContentType, mimeApplicationJSON)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on correct login, got %d: %s", rr.Code, rr.Body.String())
	}

	var sessionCookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == cookieSessionName {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatalf("expected session cookie on successful login")
	}

	// 4. Logout
	req = httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.AddCookie(sessionCookie)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on logout, got %d", rr.Code)
	}

	// 5. Cek status setelah logout (harus unauthenticated)
	req = httptest.NewRequest(http.MethodGet, "/api/auth/status", nil)
	req.AddCookie(sessionCookie)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var statusAfterLogout models.AuthStatusResponse
	_ = json.NewDecoder(rr.Body).Decode(&statusAfterLogout)
	if statusAfterLogout.Authenticated {
		t.Fatalf("expected unauthenticated after logout")
	}
}

func TestAuth_MiddlewareProtection(t *testing.T) {
	router, db := setupCleanTestRouter(t)
	defer db.Close()

	// 1. Coba akses endpoint finansial terlindungi tanpa cookie
	req := httptest.NewRequest(http.MethodGet, "/api/accounts", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized without auth cookie, got %d", rr.Code)
	}

	// 2. Setup user dan dapatkan cookie
	setupPayload := `{"username": "vigarp", "password": "supersecretpassword"}`
	req = httptest.NewRequest(http.MethodPost, "/api/auth/setup", strings.NewReader(setupPayload))
	req.Header.Set(headerContentType, mimeApplicationJSON)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	var sessionCookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == cookieSessionName {
			sessionCookie = c
			break
		}
	}

	// 3. Akses endpoint finansial dengan cookie valid (harus 200 OK)
	req = httptest.NewRequest(http.MethodGet, "/api/accounts", nil)
	req.AddCookie(sessionCookie)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK with valid auth cookie, got %d: %s", rr.Code, rr.Body.String())
	}
}
