package handlers

import (
	"encoding/json"
	"library-management-api/internal/models"
	"library-management-api/internal/response"
	"library-management-api/internal/services"
	"net/http"
)

type ReaderHandler struct {
	service *services.ReaderService
}

func NewReaderHandler(service *services.ReaderService) *ReaderHandler {
	return &ReaderHandler{service: service}
}

func (h *ReaderHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	if name := r.URL.Query().Get("name"); name != "" {
		readers, err := h.service.SearchReadersByName(r.Context(), name)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		response.JSON(w, http.StatusOK, readers)
		return
	}
	readers, err := h.service.GetAllReaders(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, readers)
}

func (h *ReaderHandler) Search(w http.ResponseWriter, r *http.Request) {
	readers, err := h.service.SearchReadersByName(r.Context(), r.URL.Query().Get("name"))
	if err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, readers)
}

func (h *ReaderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	reader, err := h.service.GetReaderByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, reader)
}

func (h *ReaderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.Reader
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Некорректный формат JSON")
		return
	}
	reader, err := h.service.CreateReader(r.Context(), input)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, reader)
}

func (h *ReaderHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	var input models.Reader
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Некорректный формат JSON")
		return
	}
	reader, err := h.service.UpdateReader(r.Context(), id, input)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, reader)
}

func (h *ReaderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	if err := h.service.DeleteReader(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "Читатель успешно удалён"})
}

func (h *ReaderHandler) GetLoans(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	loans, err := h.service.GetReaderLoans(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, loans)
}
