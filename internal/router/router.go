package router

import (
	"library-management-api/internal/handlers"
	"library-management-api/internal/middleware"
	"library-management-api/internal/response"
	"net/http"
)

type jsonStatusWrapper struct {
	http.ResponseWriter
	intercepted bool
}

func (w *jsonStatusWrapper) WriteHeader(code int) {
	switch code {
	case http.StatusNotFound:
		w.intercepted = true
		response.Error(w.ResponseWriter, http.StatusNotFound, "ROUTE_NOT_FOUND", "Запрашиваемый маршрут не найден")
	case http.StatusMethodNotAllowed:
		w.intercepted = true
		response.Error(w.ResponseWriter, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "HTTP-метод не поддерживается для данного маршрута")
	default:
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *jsonStatusWrapper) Write(b []byte) (int, error) {
	if w.intercepted {
		return len(b), nil
	}
	return w.ResponseWriter.Write(b)
}

func protect(h http.HandlerFunc, roles ...string) http.Handler {
	return middleware.Authenticate(middleware.RequireRoles(roles...)(h))
}

func NewRouter(
	bookH *handlers.BookHandler,
	readerH *handlers.ReaderHandler,
	loanH *handlers.LoanHandler,
	authH *handlers.AuthHandler,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})
	mux.HandleFunc("GET /books", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/books.html")
	})
	mux.HandleFunc("GET /readers", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/readers.html")
	})
	mux.HandleFunc("GET /loans", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/loans.html")
	})
	mux.HandleFunc("GET /style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/style.css")
	})
	mux.HandleFunc("GET /app.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/app.js")
	})

	mux.HandleFunc("GET /api/health", authH.Health)
	mux.HandleFunc("GET /api/health/db", authH.HealthDB)
	mux.HandleFunc("POST /api/auth/register", authH.Register)
	mux.HandleFunc("POST /api/auth/login", authH.Login)

	mux.Handle("GET /api/auth/me", protect(authH.Me, "USER", "ADMIN"))

	mux.Handle("GET /api/books", protect(bookH.GetAll, "USER", "ADMIN"))
	mux.Handle("GET /api/books/search", protect(bookH.Search, "USER", "ADMIN"))
	mux.Handle("GET /api/books/{id}", protect(bookH.GetByID, "USER", "ADMIN"))
	mux.Handle("POST /api/books", protect(bookH.Create, "ADMIN"))
	mux.Handle("PUT /api/books/{id}", protect(bookH.Update, "ADMIN"))
	mux.Handle("DELETE /api/books/{id}", protect(bookH.Delete, "ADMIN"))

	mux.Handle("GET /api/readers", protect(readerH.GetAll, "ADMIN"))
	mux.Handle("GET /api/readers/search", protect(readerH.Search, "ADMIN"))
	mux.Handle("GET /api/readers/{id}", protect(readerH.GetByID, "ADMIN"))
	mux.Handle("POST /api/readers", protect(readerH.Create, "ADMIN"))
	mux.Handle("PUT /api/readers/{id}", protect(readerH.Update, "ADMIN"))
	mux.Handle("DELETE /api/readers/{id}", protect(readerH.Delete, "ADMIN"))
	mux.Handle("GET /api/readers/{id}/loans", protect(readerH.GetLoans, "ADMIN"))

	mux.Handle("GET /api/loans", protect(loanH.GetAll, "ADMIN"))
	mux.Handle("GET /api/loans/active", protect(loanH.GetActive, "ADMIN"))
	mux.Handle("GET /api/loans/{id}", protect(loanH.GetByID, "ADMIN"))
	mux.Handle("POST /api/loans", protect(loanH.Borrow, "ADMIN"))
	mux.Handle("PUT /api/loans/{id}/return", protect(loanH.Return, "ADMIN"))
	mux.Handle("GET /api/statistics", protect(loanH.GetStatistics, "ADMIN"))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.ServeHTTP(&jsonStatusWrapper{ResponseWriter: w}, r)
	})
}
