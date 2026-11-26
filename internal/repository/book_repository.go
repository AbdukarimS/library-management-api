package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"library-management-api/internal/models"
	"strings"
)

type BookRepository struct {
	db *sql.DB
}

func NewBookRepository(db *sql.DB) *BookRepository {
	return &BookRepository{db: db}
}

func (r *BookRepository) Create(ctx context.Context, b *models.Book) error {
	query := `
		INSERT INTO books (title, author, year, genre, total_copies, available_copies)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`
	return r.db.QueryRowContext(ctx, query, b.Title, b.Author, b.Year, b.Genre, b.TotalCopies, b.AvailableCopies).Scan(&b.ID)
}

func (r *BookRepository) GetAll(ctx context.Context) ([]models.Book, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, title, author, year, genre, total_copies, available_copies FROM books ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := []models.Book{}
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Year, &b.Genre, &b.TotalCopies, &b.AvailableCopies); err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, rows.Err()
}

func (r *BookRepository) GetByID(ctx context.Context, id int) (*models.Book, error) {
	var b models.Book
	query := `SELECT id, title, author, year, genre, total_copies, available_copies FROM books WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&b.ID, &b.Title, &b.Author, &b.Year, &b.Genre, &b.TotalCopies, &b.AvailableCopies)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BookRepository) Update(ctx context.Context, b *models.Book) error {
	query := `
		UPDATE books
		SET title = $1, author = $2, year = $3, genre = $4, total_copies = $5, available_copies = $6
		WHERE id = $7`
	_, err := r.db.ExecContext(ctx, query, b.Title, b.Author, b.Year, b.Genre, b.TotalCopies, b.AvailableCopies, b.ID)
	return err
}

func (r *BookRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM books WHERE id = $1`, id)
	return err
}

func (r *BookRepository) Search(ctx context.Context, title, author, genre string, year int) ([]models.Book, error) {
	query := `SELECT id, title, author, year, genre, total_copies, available_copies FROM books WHERE 1=1`
	var args []any
	idx := 1

	if strings.TrimSpace(title) != "" {
		query += fmt.Sprintf(" AND LOWER(title) LIKE LOWER($%d)", idx)
		args = append(args, "%"+strings.TrimSpace(title)+"%")
		idx++
	}
	if strings.TrimSpace(author) != "" {
		query += fmt.Sprintf(" AND LOWER(author) LIKE LOWER($%d)", idx)
		args = append(args, "%"+strings.TrimSpace(author)+"%")
		idx++
	}
	if strings.TrimSpace(genre) != "" {
		query += fmt.Sprintf(" AND LOWER(genre) LIKE LOWER($%d)", idx)
		args = append(args, "%"+strings.TrimSpace(genre)+"%")
		idx++
	}
	if year > 0 {
		query += fmt.Sprintf(" AND year = $%d", idx)
		args = append(args, year)
	}
	query += " ORDER BY id"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := []models.Book{}
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Year, &b.Genre, &b.TotalCopies, &b.AvailableCopies); err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, rows.Err()
}
