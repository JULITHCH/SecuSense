package postgres

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/secusense/backend/internal/domain"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING created_at, updated_at`

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	return r.db.QueryRow(
		query,
		user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName, user.Role,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, email, password_hash, first_name, last_name, role, created_at, updated_at
			  FROM users WHERE id = $1`

	err := r.db.Get(&user, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, email, password_hash, first_name, last_name, role, created_at, updated_at
			  FROM users WHERE email = $1`

	err := r.db.Get(&user, query, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *domain.User) error {
	query := `
		UPDATE users
		SET email = $1, first_name = $2, last_name = $3, role = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at`

	return r.db.QueryRow(
		query,
		user.Email, user.FirstName, user.LastName, user.Role, user.ID,
	).Scan(&user.UpdatedAt)
}

func (r *UserRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *UserRepository) List(limit, offset int) ([]*domain.User, error) {
	var users []*domain.User
	query := `SELECT id, email, password_hash, first_name, last_name, role, created_at, updated_at
			  FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	err := r.db.Select(&users, query, limit, offset)
	if err != nil {
		return nil, err
	}
	return users, nil
}

type RefreshTokenRepository struct {
	db *sqlx.DB
}

func NewRefreshTokenRepository(db *sqlx.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, NOW())`

	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}

	_, err := r.db.Exec(query, token.ID, token.UserID, token.TokenHash, token.ExpiresAt)
	return err
}

func (r *RefreshTokenRepository) GetByHash(hash string) (*domain.RefreshToken, error) {
	var token domain.RefreshToken
	query := `SELECT id, user_id, token_hash, expires_at, created_at
			  FROM refresh_tokens WHERE token_hash = $1`

	err := r.db.Get(&token, query, hash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *RefreshTokenRepository) DeleteByUserID(userID uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.db.Exec(query, userID)
	return err
}

func (r *RefreshTokenRepository) DeleteExpired() error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`
	_, err := r.db.Exec(query)
	return err
}
