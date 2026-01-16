package enrollment

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/secusense/backend/internal/domain"
)

var (
	ErrCourseNotFound      = errors.New("course not found")
	ErrCourseNotPublished  = errors.New("course is not published")
	ErrAlreadyEnrolled     = errors.New("already enrolled in this course")
	ErrEnrollmentNotFound  = errors.New("enrollment not found")
	ErrNotOwner            = errors.New("not the owner of this enrollment")
)

type UseCase struct {
	enrollmentRepo domain.EnrollmentRepository
	courseRepo     domain.CourseRepository
}

func NewUseCase(enrollmentRepo domain.EnrollmentRepository, courseRepo domain.CourseRepository) *UseCase {
	return &UseCase{
		enrollmentRepo: enrollmentRepo,
		courseRepo:     courseRepo,
	}
}

func (uc *UseCase) Enroll(userID, courseID uuid.UUID) (*domain.Enrollment, error) {
	// Check if course exists and is published
	course, err := uc.courseRepo.GetByID(courseID)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, ErrCourseNotFound
	}
	if !course.IsPublished {
		return nil, ErrCourseNotPublished
	}

	// Check if already enrolled
	existing, err := uc.enrollmentRepo.GetByUserAndCourse(userID, courseID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyEnrolled
	}

	// Create enrollment
	enrollment := &domain.Enrollment{
		ID:                 uuid.New(),
		UserID:             userID,
		CourseID:           courseID,
		Status:             domain.EnrollmentStatusActive,
		ProgressPercentage: 0,
		VideoWatched:       false,
	}

	if err := uc.enrollmentRepo.Create(enrollment); err != nil {
		return nil, err
	}

	return enrollment, nil
}

func (uc *UseCase) GetByID(id uuid.UUID) (*domain.Enrollment, error) {
	enrollment, err := uc.enrollmentRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if enrollment == nil {
		return nil, ErrEnrollmentNotFound
	}
	return enrollment, nil
}

func (uc *UseCase) GetByUserAndCourse(userID, courseID uuid.UUID) (*domain.Enrollment, error) {
	return uc.enrollmentRepo.GetByUserAndCourse(userID, courseID)
}

func (uc *UseCase) ListByUser(userID uuid.UUID) ([]*domain.EnrollmentWithCourse, error) {
	return uc.enrollmentRepo.ListByUser(userID)
}

func (uc *UseCase) UpdateProgress(id uuid.UUID, userID uuid.UUID, progress int) (*domain.Enrollment, error) {
	enrollment, err := uc.enrollmentRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if enrollment == nil {
		return nil, ErrEnrollmentNotFound
	}
	if enrollment.UserID != userID {
		return nil, ErrNotOwner
	}

	if progress > enrollment.ProgressPercentage {
		enrollment.ProgressPercentage = progress
	}

	if err := uc.enrollmentRepo.Update(enrollment); err != nil {
		return nil, err
	}

	return enrollment, nil
}

func (uc *UseCase) CompleteVideo(id uuid.UUID, userID uuid.UUID) (*domain.Enrollment, error) {
	enrollment, err := uc.enrollmentRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if enrollment == nil {
		return nil, ErrEnrollmentNotFound
	}
	if enrollment.UserID != userID {
		return nil, ErrNotOwner
	}

	enrollment.VideoWatched = true
	enrollment.ProgressPercentage = 100

	if err := uc.enrollmentRepo.Update(enrollment); err != nil {
		return nil, err
	}

	return enrollment, nil
}

func (uc *UseCase) MarkCompleted(id uuid.UUID, userID uuid.UUID) (*domain.Enrollment, error) {
	enrollment, err := uc.enrollmentRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if enrollment == nil {
		return nil, ErrEnrollmentNotFound
	}
	if enrollment.UserID != userID {
		return nil, ErrNotOwner
	}

	now := time.Now()
	enrollment.Status = domain.EnrollmentStatusCompleted
	enrollment.CompletedAt = &now

	if err := uc.enrollmentRepo.Update(enrollment); err != nil {
		return nil, err
	}

	return enrollment, nil
}
