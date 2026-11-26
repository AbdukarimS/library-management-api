package repository

import (
	"context"
	"database/sql"
	"errors"
	"library-management-api/internal/models"
	"strings"
)

type ReaderRepository struct {
	db *sql.DB
}

func NewReaderRepository(db *sql.DB) *ReaderRepository {
	return &ReaderRepository{db: db}
}

func (r *ReaderRepository) Create(ctx context.Context, reader *models.Reader) error {
	query := `
		INSERT INTO readers (first_name, last_name, email, age, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	return r.db.QueryRowContext(ctx, query, reader.FirstName, reader.LastName, reader.Email, reader.Age, reader.IsActive).Scan(&reader.ID)
}

func (r *ReaderRepository) GetAll(ctx context.Context) ([]models.Reader, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, first_name, last_name, email, age, is_active FROM readers ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	readers := []models.Reader{}
	for rows.Next() {
		var rd models.Reader
		if err := rows.Scan(&rd.ID, &rd.FirstName, &rd.LastName, &rd.Email, &rd.Age, &rd.IsActive); err != nil {
			return nil, err
		}
		readers = append(readers, rd)
	}
	return readers, rows.Err()
}

func (r *ReaderRepository) GetByID(ctx context.Context, id int) (*models.Reader, error) {
	var rd models.Reader
	query := `SELECT id, first_name, last_name, email, age, is_active FROM readers WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&rd.ID, &rd.FirstName, &rd.LastName, &rd.Email, &rd.Age, &rd.IsActive)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rd, nil
}

func (r *ReaderRepository) SearchByName(ctx context.Context, name string) ([]models.Reader, error) {
	pattern := "%" + strings.TrimSpace(name) + "%"
	query := `
		SELECT id, first_name, last_name, email, age, is_active
		FROM readers
		WHERE LOWER(first_name || ' ' || last_name) LIKE LOWER($1)
		ORDER BY id`
	rows, err := r.db.QueryContext(ctx, query, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	readers := []models.Reader{}
	for rows.Next() {
		var rd models.Reader
		if err := rows.Scan(&rd.ID, &rd.FirstName, &rd.LastName, &rd.Email, &rd.Age, &rd.IsActive); err != nil {
			return nil, err
		}
		readers = append(readers, rd)
	}
	return readers, rows.Err()
}

func (r *ReaderRepository) ExistsByEmail(ctx context.Context, email string, excludeID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM readers WHERE LOWER(email) = LOWER($1) AND id <> $2)`
	err := r.db.QueryRowContext(ctx, query, strings.TrimSpace(email), excludeID).Scan(&exists)
	return exists, err
}

func (r *ReaderRepository) Update(ctx context.Context, rd *models.Reader) error {
	query := `
		UPDATE readers
		SET first_name = $1, last_name = $2, email = $3, age = $4, is_active = $5
		WHERE id = $6`
	_, err := r.db.ExecContext(ctx, query, rd.FirstName, rd.LastName, rd.Email, rd.Age, rd.IsActive, rd.ID)
	return err
}

func (r *ReaderRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM readers WHERE id = $1`, id)
	return err
}
