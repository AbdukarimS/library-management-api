package handlers

import (
	"errors"
	"library-management-api/internal/response"
	"library-management-api/internal/services"
	"net/http"
	"strconv"
)

func parseID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		return 0, services.NewAppError(http.StatusBadRequest, "INVALID_ID", "Идентификатор должен быть положительным числом")
	}
	return id, nil
}

func handleServiceError(w http.ResponseWriter, err error) {
	var appErr *services.AppError
	if errors.As(err, &appErr) {
		response.Error(w, appErr.Status, appErr.Code, appErr.Message)
		return
	}
	response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера")
}
