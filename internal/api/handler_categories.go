package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"neraca/internal/database"
	"neraca/internal/models"
)

const (
	errInvalidCategoryID = "Invalid category ID"
	errCategoryNotFound  = "Category not found"
)

func handleGetCategories(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := `
		SELECT id, name, type, pillar, COALESCE(icon, ''), COALESCE(color, ''), created_at
		FROM categories
		ORDER BY pillar ASC, name ASC;
		`
		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		res := models.CategoriesSummaryResponse{
			NeedsCategories:   []models.Category{},
			WantsCategories:   []models.Category{},
			SavingsCategories: []models.Category{},
			IncomeCategories:  []models.Category{},
			AllCategories:     []models.Category{},
		}

		for rows.Next() {
			var c models.Category
			if err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.Pillar, &c.Icon, &c.Color, &c.CreatedAt); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			res.AllCategories = append(res.AllCategories, c)

			switch c.Pillar {
			case "needs":
				res.NeedsCategories = append(res.NeedsCategories, c)
			case "wants":
				res.WantsCategories = append(res.WantsCategories, c)
			case "savings":
				res.SavingsCategories = append(res.SavingsCategories, c)
			case "income":
				res.IncomeCategories = append(res.IncomeCategories, c)
			default:
				res.NeedsCategories = append(res.NeedsCategories, c)
			}
		}

		writeJSON(w, http.StatusOK, res)
	}
}

func validateCategoryRequest(name, cType, pillar string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("Nama kategori wajib diisi")
	}
	if cType != "income" && cType != "expense" {
		return errors.New("Tipe kategori harus 'income' atau 'expense'")
	}
	if pillar != "needs" && pillar != "wants" && pillar != "savings" && pillar != "income" {
		return errors.New("Pilar harus 'needs', 'wants', 'savings', atau 'income'")
	}
	return nil
}

func handleCreateCategory(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateCategoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, errInvalidPayload, http.StatusBadRequest)
			return
		}

		req.Name = strings.TrimSpace(req.Name)
		if req.Type == "" {
			req.Type = "expense"
		}
		if req.Pillar == "" {
			if req.Type == "income" {
				req.Pillar = "income"
			} else {
				req.Pillar = "needs"
			}
		}

		if err := validateCategoryRequest(req.Name, req.Type, req.Pillar); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		query := `
		INSERT INTO categories (name, type, pillar, icon, color)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, created_at;
		`
		var c models.Category
		c.Name = req.Name
		c.Type = req.Type
		c.Pillar = req.Pillar
		c.Icon = req.Icon
		c.Color = req.Color

		err := db.QueryRow(query, c.Name, c.Type, c.Pillar, c.Icon, c.Color).Scan(&c.ID, &c.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusCreated, c)
	}
}

func handleUpdateCategory(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, errInvalidCategoryID, http.StatusBadRequest)
			return
		}

		var req models.UpdateCategoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, errInvalidPayload, http.StatusBadRequest)
			return
		}

		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" {
			http.Error(w, "Nama kategori wajib diisi", http.StatusBadRequest)
			return
		}

		query := `
		UPDATE categories
		SET name = ?, pillar = ?, icon = ?, color = ?
		WHERE id = ?
		RETURNING id, name, type, pillar, COALESCE(icon, ''), COALESCE(color, ''), created_at;
		`
		var c models.Category
		err = db.QueryRow(query, req.Name, req.Pillar, req.Icon, req.Color, id).
			Scan(&c.ID, &c.Name, &c.Type, &c.Pillar, &c.Icon, &c.Color, &c.CreatedAt)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, errCategoryNotFound, http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, c)
	}
}

func handleDeleteCategory(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, errInvalidCategoryID, http.StatusBadRequest)
			return
		}

		res, err := db.Exec(`DELETE FROM categories WHERE id = ?`, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rowsAff, _ := res.RowsAffected()
		if rowsAff == 0 {
			http.Error(w, errCategoryNotFound, http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
