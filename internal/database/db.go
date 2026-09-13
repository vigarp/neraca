package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func Connect(dbPath string) (*DB, error) {
	// Pastikan direktori tempat file database SQLite berada sudah ada
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// SQLite connection string dengan WAL mode dan busy timeout
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=synchronous(NORMAL)", dbPath)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Batasi koneksi untuk SQLite agar aman dari file locking
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	log.Printf("Connected to SQLite database at %s (WAL mode enabled)", dbPath)

	database := &DB{DB: db}
	if err := database.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to init database schema: %w", err)
	}

	return database, nil
}

// Inisialisasi skema (tabel akun, aset pasif, transaksi, kategori, dan settings)
func (db *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS accounts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		account_group TEXT NOT NULL DEFAULT 'operational', -- 'operational', 'passive_asset', 'emergency'
		type TEXT NOT NULL, -- bank, ewallet, cash, mutual_fund, gold, stocks, crypto, other
		balance REAL NOT NULL DEFAULT 0,
		currency TEXT NOT NULL DEFAULT 'IDR',
		institution TEXT DEFAULT '',
		is_active INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS asset_valuations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
		previous_balance REAL NOT NULL,
		new_balance REAL NOT NULL,
		difference REAL NOT NULL,
		notes TEXT,
		recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);

	INSERT OR IGNORE INTO settings (key, value) VALUES ('payday_date', '25');
	INSERT OR IGNORE INTO settings (key, value) VALUES ('monthly_income_budget', '6000000');

	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		type TEXT NOT NULL, -- income, expense
		pillar TEXT NOT NULL DEFAULT 'needs', -- needs (50%), wants (30%), savings (20%)
		icon TEXT,
		color TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
		category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
		amount REAL NOT NULL,
		type TEXT NOT NULL, -- income, expense, transfer
		description TEXT,
		transaction_date DATE NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	// Migrasi kolom jika tabel lama sudah ada
	_ = db.addColumnIfNotExists("accounts", "account_group", "TEXT NOT NULL DEFAULT 'operational'")
	_ = db.addColumnIfNotExists("accounts", "institution", "TEXT DEFAULT ''")
	_ = db.addColumnIfNotExists("accounts", "is_active", "INTEGER NOT NULL DEFAULT 1")
	_ = db.addColumnIfNotExists("categories", "pillar", "TEXT NOT NULL DEFAULT 'needs'")

	return nil
}

func (db *DB) addColumnIfNotExists(table, column, colDef string) error {
	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", table, column, colDef)
	_, err := db.Exec(query)
	return err
}
