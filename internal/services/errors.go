package services

import "net/http"

type AppError struct {
	Status  int
	Code    string
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

var (
	ErrBookNotFound     = NewAppError(http.StatusNotFound, "BOOK_NOT_FOUND", "Книга не найдена")
	ErrReaderNotFound   = NewAppError(http.StatusNotFound, "READER_NOT_FOUND", "Читатель не найден")
	ErrLoanNotFound     = NewAppError(http.StatusNotFound, "LOAN_NOT_FOUND", "Операция выдачи не найдена")
	ErrBookUnavailable  = NewAppError(http.StatusConflict, "BOOK_NOT_AVAILABLE", "Нет доступных экземпляров книги")
	ErrReaderInactive   = NewAppError(http.StatusConflict, "READER_INACTIVE", "Читатель неактивен")
	ErrLoanLimitReached = NewAppError(http.StatusConflict, "MAX_LOANS_EXCEEDED", "Читатель не может иметь более 5 активных выдач")
	ErrAlreadyReturned  = NewAppError(http.StatusConflict, "LOAN_ALREADY_RETURNED", "Книга уже была возвращена")
	ErrHasActiveLoans   = NewAppError(http.StatusConflict, "ACTIVE_LOANS_EXIST", "Невозможно удалить запись при наличии активных выдач")
	ErrDuplicateEmail   = NewAppError(http.StatusConflict, "EMAIL_ALREADY_EXISTS", "Читатель с таким Email уже зарегистрирован")
)
