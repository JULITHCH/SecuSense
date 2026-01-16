package postgres

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/secusense/backend/internal/domain"
)

type CertificateRepository struct {
	db *sqlx.DB
}

func NewCertificateRepository(db *sqlx.DB) *CertificateRepository {
	return &CertificateRepository{db: db}
}

func (r *CertificateRepository) Create(cert *domain.Certificate) error {
	query := `
		INSERT INTO certificates (id, certificate_number, user_id, course_id, test_attempt_id, pdf_url, verification_hash, issued_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), $8)
		RETURNING issued_at`

	if cert.ID == uuid.Nil {
		cert.ID = uuid.New()
	}

	return r.db.QueryRow(
		query,
		cert.ID, cert.CertificateNumber, cert.UserID, cert.CourseID, cert.TestAttemptID,
		cert.PDFURL, cert.VerificationHash, cert.ExpiresAt,
	).Scan(&cert.IssuedAt)
}

func (r *CertificateRepository) GetByID(id uuid.UUID) (*domain.Certificate, error) {
	var cert domain.Certificate
	query := `
		SELECT c.id, c.certificate_number, c.user_id, c.course_id, c.test_attempt_id, c.pdf_url,
		       c.verification_hash, c.issued_at, c.expires_at,
		       u.first_name as user_first_name, u.last_name as user_last_name, u.email as user_email,
		       co.title as course_title,
		       ta.score as score, ta.max_score as max_score
		FROM certificates c
		JOIN users u ON c.user_id = u.id
		JOIN courses co ON c.course_id = co.id
		JOIN test_attempts ta ON c.test_attempt_id = ta.id
		WHERE c.id = $1`

	err := r.db.Get(&cert, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

func (r *CertificateRepository) GetByVerificationHash(hash string) (*domain.Certificate, error) {
	var cert domain.Certificate
	query := `
		SELECT c.id, c.certificate_number, c.user_id, c.course_id, c.test_attempt_id, c.pdf_url,
		       c.verification_hash, c.issued_at, c.expires_at,
		       u.first_name as user_first_name, u.last_name as user_last_name, u.email as user_email,
		       co.title as course_title,
		       ta.score as score, ta.max_score as max_score
		FROM certificates c
		JOIN users u ON c.user_id = u.id
		JOIN courses co ON c.course_id = co.id
		JOIN test_attempts ta ON c.test_attempt_id = ta.id
		WHERE c.verification_hash = $1`

	err := r.db.Get(&cert, query, hash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

func (r *CertificateRepository) GetByUserID(userID uuid.UUID) ([]*domain.Certificate, error) {
	var certs []*domain.Certificate
	query := `
		SELECT c.id, c.certificate_number, c.user_id, c.course_id, c.test_attempt_id, c.pdf_url,
		       c.verification_hash, c.issued_at, c.expires_at,
		       u.first_name as user_first_name, u.last_name as user_last_name, u.email as user_email,
		       co.title as course_title,
		       ta.score as score, ta.max_score as max_score
		FROM certificates c
		JOIN users u ON c.user_id = u.id
		JOIN courses co ON c.course_id = co.id
		JOIN test_attempts ta ON c.test_attempt_id = ta.id
		WHERE c.user_id = $1
		ORDER BY c.issued_at DESC`

	err := r.db.Select(&certs, query, userID)
	if err != nil {
		return nil, err
	}
	return certs, nil
}

func (r *CertificateRepository) GetByUserAndCourse(userID, courseID uuid.UUID) (*domain.Certificate, error) {
	var cert domain.Certificate
	query := `
		SELECT c.id, c.certificate_number, c.user_id, c.course_id, c.test_attempt_id, c.pdf_url,
		       c.verification_hash, c.issued_at, c.expires_at,
		       u.first_name as user_first_name, u.last_name as user_last_name, u.email as user_email,
		       co.title as course_title,
		       ta.score as score, ta.max_score as max_score
		FROM certificates c
		JOIN users u ON c.user_id = u.id
		JOIN courses co ON c.course_id = co.id
		JOIN test_attempts ta ON c.test_attempt_id = ta.id
		WHERE c.user_id = $1 AND c.course_id = $2
		ORDER BY c.issued_at DESC LIMIT 1`

	err := r.db.Get(&cert, query, userID, courseID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

func (r *CertificateRepository) Update(cert *domain.Certificate) error {
	query := `
		UPDATE certificates
		SET pdf_url = $1, expires_at = $2
		WHERE id = $3`

	_, err := r.db.Exec(query, cert.PDFURL, cert.ExpiresAt, cert.ID)
	return err
}

func (r *CertificateRepository) GenerateCertificateNumber() (string, error) {
	year := time.Now().Year()
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("SS-%d-%s", year, hex.EncodeToString(bytes)), nil
}
