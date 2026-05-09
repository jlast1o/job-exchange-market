package user

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jlast1o/job-exchange/internal/auth"
)

var (
	ErrInvalidCredentials = errors.New("Инвалидные почта или пароль")
	ErrUserBanned         = errors.New("Юзер забанен")
)

type Service struct {
	repo      Repository
	pool      *pgxpool.Pool
	hasher    auth.PasswordHasher
	tokens    *auth.TokenGenerator
	tokenRepo auth.RefreshTokenRepository
}

func NewService(repo Repository, pool *pgxpool.Pool, hasher auth.PasswordHasher, tokens *auth.TokenGenerator, tokenRepo auth.RefreshTokenRepository) *Service {
	return &Service{repo, pool, hasher, tokens, tokenRepo}
}

type RegisterRequest struct {
	Email    string
	Password string
	FullName string
	Role     string
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

// Регистрирует пользователя. Используем для атомарности транзакцию.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	hashed, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	role := req.Role
	if role == "" {
		role = "user"
	}

	u := &User{
		Email:        req.Email,
		PasswordHash: hashed,
		FullName:     req.FullName,
		Role:         role,
	}

	// Репозиторий пользователя внутри транзакции
	txUserRepo := s.repo.WithTx(tx)
	if err := txUserRepo.Create(ctx, u); err != nil {
		if errors.Is(err, ErrAlreadyExist) {
			return nil, ErrAlreadyExist
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	access, err := s.tokens.GenerateAccessToken(u.ID, u.Email, u.Role)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}
	refresh, err := s.tokens.GenerateRefreshToken(u.ID)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	// Репозиторий токенов внутри той же транзакции
	txTokenRepo := s.tokenRepo.WithTx(tx)
	refreshHash := hashToken(refresh)
	expiresAt := time.Now().Add(s.tokens.RefreshTTL())
	if err := txTokenRepo.Store(ctx, u.ID, refreshHash, expiresAt); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		User:         *u,
	}, nil
}

type LoginRequest struct {
	Email    string
	Password string
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	u, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("Ошибка получения пользователя по эмейлу: %w", err)
	}

	if u.IsBanned {
		return nil, ErrUserBanned
	}

	if err := s.hasher.Check(req.Password, u.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	access, err := s.tokens.GenerateAccessToken(u.ID, u.Email, u.Role)
	if err != nil {
		return nil, fmt.Errorf("При логине возникла проблема с access токен: %w", err)
	}

	refresh, err := s.tokens.GenerateRefreshToken(u.ID)
	if err != nil {
		return nil, fmt.Errorf("При логине возникла проблема с refresh токен: %w", err)
	}

	refreshHash := hashToken(refresh)
	expiresAt := time.Now().Add(s.tokens.RefreshTTL())
	if err := s.tokenRepo.Store(ctx, u.ID, refreshHash, expiresAt); err != nil {
		return nil, fmt.Errorf("Возникла проблема при сохраниении refresh токена: %w", err)
	}

	return &AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		User:         *u,
	}, nil
}

func hashToken(refresh string) string {
	hash := sha256.Sum256([]byte(refresh))
	return hex.EncodeToString(hash[:])
}
