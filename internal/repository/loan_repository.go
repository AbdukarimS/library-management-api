package repository

import (
	"context"
	"database/sql"
	"errors"
	"library-management-api/internal/models"
	"time"
)

type LoanRepository struct {
	db *sql.DB
}

func NewLoanRepository(db *sql.DB) *LoanRepository {
	return &LoanRepository{db: db}
}

func (r *LoanRepository) DB() *sql.DB {
	return r.db
}

func (r *LoanRepository) HasActiveLoansForBook(ctx context.Context, bookID int) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM loans WHERE book_id = $1 AND is_returned = FALSE)`, bookID).Scan(&exists)
	return exists, err
}

func (r *LoanRepository) CountActiveLoansForReader(ctx context.Context, readerID int) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM loans WHERE reader_id = $1 AND is_returned = FALSE`, readerID).Scan(&count)
	return count, err
}

// BorrowBookTx выполняет все проверки и выдачу книги в единой транзакции
func (r *LoanRepository) BorrowBookTx(ctx context.Context, bookID, readerID, days, maxLoans int, errBookNotFound, errReaderNotFound, errReaderInactive, errBookUnavail, errLimit error) (*models.Loan, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// 1. Блокируем строку книги FOR UPDATE для защиты от гонки при одновременных запросах
	var availableCopies int
	err = tx.QueryRowContext(ctx, `SELECT available_copies FROM books WHERE id = $1 FOR UPDATE`, bookID).Scan(&availableCopies)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errBookNotFound
	}
	if err != nil {
		return nil, err
	}
	if availableCopies <= 0 {
		return nil, errBookUnavail
	}

	// 2. Блокируем строку читателя FOR UPDATE и проверяем активность
	var isActive bool
	err = tx.QueryRowContext(ctx, `SELECT is_active FROM readers WHERE id = $1 FOR UPDATE`, readerID).Scan(&isActive)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errReaderNotFound
	}
	if err != nil {
		return nil, err
	}
	if !isActive {
		return nil, errReaderInactive
	}

	// 3. Проверяем лимит активных выдач (<= 5)
	var activeCount int
	err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM loans WHERE reader_id = $1 AND is_returned = FALSE`, readerID).Scan(&activeCount)
	if err != nil {
		return nil, err
	}
	if activeCount >= maxLoans {
		return nil, errLimit
	}

	// 4. Уменьшаем available_copies на 1
	if _, err := tx.ExecContext(ctx, `UPDATE books SET available_copies = available_copies - 1 WHERE id = $1`, bookID); err != nil {
		return nil, err
	}

	// 5. Создаём запись о выдаче
	now := time.Now().UTC()
	loan := &models.Loan{
		BookID:     bookID,
		ReaderID:   readerID,
		BorrowedAt: now,
		DueDate:    now.AddDate(0, 0, days),
		IsReturned: false,
	}
	err = tx.QueryRowContext(ctx,
		`INSERT INTO loans (book_id, reader_id, borrowed_at, due_date, is_returned) VALUES ($1, $2, $3, $4, FALSE) RETURNING id`,
		loan.BookID, loan.ReaderID, loan.BorrowedAt, loan.DueDate,
	).Scan(&loan.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return loan, nil
}

// ReturnBookTx выполняет возврат книги и увеличение available_copies в единой транзакции
func (r *LoanRepository) ReturnBookTx(ctx context.Context, loanID int, errLoanNotFound, errAlreadyReturned error) (*models.Loan, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var loan models.Loan
	err = tx.QueryRowContext(ctx,
		`SELECT id, book_id, reader_id, borrowed_at, due_date, returned_at, is_returned FROM loans WHERE id = $1 FOR UPDATE`,
		loanID,
	).Scan(&loan.ID, &loan.BookID, &loan.ReaderID, &loan.BorrowedAt, &loan.DueDate, &loan.ReturnedAt, &loan.IsReturned)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errLoanNotFound
	}
	if err != nil {
		return nil, err
	}
	if loan.IsReturned {
		return nil, errAlreadyReturned
	}

	now := time.Now().UTC()
	loan.IsReturned = true
	loan.ReturnedAt = &now

	if _, err := tx.ExecContext(ctx, `UPDATE loans SET is_returned = TRUE, returned_at = $1 WHERE id = $2`, now, loanID); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE books SET available_copies = available_copies + 1 WHERE id = $1`, loan.BookID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &loan, nil
}

func (r *LoanRepository) GetAll(ctx context.Context, onlyActive bool) ([]models.Loan, error) {
	query := `SELECT id, book_id, reader_id, borrowed_at, due_date, returned_at, is_returned FROM loans`
	if onlyActive {
		query += ` WHERE is_returned = FALSE`
	}
	query += ` ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	loans := []models.Loan{}
	for rows.Next() {
		var l models.Loan
		if err := rows.Scan(&l.ID, &l.BookID, &l.ReaderID, &l.BorrowedAt, &l.DueDate, &l.ReturnedAt, &l.IsReturned); err != nil {
			return nil, err
		}
		loans = append(loans, l)
	}
	return loans, rows.Err()
}

func (r *LoanRepository) GetByID(ctx context.Context, id int) (*models.Loan, error) {
	var l models.Loan
	query := `SELECT id, book_id, reader_id, borrowed_at, due_date, returned_at, is_returned FROM loans WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&l.ID, &l.BookID, &l.ReaderID, &l.BorrowedAt, &l.DueDate, &l.ReturnedAt, &l.IsReturned)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *LoanRepository) GetByReaderID(ctx context.Context, readerID int) ([]models.Loan, error) {
	query := `SELECT id, book_id, reader_id, borrowed_at, due_date, returned_at, is_returned FROM loans WHERE reader_id = $1 ORDER BY id`
	rows, err := r.db.QueryContext(ctx, query, readerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	loans := []models.Loan{}
	for rows.Next() {
		var l models.Loan
		if err := rows.Scan(&l.ID, &l.BookID, &l.ReaderID, &l.BorrowedAt, &l.DueDate, &l.ReturnedAt, &l.IsReturned); err != nil {
			return nil, err
		}
		loans = append(loans, l)
	}
	return loans, rows.Err()
}

func (r *LoanRepository) GetTopBooks(ctx context.Context, limit int) ([]models.TopBookStat, error) {
	query := `
		SELECT b.id, b.title, b.author, COUNT(l.id) AS loans_count
		FROM loans l
		JOIN books b ON b.id = l.book_id
		GROUP BY b.id, b.title, b.author
		ORDER BY loans_count DESC, b.id ASC
		LIMIT $1`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	top := []models.TopBookStat{}
	for rows.Next() {
		var t models.TopBookStat
		if err := rows.Scan(&t.BookID, &t.Title, &t.Author, &t.LoansCount); err != nil {
			return nil, err
		}
		top = append(top, t)
	}
	return top, rows.Err()
}
