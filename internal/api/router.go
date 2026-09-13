package api

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"neraca/internal/config"
	"neraca/internal/database"
)

func NewRouter(cfg *config.Config, db *database.DB, staticFS fs.FS) http.Handler {
	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS config (terutama saat development ketika Vite jalan di port 5173)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API routes
	r.Route("/api", func(api chi.Router) {
		api.Get("/health", handleHealth(db))
	})

	// Static SPA Handler (jika ada embedded static files)
	if staticFS != nil {
		fileServer := http.FileServer(http.FS(staticFS))

		spaHandler := func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/")

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

			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
		}

		r.Get("/*", spaHandler)
		r.Head("/*", spaHandler)
	}

	return r
}
