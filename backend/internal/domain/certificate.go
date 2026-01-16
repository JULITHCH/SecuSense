package domain

import (
	"time"

	"github.com/google/uuid"
)

type Certificate struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	CertificateNumber string     `db:"certificate_number" json:"certificateNumber"`
	UserID            uuid.UUID  `db:"user_id" json:"userId"`
	CourseID          uuid.UUID  `db:"course_id" json:"courseId"`
	TestAttemptID     uuid.UUID  `db:"test_attempt_id" json:"testAttemptId"`
	PDFURL            *string    `db:"pdf_url" json:"pdfUrl,omitempty"`
	VerificationHash  string     `db:"verification_hash" json:"verificationHash"`
	IssuedAt          time.Time  `db:"issued_at" json:"issuedAt"`
	ExpiresAt         *time.Time `db:"expires_at" json:"expiresAt,omitempty"`

	// Joined fields
	UserFirstName string `db:"user_first_name" json:"userFirstName,omitempty"`
	UserLastName  string `db:"user_last_name" json:"userLastName,omitempty"`
	UserEmail     string `db:"user_email" json:"userEmail,omitempty"`
	CourseTitle   string `db:"course_title" json:"courseTitle,omitempty"`
	Score         int    `db:"score" json:"score,omitempty"`
	MaxScore      int    `db:"max_score" json:"maxScore,omitempty"`
}

type CertificateVerification struct {
	Valid            bool      `json:"valid"`
	CertificateNumber string   `json:"certificateNumber,omitempty"`
	HolderName       string    `json:"holderName,omitempty"`
	CourseTitle      string    `json:"courseTitle,omitempty"`
	IssuedAt         time.Time `json:"issuedAt,omitempty"`
	Score            int       `json:"score,omitempty"`
	MaxScore         int       `json:"maxScore,omitempty"`
}

type CertificateRepository interface {
	Create(cert *Certificate) error
	GetByID(id uuid.UUID) (*Certificate, error)
	GetByVerificationHash(hash string) (*Certificate, error)
	GetByUserID(userID uuid.UUID) ([]*Certificate, error)
	GetByUserAndCourse(userID, courseID uuid.UUID) (*Certificate, error)
	Update(cert *Certificate) error
	GenerateCertificateNumber() (string, error)
}
