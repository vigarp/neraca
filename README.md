# Neraca ⚖️

> **Ultra-lightweight Personal Financial Dashboard**  
> Didesain khusus untuk efisiensi maksimal pada VPS resource terbatas (RAM footprint < 30 MB).

---

## 🚀 Tech Stack

* **Backend**: Go (Golang) + [Chi Router v5](https://github.com/go-chi/chi)
* **Database**: SQLite (Pure Go driver via [`modernc.org/sqlite`](https://gitlab.com/cznic/sqlite), WAL mode)
* **Frontend**: [Vue 3](https://vuejs.org/) + [Tailwind CSS](https://tailwindcss.com/) + [Vite](https://vitejs.dev/) + [Lucide Icons](https://lucide.dev/)
* **Packaging**: Single Binary via Go `//go:embed` & Docker Multi-Stage Build

---

## 📁 Struktur Proyek

```text
neraca/
├── cmd/
│   └── server/
│       └── main.go           # Entry point backend Go
├── internal/
│   ├── api/                  # HTTP Router, Middleware & Handlers
│   ├── config/               # App configuration (PORT, DB_PATH, ENV)
│   └── database/             # SQLite connection & migrations (WAL mode)
├── web/                      # Frontend Vue 3 + Tailwind CSS
│   ├── src/                  # Source code Vue (components, views, assets)
│   ├── dist/                 # Hasil build statis yang di-embed oleh Go
│   ├── embed.go              # Embedding go:embed untuk web/dist
│   └── vite.config.js        # Konfigurasi Vite & proxy API
├── data/                     # Volume direktori SQLite database (neraca.db)
├── Dockerfile                # Multi-stage build (Node -> Go Embed -> Alpine)
├── docker-compose.yml        # Orchestration Docker
├── Makefile                  # Perintah praktis developer
└── README.md
```

---

## 🛠️ Panduan Pengembangan (Local Development)

### 1. Prasyarat
* Go 1.22+
* Node.js 20+ & npm

### 2. Menjalankan Mode Development
Buka 2 terminal:

**Terminal 1 (Backend Go):**
```bash
make dev-backend
# atau: go run ./cmd/server
# Berjalan di http://localhost:8088
```

**Terminal 2 (Frontend Vue dengan Hot Module Replacement):**
```bash
make dev-frontend
# atau: cd web && npm run dev
# Berjalan di http://localhost:5173 (Otomatis proxy /api ke :8080)
```

---

## 📦 Build & Production

### 1. Build Single Binary Mandiri
Frontend akan di-build dan di-embed langsung ke dalam 1 file binary Go:
```bash
make build
./bin/neraca
```
Akses `http://localhost:8080` di browser.

### 2. Menjalankan via Docker / VPS
```bash
# Build dan jalankan container
docker compose up -d --build

# Cek logs
docker compose logs -f

# Matikan container
docker compose down
```
File database SQLite otomatis tersimpan aman di folder `./data/neraca.db` pada host VPS.
