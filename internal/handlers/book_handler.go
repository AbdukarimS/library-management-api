package handlers

import (
	"encoding/json"
	"library-management-api/internal/models"
	"library-management-api/internal/response"
	"library-management-api/internal/services"
	"net/http"
	"strconv"
)

type BookHandler struct {
	service *services.BookService
}

func NewBookHandler(service *services.BookService) *BookHandler {
	return &BookHandler{service: service}
}

func (h *BookHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	books, err := h.service.GetAllBooks(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, books)
}

func (h *BookHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	book, err := h.service.GetBookByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, book)
}

func (h *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.Book
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Некорректный формат JSON")
		return
	}
	book, err := h.service.CreateBook(r.Context(), input)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, book)
}

func (h *BookHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	var input models.Book
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Некорректный формат JSON")
		return
	}
	book, err := h.service.UpdateBook(r.Context(), id, input)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, book)
}

func (h *BookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	if err := h.service.DeleteBook(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "Книга успешно удалена"})
}

func (h *BookHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	year, _ := strconv.Atoi(q.Get("year"))
	books, err := h.service.SearchBooks(r.Context(), q.Get("title"), q.Get("author"), q.Get("genre"), year)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, books)
}
