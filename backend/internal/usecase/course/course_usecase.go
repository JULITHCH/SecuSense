package course

import (
	"errors"

	"github.com/google/uuid"
	"github.com/secusense/backend/internal/domain"
)

var (
	ErrCourseNotFound = errors.New("course not found")
)

type UseCase struct {
	courseRepo        domain.CourseRepository
	courseContentRepo domain.CourseContentRepository
}

func NewUseCase(courseRepo domain.CourseRepository, courseContentRepo domain.CourseContentRepository) *UseCase {
	return &UseCase{
		courseRepo:        courseRepo,
		courseContentRepo: courseContentRepo,
	}
}

func (uc *UseCase) Create(req *domain.CreateCourseRequest) (*domain.Course, error) {
	course := &domain.Course{
		ID:             uuid.New(),
		Title:          req.Title,
		Description:    req.Description,
		PassPercentage: req.PassPercentage,
		IsPublished:    false,
	}

	if err := uc.courseRepo.Create(course); err != nil {
		return nil, err
	}

	return course, nil
}

func (uc *UseCase) GetByID(id uuid.UUID) (*domain.Course, error) {
	course, err := uc.courseRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, ErrCourseNotFound
	}
	return course, nil
}

func (uc *UseCase) Update(id uuid.UUID, req *domain.UpdateCourseRequest) (*domain.Course, error) {
	course, err := uc.courseRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, ErrCourseNotFound
	}

	if req.Title != nil {
		course.Title = *req.Title
	}
	if req.Description != nil {
		course.Description = *req.Description
	}
	if req.VideoURL != nil {
		course.VideoURL = req.VideoURL
	}
	if req.ThumbnailURL != nil {
		course.ThumbnailURL = req.ThumbnailURL
	}
	if req.PassPercentage != nil {
		course.PassPercentage = *req.PassPercentage
	}
	if req.IsPublished != nil {
		course.IsPublished = *req.IsPublished
	}

	if err := uc.courseRepo.Update(course); err != nil {
		return nil, err
	}

	return course, nil
}

func (uc *UseCase) Delete(id uuid.UUID) error {
	course, err := uc.courseRepo.GetByID(id)
	if err != nil {
		return err
	}
	if course == nil {
		return ErrCourseNotFound
	}

	return uc.courseRepo.Delete(id)
}

func (uc *UseCase) List(page, pageSize int, publishedOnly bool) ([]*domain.Course, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	courses, err := uc.courseRepo.List(pageSize, offset, publishedOnly)
	if err != nil {
		return nil, 0, err
	}

	total, err := uc.courseRepo.Count(publishedOnly)
	if err != nil {
		return nil, 0, err
	}

	return courses, total, nil
}

func (uc *UseCase) GetContent(courseID uuid.UUID) (*domain.CourseContent, error) {
	return uc.courseContentRepo.GetByCourseID(courseID)
}

func (uc *UseCase) SaveContent(content *domain.CourseContent) error {
	existing, err := uc.courseContentRepo.GetByCourseID(content.CourseID)
	if err != nil {
		return err
	}

	if existing != nil {
		content.ID = existing.ID
		return uc.courseContentRepo.Update(content)
	}

	return uc.courseContentRepo.Create(content)
}

func (uc *UseCase) Publish(id uuid.UUID) error {
	course, err := uc.courseRepo.GetByID(id)
	if err != nil {
		return err
	}
	if course == nil {
		return ErrCourseNotFound
	}

	course.IsPublished = true
	return uc.courseRepo.Update(course)
}

func (uc *UseCase) Unpublish(id uuid.UUID) error {
	course, err := uc.courseRepo.GetByID(id)
	if err != nil {
		return err
	}
	if course == nil {
		return ErrCourseNotFound
	}

	course.IsPublished = false
	return uc.courseRepo.Update(course)
}
