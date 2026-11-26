package services

import (
	"context"
	"library-management-api/internal/models"
	"library-management-api/internal/repository"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "super-secret-jwt-key-change-in-production"
	}
	return []byte(secret)
}

func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.User, error) {
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Email) == "" || !strings.Contains(req.Email, "@") {
		return nil, NewAppError(http.StatusBadRequest, "INVALID_REGISTER_DATA", "Укажите корректные username и email")
	}
	if len(req.Password) < 6 {
		return nil, NewAppError(http.StatusBadRequest, "WEAK_PASSWORD", "Пароль должен содержать минимум 6 символов")
	}

	exists, err := s.userRepo.ExistsByUsernameOrEmail(ctx, req.Username, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, NewAppError(http.StatusConflict, "USER_ALREADY_EXISTS", "Пользователь с таким username или email уже существует")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:     strings.TrimSpace(req.Username),
		Email:        strings.TrimSpace(req.Email),
		PasswordHash: string(hash),
		Role:         "USER",
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, NewAppError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "Неверный email или пароль")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, NewAppError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "Неверный email или пароль")
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(getJWTSecret())
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{Token: signed, User: *user}, nil
}

func (s *AuthService) GetCurrentUser(ctx context.Context, userID int) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, NewAppError(http.StatusNotFound, "USER_NOT_FOUND", "Пользователь не найден")
	}
	return user, nil
}

func (s *AuthService) SeedAdmin(ctx context.Context) error {
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.userRepo.EnsureAdmin(ctx, "admin", "admin@library.local", string(hash))
}
