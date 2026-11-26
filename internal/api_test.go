package internal_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"library-management-api/internal/database"
	"library-management-api/internal/handlers"
	"library-management-api/internal/models"
	"library-management-api/internal/repository"
	"library-management-api/internal/response"
	"library-management-api/internal/router"
	"library-management-api/internal/services"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func loadTestEnv() {
	for _, path := range []string{".env", "../.env"} {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
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
		_ = scanner.Err()
		_ = f.Close()
		break
	}
}

func setupTestServer(t *testing.T) (http.Handler, *repository.BookRepository, *repository.ReaderRepository, *repository.LoanRepository) {
	t.Helper()
	loadTestEnv()
	_ = os.Setenv("JWT_SECRET", "test-secret-key")

	db, err := database.NewPostgresDB()
	if err != nil {
		t.Fatalf("PostgreSQL недоступен для теста: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	bookRepo := repository.NewBookRepository(db)
	readerRepo := repository.NewReaderRepository(db)
	loanRepo := repository.NewLoanRepository(db)
	userRepo := repository.NewUserRepository(db)

	bookSvc := services.NewBookService(bookRepo, loanRepo)
	readerSvc := services.NewReaderService(readerRepo, loanRepo)
	loanSvc := services.NewLoanService(loanRepo, bookRepo, readerRepo)
	authSvc := services.NewAuthService(userRepo)

	r := router.NewRouter(
		handlers.NewBookHandler(bookSvc),
		handlers.NewReaderHandler(readerSvc),
		handlers.NewLoanHandler(loanSvc),
		handlers.NewAuthHandler(authSvc, loanSvc),
	)
	return r, bookRepo, readerRepo, loanRepo
}

func makeToken(userID int, role string, exp time.Time) string {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     exp.Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := tok.SignedString([]byte("test-secret-key"))
	return s
}

func TestHealthEndpoint(t *testing.T) {
	srv, _, _, _ := setupTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200, получен %d", rec.Code)
	}
}

func TestHealthDBEndpoint(t *testing.T) {
	srv, _, _, _ := setupTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/health/db", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200, получен %d", rec.Code)
	}
}

func TestUnknownRouteReturnsJSON404(t *testing.T) {
	srv, _, _, _ := setupTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/unknown", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("ожидался статус 404, получен %d", rec.Code)
	}
	var resp response.APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil || resp.IsSuccess {
		t.Fatalf("ожидался единый JSON-ответ с ошибкой")
	}
}

func TestUnauthorizedWithoutToken(t *testing.T) {
	srv, _, _, _ := setupTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/books", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("ожидался статус 401, получен %d", rec.Code)
	}
}

func TestExpiredTokenReturns401(t *testing.T) {
	srv, _, _, _ := setupTestServer(t)
	expired := makeToken(1, "ADMIN", time.Now().Add(-1*time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/api/books", nil)
	req.Header.Set("Authorization", "Bearer "+expired)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("ожидался статус 401 для просроченного токена, получен %d", rec.Code)
	}
}

func TestRoleAuthorizationUserVsAdmin(t *testing.T) {
	srv, _, _, _ := setupTestServer(t)
	userToken := makeToken(2, "USER", time.Now().Add(1*time.Hour))

	reqGet := httptest.NewRequest(http.MethodGet, "/api/books", nil)
	reqGet.Header.Set("Authorization", "Bearer "+userToken)
	recGet := httptest.NewRecorder()
	srv.ServeHTTP(recGet, reqGet)
	if recGet.Code != http.StatusOK {
		t.Fatalf("USER должен иметь доступ к GET /api/books, получен %d", recGet.Code)
	}

	body := bytes.NewBufferString(`{"title":"Test","author":"Author","year":2020,"genre":"IT","totalCopies":1}`)
	reqPost := httptest.NewRequest(http.MethodPost, "/api/books", body)
	reqPost.Header.Set("Authorization", "Bearer "+userToken)
	recPost := httptest.NewRecorder()
	srv.ServeHTTP(recPost, reqPost)
	if recPost.Code != http.StatusForbidden {
		t.Fatalf("USER не должен создавать книги (ожидался 403), получен %d", recPost.Code)
	}
}

func TestRegisterAndLoginFlow(t *testing.T) {
	srv, _, _, _ := setupTestServer(t)
	uniq := time.Now().UnixNano()
	email := fmt.Sprintf("user_%d@test.local", uniq)
	username := fmt.Sprintf("user_%d", uniq)

	regBody := bytes.NewBufferString(fmt.Sprintf(`{"username":"%s","email":"%s","password":"secret123"}`, username, email))
	recReg := httptest.NewRecorder()
	srv.ServeHTTP(recReg, httptest.NewRequest(http.MethodPost, "/api/auth/register", regBody))
	if recReg.Code != http.StatusCreated {
		t.Fatalf("ожидался 201 при регистрации, получен %d", recReg.Code)
	}

	loginBody := bytes.NewBufferString(fmt.Sprintf(`{"email":"%s","password":"secret123"}`, email))
	recLogin := httptest.NewRecorder()
	srv.ServeHTTP(recLogin, httptest.NewRequest(http.MethodPost, "/api/auth/login", loginBody))
	if recLogin.Code != http.StatusOK {
		t.Fatalf("ожидался 200 при входе, получен %d", recLogin.Code)
	}
}

func TestPostgresBookCRUD(t *testing.T) {
	srv, _, _, _ := setupTestServer(t)
	adminToken := makeToken(1, "ADMIN", time.Now().Add(1*time.Hour))

	createBody := bytes.NewBufferString(`{"title":"Integration Book","author":"Tester","year":2022,"genre":"Science","totalCopies":2}`)
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/books", createBody)
	reqCreate.Header.Set("Authorization", "Bearer "+adminToken)
	recCreate := httptest.NewRecorder()
	srv.ServeHTTP(recCreate, reqCreate)
	if recCreate.Code != http.StatusCreated {
		t.Fatalf("ожидался 201, получен %d", recCreate.Code)
	}

	var createdResp struct {
		Result models.Book `json:"result"`
	}
	_ = json.NewDecoder(recCreate.Body).Decode(&createdResp)
	bookID := createdResp.Result.ID

	reqDel := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/books/%d", bookID), nil)
	reqDel.Header.Set("Authorization", "Bearer "+adminToken)
	recDel := httptest.NewRecorder()
	srv.ServeHTTP(recDel, reqDel)
	if recDel.Code != http.StatusOK {
		t.Fatalf("ожидался 200 при удалении, получен %d", recDel.Code)
	}
}

func TestSQLInjectionSafeSearch(t *testing.T) {
	srv, _, _, _ := setupTestServer(t)
	userToken := makeToken(2, "USER", time.Now().Add(1*time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/api/books/search?author='+OR+'1'='1", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("поиск с SQL-инъекцией должен безопасно вернуть 200, получен %d", rec.Code)
	}
}

func TestConcurrentBorrowLastCopyTransaction(t *testing.T) {
	srv, bookRepo, readerRepo, _ := setupTestServer(t)
	ctx := context.Background()
	adminToken := makeToken(1, "ADMIN", time.Now().Add(1*time.Hour))

	book := &models.Book{Title: "Last Copy Book", Author: "Concurrent", Year: 2023, Genre: "IT", TotalCopies: 1, AvailableCopies: 1}
	if err := bookRepo.Create(ctx, book); err != nil {
		t.Fatal(err)
	}
	uniq := time.Now().UnixNano()
	r1 := &models.Reader{FirstName: "R1", LastName: "Test", Email: fmt.Sprintf("r1_%d@test.local", uniq), Age: 25, IsActive: true}
	r2 := &models.Reader{FirstName: "R2", LastName: "Test", Email: fmt.Sprintf("r2_%d@test.local", uniq), Age: 26, IsActive: true}
	_ = readerRepo.Create(ctx, r1)
	_ = readerRepo.Create(ctx, r2)

	var wg sync.WaitGroup
	codes := make([]int, 2)
	readers := []int{r1.ID, r2.ID}

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			payload := bytes.NewBufferString(fmt.Sprintf(`{"bookId":%d,"readerId":%d,"days":7}`, book.ID, readers[idx]))
			req := httptest.NewRequest(http.MethodPost, "/api/loans", payload)
			req.Header.Set("Authorization", "Bearer "+adminToken)
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, req)
			codes[idx] = rec.Code
		}(i)
	}
	wg.Wait()

	successCount := 0
	for _, c := range codes {
		if c == http.StatusCreated {
			successCount++
		}
	}
	if successCount != 1 {
		t.Fatalf("ожидалась ровно 1 успешная выдача последнего экземпляра, получено %d (коды: %v)", successCount, codes)
	}
}

func TestStatisticsEndpoint(t *testing.T) {
	srv, _, _, _ := setupTestServer(t)
	adminToken := makeToken(1, "ADMIN", time.Now().Add(1*time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/api/statistics", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался статус 200 для статистики, получен %d", rec.Code)
	}
}
