package main

import (
	"bufio"
	"context"
	"library-management-api/internal/database"
	"library-management-api/internal/handlers"
	"library-management-api/internal/middleware"
	"library-management-api/internal/repository"
	"library-management-api/internal/router"
	"library-management-api/internal/services"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if k, v, ok := strings.Cut(line, "="); ok {
			_ = os.Setenv(strings.TrimSpace(k), strings.TrimSpace(v))
		}
	}
}

func main() {
	loadEnvFile(".env")

	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}
	defer db.Close()

	bookRepo := repository.NewBookRepository(db)
	readerRepo := repository.NewReaderRepository(db)
	loanRepo := repository.NewLoanRepository(db)
	userRepo := repository.NewUserRepository(db)

	bookService := services.NewBookService(bookRepo, loanRepo)
	readerService := services.NewReaderService(readerRepo, loanRepo)
	loanService := services.NewLoanService(loanRepo, bookRepo, readerRepo)
	authService := services.NewAuthService(userRepo)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := authService.SeedAdmin(ctx); err != nil {
		cancel()
		log.Fatalf("Ошибка создания администратора: %v", err)
	}
	cancel()

	bookHandler := handlers.NewBookHandler(bookService)
	readerHandler := handlers.NewReaderHandler(readerService)
	loanHandler := handlers.NewLoanHandler(loanService)
	authHandler := handlers.NewAuthHandler(authService, loanService)

	r := router.NewRouter(bookHandler, readerHandler, loanHandler, authHandler)
	handler := middleware.Logging(r)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Library Management API (PostgreSQL + JWT) запущен на порту :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
