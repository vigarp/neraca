package main

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"neraca/internal/database"
)

func setupTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "neraca-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := database.Connect(dbPath)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to connect to test db: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tempDir)
	}

	return db, cleanup
}

func TestResetToSetupWizard(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Insert dummy user and session
	hash, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
	res, err := db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", "admin", string(hash))
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}
	userID, _ := res.LastInsertId()
	_, err = db.Exec("INSERT INTO sessions (user_id, token_hash, expires_at) VALUES (?, ?, datetime('now', '+1 day'))", userID, "dummyhash")
	if err != nil {
		t.Fatalf("failed to insert test session: %v", err)
	}

	// Execute reset to setup wizard (empty password)
	err = executeAuthReset(db, "", "", "8088")
	if err != nil {
		t.Fatalf("executeAuthReset failed: %v", err)
	}

	var userCount, sessionCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount); err != nil {
		t.Fatalf("failed to count users: %v", err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&sessionCount); err != nil {
		t.Fatalf("failed to count sessions: %v", err)
	}

	if userCount != 0 {
		t.Errorf("expected 0 users, got %d", userCount)
	}
	if sessionCount != 0 {
		t.Errorf("expected 0 sessions, got %d", sessionCount)
	}
}

func TestResetDirectPassword_ExistingUser(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Insert dummy user and session
	oldHash, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
	res, err := db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", "myadmin", string(oldHash))
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}
	userID, _ := res.LastInsertId()
	_, _ = db.Exec("INSERT INTO sessions (user_id, token_hash, expires_at) VALUES (?, ?, datetime('now', '+1 day'))", userID, "dummyhash")

	// Update password directly
	newPass := "newSuperSecret123"
	err = executeAuthReset(db, newPass, "", "8088")
	if err != nil {
		t.Fatalf("executeAuthReset failed: %v", err)
	}

	// Verify session revoked
	var sessionCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&sessionCount)
	if sessionCount != 0 {
		t.Errorf("expected sessions to be revoked, got %d", sessionCount)
	}

	// Verify password updated
	var updatedHash string
	err = db.QueryRow("SELECT password_hash FROM users WHERE username = ?", "myadmin").Scan(&updatedHash)
	if err != nil {
		t.Fatalf("failed to query updated user: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(updatedHash), []byte(newPass)); err != nil {
		t.Errorf("password hash does not match new password: %v", err)
	}
}

func TestResetDirectPassword_NewUserWhenEmpty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// No users initially
	newPass := "brandNewPass123"
	err := executeAuthReset(db, newPass, "customAdmin", "8088")
	if err != nil {
		t.Fatalf("executeAuthReset failed: %v", err)
	}

	var username, hash string
	err = db.QueryRow("SELECT username, password_hash FROM users WHERE username = ?", "customAdmin").Scan(&username, &hash)
	if err != nil {
		t.Fatalf("user customAdmin not found: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(newPass)); err != nil {
		t.Errorf("password hash does not match: %v", err)
	}
}

func TestResetDirectPassword_ShortPasswordValidation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	err := executeAuthReset(db, "123", "", "8088")
	if err == nil {
		t.Fatal("expected error for short password, got nil")
	}
}
