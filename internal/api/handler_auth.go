package api

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"neraca/internal/database"
	"neraca/internal/models"
)

const (
	cookieSessionName        = "neraca_session"
	sessionDurationStandard  = 24 * time.Hour
	sessionDurationRemember  = 30 * 24 * time.Hour
	errUnauthorizedMsg       = "Unauthorized"
	errInvalidCredentialsMsg = "Username atau password salah"
)

type loginRateLimiter struct {
	mu          sync.Mutex
	attempts    map[string][]time.Time
	maxAttempts int
	window      time.Duration
}

var globalLoginLimiter = newLoginRateLimiter(5, 1*time.Minute)

func newLoginRateLimiter(max int, window time.Duration) *loginRateLimiter {
	return &loginRateLimiter{
		attempts:    make(map[string][]time.Time),
		maxAttempts: max,
		window:      window,
	}
}

func (l *loginRateLimiter) isAllowed(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	valid := make([]time.Time, 0, len(l.attempts[ip]))
	for _, t := range l.attempts[ip] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	l.attempts[ip] = valid

	return len(valid) < l.maxAttempts
}

func (l *loginRateLimiter) recordFailure(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.attempts[ip] = append(l.attempts[ip], time.Now())
}

func (l *loginRateLimiter) reset(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, ip)
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func hashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

func generateSecureToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	rawToken := hex.EncodeToString(bytes)
	tokenHash := hashToken(rawToken)
	return rawToken, tokenHash, nil
}

func setSessionCookie(w http.ResponseWriter, rawToken string, duration time.Duration) {
	maxAge := int(duration.Seconds())
	http.SetCookie(w, &http.Cookie{
		Name:     cookieSessionName,
		Value:    rawToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Tetap false agar bekerja saat akses via HTTP IP di VPS atau lokal
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
		Expires:  time.Now().Add(duration),
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieSessionName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func createSession(db *database.DB, userID int64, duration time.Duration, r *http.Request) (string, error) {
	rawToken, tokenHash, err := generateSecureToken()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(duration).Format(time.RFC3339)
	userAgent := r.UserAgent()
	ipAddress := r.RemoteAddr

	query := `
	INSERT INTO sessions (user_id, token_hash, user_agent, ip_address, expires_at)
	VALUES (?, ?, ?, ?, ?);
	`
	_, err = db.Exec(query, userID, tokenHash, userAgent, ipAddress, expiresAt)
	if err != nil {
		return "", err
	}

	return rawToken, nil
}

func validateSessionToken(db *database.DB, rawToken string) (*models.User, error) {
	if rawToken == "" {
		return nil, sql.ErrNoRows
	}

	tokenHash := hashToken(rawToken)
	query := `
	SELECT u.id, u.username
	FROM sessions s
	JOIN users u ON s.user_id = u.id
	WHERE s.token_hash = ?
	  AND s.expires_at > CURRENT_TIMESTAMP;
	`
	var user models.User
	err := db.QueryRow(query, tokenHash).Scan(&user.ID, &user.Username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func countUsers(db *database.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users;").Scan(&count)
	return count, err
}

func handleAuthStatus(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCount, err := countUsers(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if userCount == 0 {
			writeJSON(w, http.StatusOK, models.AuthStatusResponse{
				Initialized:   false,
				Authenticated: false,
			})
			return
		}

		cookie, err := r.Cookie(cookieSessionName)
		if err != nil {
			writeJSON(w, http.StatusOK, models.AuthStatusResponse{
				Initialized:   true,
				Authenticated: false,
			})
			return
		}

		user, err := validateSessionToken(db, cookie.Value)
		if err != nil {
			writeJSON(w, http.StatusOK, models.AuthStatusResponse{
				Initialized:   true,
				Authenticated: false,
			})
			return
		}

		writeJSON(w, http.StatusOK, models.AuthStatusResponse{
			Initialized:   true,
			Authenticated: true,
			Username:      user.Username,
		})
	}
}

func handleAuthSetup(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCount, err := countUsers(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if userCount > 0 {
			http.Error(w, "Setup akun pengelola sudah selesai", http.StatusBadRequest)
			return
		}

		var req models.SetupAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Payload tidak valid", http.StatusBadRequest)
			return
		}

		req.Username = strings.TrimSpace(req.Username)
		if len(req.Username) < 3 {
			http.Error(w, "Username minimal 3 karakter", http.StatusBadRequest)
			return
		}
		if len(req.Password) < 6 {
			http.Error(w, "Password minimal 6 karakter", http.StatusBadRequest)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Gagal mengenkripsi password", http.StatusInternalServerError)
			return
		}

		res, err := db.Exec(
			"INSERT INTO users (username, password_hash) VALUES (?, ?);",
			req.Username, string(hash),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userID, _ := res.LastInsertId()
		duration := sessionDurationRemember
		rawToken, err := createSession(db, userID, duration, r)
		if err == nil {
			setSessionCookie(w, rawToken, duration)
		}

		writeJSON(w, http.StatusCreated, models.AuthStatusResponse{
			Initialized:   true,
			Authenticated: true,
			Username:      req.Username,
		})
	}
}

func handleAuthLogin(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		if !globalLoginLimiter.isAllowed(clientIP) {
			http.Error(w, "Terlalu banyak percobaan login yang gagal. Silakan coba lagi dalam 1 menit.", http.StatusTooManyRequests)
			return
		}

		var req models.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Payload tidak valid", http.StatusBadRequest)
			return
		}

		req.Username = strings.TrimSpace(req.Username)
		var user models.User
		err := db.QueryRow(
			"SELECT id, username, password_hash FROM users WHERE username = ?;",
			req.Username,
		).Scan(&user.ID, &user.Username, &user.PasswordHash)
		if err != nil {
			globalLoginLimiter.recordFailure(clientIP)
			http.Error(w, errInvalidCredentialsMsg, http.StatusUnauthorized)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			globalLoginLimiter.recordFailure(clientIP)
			http.Error(w, errInvalidCredentialsMsg, http.StatusUnauthorized)
			return
		}

		globalLoginLimiter.reset(clientIP)

		duration := sessionDurationStandard
		if req.RememberMe {
			duration = sessionDurationRemember
		}

		rawToken, err := createSession(db, user.ID, duration, r)
		if err != nil {
			http.Error(w, "Gagal membuat sesi", http.StatusInternalServerError)
			return
		}

		setSessionCookie(w, rawToken, duration)

		writeJSON(w, http.StatusOK, models.AuthStatusResponse{
			Initialized:   true,
			Authenticated: true,
			Username:      user.Username,
		})
	}
}

func handleAuthLogout(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie(cookieSessionName); err == nil && cookie.Value != "" {
			tokenHash := hashToken(cookie.Value)
			_, _ = db.Exec("DELETE FROM sessions WHERE token_hash = ?;", tokenHash)
		}

		clearSessionCookie(w)
		writeJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
	}
}

func authMiddleware(db *database.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cookieSessionName)
			if err != nil || cookie.Value == "" {
				http.Error(w, errUnauthorizedMsg, http.StatusUnauthorized)
				return
			}

			user, err := validateSessionToken(db, cookie.Value)
			if err != nil || user == nil {
				clearSessionCookie(w)
				http.Error(w, errUnauthorizedMsg, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
