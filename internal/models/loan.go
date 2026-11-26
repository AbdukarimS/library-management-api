package models

import "time"

type Loan struct {
	ID         int        `json:"id"`
	BookID     int        `json:"bookId"`
	ReaderID   int        `json:"readerId"`
	BorrowedAt time.Time  `json:"borrowedAt"`
	DueDate    time.Time  `json:"dueDate"`
	ReturnedAt *time.Time `json:"returnedAt"`
	IsReturned bool       `json:"isReturned"`
}

type TopBookStat struct {
	BookID     int    `json:"bookId"`
	Title      string `json:"title"`
	Author     string `json:"author"`
	LoansCount int    `json:"loansCount"`
}

type Statistics struct {
	TotalBooks          int           `json:"totalBooks"`
	TotalCopies         int           `json:"totalCopies"`
	AvailableCopies     int           `json:"availableCopies"`
	AvailabilityPercent float64       `json:"availabilityPercent"`
	TotalReaders        int           `json:"totalReaders"`
	ActiveLoans         int           `json:"activeLoans"`
	OverdueLoans        int           `json:"overdueLoans"`
	TopBooks            []TopBookStat `json:"topBooks"`
	LoansByMonth        [12]int       `json:"loansByMonth"`
}
