package service

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/kento/driver/backend/internal/dto"
	"github.com/kento/driver/backend/internal/model"
	"github.com/kento/driver/backend/internal/repository"
	"github.com/kento/driver/backend/pkg/apperror"
	jwtpkg "github.com/kento/driver/backend/pkg/jwt"
)

type AuthService struct {
	userRepo     *repository.UserRepo
	jwtSecret    string
	accessExpiry time.Duration
	refreshExpiry time.Duration
}

func NewAuthService(userRepo *repository.UserRepo, jwtSecret string, accessExpiry, refreshExpiry time.Duration) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		jwtSecret:     jwtSecret,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.GetByEmployeeID(ctx, req.EmployeeID)
	if err != nil {
		return nil, apperror.ErrInternal
	}
	if user == nil || !user.IsActive {
		return nil, apperror.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperror.ErrInvalidCredentials
	}

	accessToken, err := jwtpkg.GenerateAccessToken(s.jwtSecret, s.accessExpiry, user.ID, user.EmployeeID, string(user.Role))
	if err != nil {
		return nil, apperror.ErrInternal
	}

	refreshToken, err := jwtpkg.GenerateRefreshToken(s.jwtSecret, s.refreshExpiry, user.ID, user.EmployeeID, string(user.Role))
	if err != nil {
		return nil, apperror.ErrInternal
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserInfo{
			ID:            user.ID,
			EmployeeID:    user.EmployeeID,
			Name:          user.Name,
			Role:          string(user.Role),
			PriorityLevel: user.PriorityLevel,
		},
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*dto.RefreshResponse, error) {
	claims, err := jwtpkg.Parse(refreshToken, s.jwtSecret)
	if err != nil {
		return nil, apperror.ErrUnauthorized
	}
	if claims.TokenType != "refresh" {
		return nil, apperror.ErrUnauthorized
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil || user == nil || !user.IsActive {
		return nil, apperror.ErrUnauthorized
	}

	accessToken, err := jwtpkg.GenerateAccessToken(s.jwtSecret, s.accessExpiry, user.ID, user.EmployeeID, string(user.Role))
	if err != nil {
		return nil, apperror.ErrInternal
	}

	return &dto.RefreshResponse{AccessToken: accessToken}, nil
}

func (s *AuthService) GetUser(ctx context.Context, userID string) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}
