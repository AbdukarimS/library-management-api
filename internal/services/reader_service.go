package services

import (
	"context"
	"library-management-api/internal/models"
	"library-management-api/internal/repository"
	"net/http"
	"strings"
)

const MinReaderAge = 6

type ReaderService struct {
	readerRepo *repository.ReaderRepository
	loanRepo   *repository.LoanRepository
}

func NewReaderService(readerRepo *repository.ReaderRepository, loanRepo *repository.LoanRepository) *ReaderService {
	return &ReaderService{readerRepo: readerRepo, loanRepo: loanRepo}
}

func validateReaderInput(r models.Reader) error {
	if strings.TrimSpace(r.FirstName) == "" || strings.TrimSpace(r.LastName) == "" {
		return NewAppError(http.StatusBadRequest, "INVALID_READER_NAME", "Имя и фамилия читателя обязательны")
	}
	if !strings.Contains(r.Email, "@") || strings.TrimSpace(r.Email) == "" {
		return NewAppError(http.StatusBadRequest, "INVALID_EMAIL", "Некорректный адрес электронной почты")
	}
	if r.Age < MinReaderAge {
		return NewAppError(http.StatusBadRequest, "INVALID_AGE", "Возраст читателя должен быть не менее 6 лет")
	}
	return nil
}

func (s *ReaderService) CreateReader(ctx context.Context, input models.Reader) (*models.Reader, error) {
	if err := validateReaderInput(input); err != nil {
		return nil, err
	}
	taken, err := s.readerRepo.ExistsByEmail(ctx, input.Email, 0)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, ErrDuplicateEmail
	}
	if err := s.readerRepo.Create(ctx, &input); err != nil {
		return nil, err
	}
	return &input, nil
}

func (s *ReaderService) GetAllReaders(ctx context.Context) ([]models.Reader, error) {
	return s.readerRepo.GetAll(ctx)
}

func (s *ReaderService) GetReaderByID(ctx context.Context, id int) (*models.Reader, error) {
	rd, err := s.readerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rd == nil {
		return nil, ErrReaderNotFound
	}
	return rd, nil
}

func (s *ReaderService) SearchReadersByName(ctx context.Context, name string) ([]models.Reader, error) {
	return s.readerRepo.SearchByName(ctx, name)
}

func (s *ReaderService) UpdateReader(ctx context.Context, id int, input models.Reader) (*models.Reader, error) {
	if err := validateReaderInput(input); err != nil {
		return nil, err
	}
	existing, err := s.readerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrReaderNotFound
	}
	taken, err := s.readerRepo.ExistsByEmail(ctx, input.Email, id)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, ErrDuplicateEmail
	}

	existing.FirstName = input.FirstName
	existing.LastName = input.LastName
	existing.Email = input.Email
	existing.Age = input.Age
	existing.IsActive = input.IsActive

	if err := s.readerRepo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *ReaderService) DeleteReader(ctx context.Context, id int) error {
	existing, err := s.readerRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrReaderNotFound
	}
	activeCount, err := s.loanRepo.CountActiveLoansForReader(ctx, id)
	if err != nil {
		return err
	}
	if activeCount > 0 {
		return ErrHasActiveLoans
	}
	return s.readerRepo.Delete(ctx, id)
}

func (s *ReaderService) GetReaderLoans(ctx context.Context, readerID int) ([]models.Loan, error) {
	existing, err := s.readerRepo.GetByID(ctx, readerID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrReaderNotFound
	}
	return s.loanRepo.GetByReaderID(ctx, readerID)
}
