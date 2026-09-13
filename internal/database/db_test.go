package database

import (
	"path/filepath"
	"testing"
)

func TestConnect_Success(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_neraca.db")

	db, err := Connect(dbPath)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping db: %v", err)
	}

	// Verifikasi bahwa tabel-tabel utama terbuat oleh initSchema()
	tables := []string{"accounts", "categories", "transactions"}
	for _, tbl := range tables {
		var name string
		query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?;`
		err := db.QueryRow(query, tbl).Scan(&name)
		if err != nil {
			t.Errorf("expected table '%s' to exist in database, got error: %v", tbl, err)
		}
		if name != tbl {
			t.Errorf("expected table name '%s', got '%s'", tbl, name)
		}
	}
}

func TestDatabase_InsertAndQuery(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_crud.db")

	db, err := Connect(dbPath)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	// Uji insert akun
	res, err := db.Exec(`INSERT INTO accounts (name, type, balance, currency) VALUES (?, ?, ?, ?)`,
		"BCA Tahapan", "bank", 5000000.0, "IDR")
	if err != nil {
		t.Fatalf("failed to insert account: %v", err)
	}

	accID, err := res.LastInsertId()
	if err != nil || accID <= 0 {
		t.Fatalf("expected valid last insert id, got %d, err: %v", accID, err)
	}

	// Query kembali
	var name, accType string
	var balance float64
	err = db.QueryRow(`SELECT name, type, balance FROM accounts WHERE id = ?`, accID).
		Scan(&name, &accType, &balance)
	if err != nil {
		t.Fatalf("failed to query inserted account: %v", err)
	}

	if name != "BCA Tahapan" || accType != "bank" || balance != 5000000.0 {
		t.Errorf("account data mismatch: got (%s, %s, %f)", name, accType, balance)
	}
}
