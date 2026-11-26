package services

import (
	"context"
	"library-management-api/internal/models"
	"library-management-api/internal/repository"
	"math"
	"net/http"
	"time"
)

const (
	MaxActiveLoansPerReader = 5
	DefaultLoanDays         = 14
)

type LoanService struct {
	loanRepo   *repository.LoanRepository
	bookRepo   *repository.BookRepository
	readerRepo *repository.ReaderRepository
}

func NewLoanService(loanRepo *repository.LoanRepository, bookRepo *repository.BookRepository, readerRepo *repository.ReaderRepository) *LoanService {
	return &LoanService{loanRepo: loanRepo, bookRepo: bookRepo, readerRepo: readerRepo}
}

func (s *LoanService) BorrowBook(ctx context.Context, bookID, readerID, days int) (*models.Loan, error) {
	if bookID <= 0 || readerID <= 0 {
		return nil, NewAppError(http.StatusBadRequest, "INVALID_LOAN_INPUT", "Укажите корректные bookId и readerId")
	}
	if days <= 0 {
		days = DefaultLoanDays
	}
	return s.loanRepo.BorrowBookTx(
		ctx, bookID, readerID, days, MaxActiveLoansPerReader,
		ErrBookNotFound, ErrReaderNotFound, ErrReaderInactive, ErrBookUnavailable, ErrLoanLimitReached,
	)
}

func (s *LoanService) ReturnBook(ctx context.Context, loanID int) (*models.Loan, error) {
	return s.loanRepo.ReturnBookTx(ctx, loanID, ErrLoanNotFound, ErrAlreadyReturned)
}

func (s *LoanService) GetAllLoans(ctx context.Context) ([]models.Loan, error) {
	return s.loanRepo.GetAll(ctx, false)
}

func (s *LoanService) GetActiveLoans(ctx context.Context) ([]models.Loan, error) {
	return s.loanRepo.GetAll(ctx, true)
}

func (s *LoanService) GetLoanByID(ctx context.Context, id int) (*models.Loan, error) {
	loan, err := s.loanRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if loan == nil {
		return nil, ErrLoanNotFound
	}
	return loan, nil
}

func (s *LoanService) GetStatistics(ctx context.Context) (*models.Statistics, error) {
	books, err := s.bookRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	readers, err := s.readerRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	loans, err := s.loanRepo.GetAll(ctx, false)
	if err != nil {
		return nil, err
	}
	topBooks, err := s.loanRepo.GetTopBooks(ctx, 3)
	if err != nil {
		return nil, err
	}

	stats := &models.Statistics{
		TotalBooks:   len(books),
		TotalReaders: len(readers),
		TopBooks:     topBooks,
	}

	for _, b := range books {
		stats.TotalCopies += b.TotalCopies
		stats.AvailableCopies += b.AvailableCopies
	}
	if stats.TotalCopies > 0 {
		pct := (float64(stats.AvailableCopies) / float64(stats.TotalCopies)) * 100
		stats.AvailabilityPercent = math.Round(pct*100) / 100
	}

	now := time.Now().UTC()
	var monthly [12]int
	for _, l := range loans {
		m := int(l.BorrowedAt.Month()) - 1
		if m >= 0 && m < 12 {
			monthly[m]++
		}
		if !l.IsReturned {
			stats.ActiveLoans++
			if now.After(l.DueDate) {
				stats.OverdueLoans++
			}
		}
	}
	stats.LoansByMonth = monthly
	return stats, nil
}

func (s *LoanService) CheckDBHealth(ctx context.Context) error {
	return s.loanRepo.DB().PingContext(ctx)
}
