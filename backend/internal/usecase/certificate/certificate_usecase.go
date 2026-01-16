package certificate

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/secusense/backend/internal/domain"
)

var (
	ErrCertificateNotFound    = errors.New("certificate not found")
	ErrTestNotPassed          = errors.New("test must be passed to get certificate")
	ErrCertificateExists      = errors.New("certificate already exists for this course")
	ErrAttemptNotFound        = errors.New("test attempt not found")
)

type PDFGenerator interface {
	Generate(cert *domain.Certificate) ([]byte, error)
}

type UseCase struct {
	certRepo    domain.CertificateRepository
	attemptRepo domain.TestAttemptRepository
	pdfGen      PDFGenerator
}

func NewUseCase(
	certRepo domain.CertificateRepository,
	attemptRepo domain.TestAttemptRepository,
	pdfGen PDFGenerator,
) *UseCase {
	return &UseCase{
		certRepo:    certRepo,
		attemptRepo: attemptRepo,
		pdfGen:      pdfGen,
	}
}

func (uc *UseCase) Generate(userID, courseID, attemptID uuid.UUID) (*domain.Certificate, error) {
	// Check if certificate already exists
	existing, err := uc.certRepo.GetByUserAndCourse(userID, courseID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrCertificateExists
	}

	// Get the attempt and verify it passed
	attempt, err := uc.attemptRepo.GetByID(attemptID)
	if err != nil {
		return nil, err
	}
	if attempt == nil {
		return nil, ErrAttemptNotFound
	}
	if attempt.UserID != userID {
		return nil, ErrAttemptNotFound
	}
	if attempt.Passed == nil || !*attempt.Passed {
		return nil, ErrTestNotPassed
	}

	// Generate certificate number
	certNumber, err := uc.certRepo.GenerateCertificateNumber()
	if err != nil {
		return nil, err
	}

	// Generate verification hash
	hashData := fmt.Sprintf("%s-%s-%s-%s", certNumber, userID, courseID, time.Now().Format(time.RFC3339))
	hash := sha256.Sum256([]byte(hashData))
	verificationHash := hex.EncodeToString(hash[:])[:16]

	cert := &domain.Certificate{
		ID:                uuid.New(),
		CertificateNumber: certNumber,
		UserID:            userID,
		CourseID:          courseID,
		TestAttemptID:     attemptID,
		VerificationHash:  verificationHash,
	}

	if err := uc.certRepo.Create(cert); err != nil {
		return nil, err
	}

	// Get full certificate data for PDF generation
	fullCert, err := uc.certRepo.GetByID(cert.ID)
	if err != nil {
		return nil, err
	}

	return fullCert, nil
}

func (uc *UseCase) GetByID(id uuid.UUID) (*domain.Certificate, error) {
	cert, err := uc.certRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if cert == nil {
		return nil, ErrCertificateNotFound
	}
	return cert, nil
}

func (uc *UseCase) GetByUserID(userID uuid.UUID) ([]*domain.Certificate, error) {
	return uc.certRepo.GetByUserID(userID)
}

func (uc *UseCase) Verify(hash string) (*domain.CertificateVerification, error) {
	cert, err := uc.certRepo.GetByVerificationHash(hash)
	if err != nil {
		return nil, err
	}
	if cert == nil {
		return &domain.CertificateVerification{
			Valid: false,
		}, nil
	}

	// Check expiration
	if cert.ExpiresAt != nil && time.Now().After(*cert.ExpiresAt) {
		return &domain.CertificateVerification{
			Valid: false,
		}, nil
	}

	return &domain.CertificateVerification{
		Valid:            true,
		CertificateNumber: cert.CertificateNumber,
		HolderName:       fmt.Sprintf("%s %s", cert.UserFirstName, cert.UserLastName),
		CourseTitle:      cert.CourseTitle,
		IssuedAt:         cert.IssuedAt,
		Score:            cert.Score,
		MaxScore:         cert.MaxScore,
	}, nil
}

func (uc *UseCase) GeneratePDF(id uuid.UUID, userID uuid.UUID) ([]byte, error) {
	cert, err := uc.certRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if cert == nil {
		return nil, ErrCertificateNotFound
	}
	if cert.UserID != userID {
		return nil, ErrCertificateNotFound
	}

	return uc.pdfGen.Generate(cert)
}
