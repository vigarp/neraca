package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"neraca/internal/database"
	"neraca/internal/models"
)

func executeAuthReset(db *database.DB, newPassword, targetUser, port string) error {
	newPassword = strings.TrimSpace(newPassword)
	if newPassword != "" {
		return resetDirectPassword(db, newPassword, targetUser, port)
	}
	return resetToSetupWizard(db, port)
}

const separatorLine = "=================================================="

func resetToSetupWizard(db *database.DB, port string) error {
	_, err := db.Exec("DELETE FROM sessions; DELETE FROM users;")
	if err != nil {
		return fmt.Errorf("gagal mereset tabel autentikasi: %w", err)
	}

	fmt.Println(separatorLine)
	fmt.Println("  🔐 Reset Otentikasi Berhasil!")
	fmt.Println(separatorLine)
	fmt.Println("✔ Seluruh sesi aktif dan akun login telah dihapus.")
	fmt.Println("✔ Data keuangan (rekening, transaksi, kategori, aset) tetap aman 100%.")
	fmt.Printf("👉 Buka Neraca di browser (http://localhost:%s) untuk membuat akun/password baru via Setup Wizard.\n", port)
	fmt.Println(separatorLine)
	return nil
}

func resetDirectPassword(db *database.DB, newPassword, targetUser, port string) error {
	if len(newPassword) < 6 {
		return errors.New("password minimal 6 karakter")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("gagal mengenkripsi password: %w", err)
	}

	var user models.User
	err = db.QueryRow("SELECT id, username FROM users ORDER BY id ASC LIMIT 1;").Scan(&user.ID, &user.Username)
	if errors.Is(err, sql.ErrNoRows) {
		uName := strings.TrimSpace(targetUser)
		if uName == "" {
			uName = "admin"
		}
		_, err = db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?);", uName, string(hash))
		if err != nil {
			return fmt.Errorf("gagal membuat user: %w", err)
		}
		fmt.Printf("✔ User admin baru '%s' berhasil dibuat dengan password baru.\n", uName)
	} else if err != nil {
		return fmt.Errorf("gagal query user: %w", err)
	} else {
		uName := user.Username
		if targetUser != "" {
			uName = strings.TrimSpace(targetUser)
		}
		_, err = db.Exec("UPDATE users SET username = ?, password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;", uName, string(hash), user.ID)
		if err != nil {
			return fmt.Errorf("gagal memperbarui password: %w", err)
		}
		fmt.Printf("✔ Password untuk user '%s' berhasil diperbarui!\n", uName)
	}

	_, _ = db.Exec("DELETE FROM sessions;")
	fmt.Println("✔ Semua sesi aktif lama telah dicabut demi keamanan.")
	fmt.Printf("👉 Silakan login di: http://localhost:%s\n", port)
	return nil
}

