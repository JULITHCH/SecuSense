package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/secusense/backend/internal/domain"
	"github.com/secusense/backend/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
)

type UseCase struct {
	userRepo         domain.UserRepository
	refreshTokenRepo domain.RefreshTokenRepository
	jwtManager       *jwt.Manager
}

func NewUseCase(
	userRepo domain.UserRepository,
	refreshTokenRepo domain.RefreshTokenRepository,
	jwtManager *jwt.Manager,
) *UseCase {
	return &UseCase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
	}
}

func (uc *UseCase) Register(req *domain.CreateUserRequest) (*domain.AuthResponse, error) {
	// Check if user already exists
	existingUser, err := uc.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &domain.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         domain.RoleUser,
	}

	if err := uc.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Generate tokens
	return uc.generateAuthResponse(user)
}

func (uc *UseCase) Login(req *domain.LoginRequest) (*domain.AuthResponse, error) {
	// Find user
	user, err := uc.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	return uc.generateAuthResponse(user)
}

func (uc *UseCase) RefreshToken(refreshToken string) (*domain.AuthResponse, error) {
	// Hash the token
	tokenHash := jwt.HashToken(refreshToken)

	// Find the refresh token
	storedToken, err := uc.refreshTokenRepo.GetByHash(tokenHash)
	if err != nil {
		return nil, err
	}
	if storedToken == nil {
		return nil, ErrInvalidRefreshToken
	}

	// Check if token is expired
	if time.Now().After(storedToken.ExpiresAt) {
		return nil, ErrInvalidRefreshToken
	}

	// Get user
	user, err := uc.userRepo.GetByID(storedToken.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidRefreshToken
	}

	// Delete old refresh tokens for this user
	if err := uc.refreshTokenRepo.DeleteByUserID(user.ID); err != nil {
		return nil, err
	}

	// Generate new tokens
	return uc.generateAuthResponse(user)
}

func (uc *UseCase) GetCurrentUser(userID uuid.UUID) (*domain.User, error) {
	return uc.userRepo.GetByID(userID)
}

func (uc *UseCase) generateAuthResponse(user *domain.User) (*domain.AuthResponse, error) {
	tokenPair, err := uc.jwtManager.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}

	// Store refresh token
	refreshToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: jwt.HashToken(tokenPair.RefreshToken),
		ExpiresAt: time.Now().Add(uc.jwtManager.GetRefreshExpiry()),
	}

	if err := uc.refreshTokenRepo.Create(refreshToken); err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		User:         *user,
	}, nil
}
