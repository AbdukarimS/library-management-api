package services

import (
	"context"
	"library-management-api/internal/models"
	"library-management-api/internal/repository"
	"net/http"
	"strings"
)

const MinBookYear = 1450

type BookService struct {
	bookRepo *repository.BookRepository
	loanRepo *repository.LoanRepository
}

func NewBookService(bookRepo *repository.BookRepository, loanRepo *repository.LoanRepository) *BookService {
	return &BookService{bookRepo: bookRepo, loanRepo: loanRepo}
}

func validateBookInput(b models.Book) error {
	if strings.TrimSpace(b.Title) == "" || strings.TrimSpace(b.Author) == "" {
		return NewAppError(http.StatusBadRequest, "INVALID_BOOK_DATA", "Название и автор книги обязательны")
	}
	if b.Year < MinBookYear {
		return NewAppError(http.StatusBadRequest, "INVALID_YEAR", "Некорректный год издания")
	}
	if b.TotalCopies < 1 {
		return NewAppError(http.StatusBadRequest, "INVALID_COPIES", "Количество экземпляров должно быть больше 0")
	}
	return nil
}

func (s *BookService) CreateBook(ctx context.Context, input models.Book) (*models.Book, error) {
	if err := validateBookInput(input); err != nil {
		return nil, err
	}
	input.AvailableCopies = input.TotalCopies
	if err := s.bookRepo.Create(ctx, &input); err != nil {
		return nil, err
	}
	return &input, nil
}

func (s *BookService) GetAllBooks(ctx context.Context) ([]models.Book, error) {
	return s.bookRepo.GetAll(ctx)
}

func (s *BookService) GetBookByID(ctx context.Context, id int) (*models.Book, error) {
	book, err := s.bookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if book == nil {
		return nil, ErrBookNotFound
	}
	return book, nil
}

func (s *BookService) UpdateBook(ctx context.Context, id int, input models.Book) (*models.Book, error) {
	if err := validateBookInput(input); err != nil {
		return nil, err
	}
	existing, err := s.bookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrBookNotFound
	}

	borrowedCount := existing.TotalCopies - existing.AvailableCopies
	if input.TotalCopies < borrowedCount {
		return nil, NewAppError(http.StatusConflict, "INVALID_TOTAL_COPIES", "Новое общее количество экземпляров меньше числа выданных книг")
	}

	existing.Title = input.Title
	existing.Author = input.Author
	existing.Year = input.Year
	existing.Genre = input.Genre
	existing.TotalCopies = input.TotalCopies
	existing.AvailableCopies = input.TotalCopies - borrowedCount

	if err := s.bookRepo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *BookService) DeleteBook(ctx context.Context, id int) error {
	existing, err := s.bookRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrBookNotFound
	}

	hasActive, err := s.loanRepo.HasActiveLoansForBook(ctx, id)
	if err != nil {
		return err
	}
	if hasActive {
		return ErrHasActiveLoans
	}
	return s.bookRepo.Delete(ctx, id)
}

func (s *BookService) SearchBooks(ctx context.Context, title, author, genre string, year int) ([]models.Book, error) {
	return s.bookRepo.Search(ctx, title, author, genre, year)
}
