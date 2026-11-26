package handlers

import (
	"encoding/json"
	"library-management-api/internal/response"
	"library-management-api/internal/services"
	"net/http"
)

type LoanHandler struct {
	service *services.LoanService
}

func NewLoanHandler(service *services.LoanService) *LoanHandler {
	return &LoanHandler{service: service}
}

type borrowRequest struct {
	BookID   int `json:"bookId"`
	ReaderID int `json:"readerId"`
	Days     int `json:"days"`
}

func (h *LoanHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	loans, err := h.service.GetAllLoans(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, loans)
}

func (h *LoanHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	loan, err := h.service.GetLoanByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, loan)
}

func (h *LoanHandler) Borrow(w http.ResponseWriter, r *http.Request) {
	var req borrowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Некорректный формат JSON")
		return
	}
	loan, err := h.service.BorrowBook(r.Context(), req.BookID, req.ReaderID, req.Days)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, loan)
}

func (h *LoanHandler) Return(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	loan, err := h.service.ReturnBook(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, loan)
}

func (h *LoanHandler) GetActive(w http.ResponseWriter, r *http.Request) {
	loans, err := h.service.GetActiveLoans(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, loans)
}

func (h *LoanHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStatistics(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, stats)
}
