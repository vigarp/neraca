package api

import (
	"io/fs"
	"mime"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"neraca/internal/config"
	"neraca/internal/database"
)

func init() {
	_ = mime.AddExtensionType(".webmanifest", "application/manifest+json")
}

const headerCacheControl = "Cache-Control"

func NewRouter(cfg *config.Config, db *database.DB, staticFS fs.FS) http.Handler {
	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// CORS config (terutama saat development ketika Vite jalan di port 5173)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Security Headers (Lighthouse Best Practices)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			next.ServeHTTP(w, r)
		})
	})

	// API routes
	r.Route("/api", func(api chi.Router) {
		api.Get("/health", handleHealth(db))

		// Public Auth Endpoints
		api.Get("/auth/status", handleAuthStatus(db))
		api.Post("/auth/setup", handleAuthSetup(db))
		api.Post("/auth/login", handleAuthLogin(db))
		api.Post("/auth/logout", handleAuthLogout(db))

		// Protected Financial Routes (Membutuhkan Cookie Sesi Aktif)
		api.Group(func(protected chi.Router) {
			protected.Use(authMiddleware(db))

			// Accounts & Wealth Management
			protected.Get("/accounts", handleGetAccounts(db))
			protected.Post("/accounts", handleCreateAccount(db))
			protected.Put("/accounts/{id}", handleUpdateAccount(db))
			protected.Delete("/accounts/{id}", handleDeleteAccount(db))
			protected.Post("/accounts/{id}/revalue", handleRevalueAccount(db))

			// Settings & Data Management
			protected.Get("/settings", handleGetSettings(db))
			protected.Put("/settings", handleUpdateSettings(db))
			protected.Post("/settings/reset-data", handleResetFinancialData(db))
			protected.Get("/settings/backup", handleDownloadBackup(db))

			// Categories (50-30-20 Framework)
			protected.Get("/categories", handleGetCategories(db))
			protected.Post("/categories", handleCreateCategory(db))
			protected.Put("/categories/{id}", handleUpdateCategory(db))
			protected.Delete("/categories/{id}", handleDeleteCategory(db))

			// Transactions (Daily expense/income/transfer)
			protected.Get("/transactions", handleGetTransactions(db))
			protected.Post("/transactions", handleCreateTransaction(db))
			protected.Post("/transactions/transfer", handleCreateTransfer(db))
			protected.Delete("/transactions/{id}", handleDeleteTransaction(db))

			// Analytics (Payday countdown, prospect daily limit, burn rate)
			protected.Get("/analytics/burn-rate", handleGetBurnRateAnalytics(db))
		})
	})

	// Static SPA Handler (jika ada embedded static files)
	if staticFS != nil {
		fileServer := http.FileServer(http.FS(staticFS))

		spaHandler := func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/")

			// Caching Headers (Lighthouse Performance & PWA)
			if strings.HasPrefix(path, "assets/") {
				// Hashed Vite static assets aman di-cache 1 tahun
				w.Header().Set(headerCacheControl, "public, max-age=31536000, immutable")
			} else if path == "sw.js" {
				// Service worker tidak boleh di-cache oleh browser agar auto-update bekerja
				w.Header().Set(headerCacheControl, "no-cache, no-store, must-revalidate")
			} else if path == "manifest.webmanifest" || path == "icon.svg" {
				w.Header().Set(headerCacheControl, "public, max-age=86400")
			}

			// Cek apakah file ada di filesystem statis
			if path != "" {
				if f, err := staticFS.Open(path); err == nil {
					_ = f.Close()
					fileServer.ServeHTTP(w, r)
					return
				}
			}

			// Fallback ke index.html untuk Single Page Application (SPA) client-side routing
			indexFile, err := staticFS.Open("index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			_ = indexFile.Close()

			w.Header().Set(headerCacheControl, "no-cache")
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
		}

		r.Get("/*", spaHandler)
		r.Head("/*", spaHandler)
	}

	return r
}
