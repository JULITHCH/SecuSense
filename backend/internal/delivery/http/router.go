package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/secusense/backend/internal/delivery/http/handler"
	authMiddleware "github.com/secusense/backend/internal/delivery/http/middleware"
)

type Router struct {
	chi            *chi.Mux
	authMiddleware *authMiddleware.AuthMiddleware
	authHandler    *handler.AuthHandler
	courseHandler  *handler.CourseHandler
	enrollHandler  *handler.EnrollmentHandler
	testHandler    *handler.TestHandler
	certHandler    *handler.CertificateHandler
	aiHandler      *handler.AIHandler
}

type RouterConfig struct {
	AllowOrigins []string
}

func NewRouter(
	cfg RouterConfig,
	authMW *authMiddleware.AuthMiddleware,
	authH *handler.AuthHandler,
	courseH *handler.CourseHandler,
	enrollH *handler.EnrollmentHandler,
	testH *handler.TestHandler,
	certH *handler.CertificateHandler,
	aiH *handler.AIHandler,
) *Router {
	r := &Router{
		chi:            chi.NewRouter(),
		authMiddleware: authMW,
		authHandler:    authH,
		courseHandler:  courseH,
		enrollHandler:  enrollH,
		testHandler:    testH,
		certHandler:    certH,
		aiHandler:      aiH,
	}

	// Global middleware
	r.chi.Use(middleware.Logger)
	r.chi.Use(middleware.Recoverer)
	r.chi.Use(middleware.RequestID)
	r.chi.Use(middleware.RealIP)

	// CORS
	r.chi.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.setupRoutes()

	return r
}

func (r *Router) setupRoutes() {
	// Health check
	r.chi.Get("/health", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	})

	r.chi.Route("/api/v1", func(api chi.Router) {
		// Public routes
		api.Route("/auth", func(auth chi.Router) {
			auth.Post("/register", r.authHandler.Register)
			auth.Post("/login", r.authHandler.Login)
			auth.Post("/refresh", r.authHandler.Refresh)
		})

		// Public certificate verification
		api.Get("/certificates/verify/{hash}", r.certHandler.Verify)

		// Public course listing
		api.Get("/courses", r.courseHandler.List)
		api.Get("/courses/{id}", r.courseHandler.GetByID)

		// Synthesia webhook (public but should be secured in production)
		api.Post("/webhooks/synthesia", r.aiHandler.SynthesiaWebhook)

		// Protected routes
		api.Group(func(protected chi.Router) {
			protected.Use(r.authMiddleware.Authenticate)

			// Auth
			protected.Get("/auth/me", r.authHandler.Me)

			// Course enrollment
			protected.Post("/courses/{id}/enroll", r.enrollHandler.Enroll)
			protected.Get("/courses/{courseId}/test", r.testHandler.GetByCourseID)

			// Enrollments
			protected.Get("/enrollments", r.enrollHandler.List)
			protected.Get("/enrollments/{id}", r.enrollHandler.GetByID)
			protected.Put("/enrollments/{id}/progress", r.enrollHandler.UpdateProgress)
			protected.Post("/enrollments/{id}/complete-video", r.enrollHandler.CompleteVideo)

			// Tests
			protected.Post("/tests/{testId}/attempts", r.testHandler.StartAttempt)
			protected.Post("/attempts/{attemptId}/submit", r.testHandler.SubmitAttempt)
			protected.Get("/attempts/{attemptId}/results", r.testHandler.GetAttemptResults)

			// Certificates
			protected.Get("/certificates", r.certHandler.List)
			protected.Get("/certificates/{id}", r.certHandler.GetByID)
			protected.Post("/certificates", r.certHandler.Generate)
			protected.Get("/certificates/{id}/download", r.certHandler.Download)

			// Admin routes
			protected.Group(func(admin chi.Router) {
				admin.Use(r.authMiddleware.RequireAdmin)

				// Course management
				admin.Get("/admin/courses", r.courseHandler.ListAll)
				admin.Post("/admin/courses", r.courseHandler.Create)
				admin.Put("/admin/courses/{id}", r.courseHandler.Update)
				admin.Delete("/admin/courses/{id}", r.courseHandler.Delete)
				admin.Post("/admin/courses/{id}/publish", r.courseHandler.Publish)
				admin.Post("/admin/courses/{id}/unpublish", r.courseHandler.Unpublish)

				// Test management
				admin.Post("/admin/tests", r.testHandler.CreateTest)
				admin.Post("/admin/questions", r.testHandler.CreateQuestion)

				// AI generation
				admin.Post("/admin/generate/course", r.aiHandler.GenerateCourse)
				admin.Get("/admin/generate/jobs/{id}", r.aiHandler.GetJob)

				// Video generation
				admin.Post("/admin/courses/{id}/refresh-video", r.aiHandler.RefreshVideoStatus)
				admin.Post("/admin/videos/poll", r.aiHandler.PollPendingVideos)
			})
		})
	})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.chi.ServeHTTP(w, req)
}
