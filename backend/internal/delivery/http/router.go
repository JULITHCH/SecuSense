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
	chi             *chi.Mux
	authMiddleware  *authMiddleware.AuthMiddleware
	authHandler     *handler.AuthHandler
	courseHandler   *handler.CourseHandler
	enrollHandler   *handler.EnrollmentHandler
	testHandler     *handler.TestHandler
	certHandler     *handler.CertificateHandler
	aiHandler       *handler.AIHandler
	workflowHandler *handler.WorkflowHandler
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
	workflowH *handler.WorkflowHandler,
) *Router {
	r := &Router{
		chi:             chi.NewRouter(),
		authMiddleware:  authMW,
		authHandler:     authH,
		courseHandler:   courseH,
		enrollHandler:   enrollH,
		testHandler:     testH,
		certHandler:     certH,
		aiHandler:       aiH,
		workflowHandler: workflowH,
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
		api.Get("/courses/{courseId}/lessons", r.workflowHandler.GetCourseLessons)

		// Synthesia webhook (public but should be secured in production)
		api.Post("/webhooks/synthesia", r.aiHandler.SynthesiaWebhook)

		// Serve generated audio files for presentations (public)
		api.Get("/audio/*", func(w http.ResponseWriter, req *http.Request) {
			// Serve files from the generated/audio directory
			fs := http.StripPrefix("/api/v1/audio/", http.FileServer(http.Dir("./generated/audio")))
			fs.ServeHTTP(w, req)
		})

		// Protected routes
		api.Group(func(protected chi.Router) {
			protected.Use(r.authMiddleware.Authenticate)

			// Auth
			protected.Get("/auth/me", r.authHandler.Me)

			// Course enrollment
			protected.Post("/courses/{id}/enroll", r.enrollHandler.Enroll)
			protected.Get("/courses/{courseId}/test", r.testHandler.GetByCourseID)
			protected.Get("/courses/{id}/video-url", r.aiHandler.GetVideoURL)

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

				// Question management for existing courses
				admin.Get("/admin/courses/{courseId}/questions", r.testHandler.GetQuestionsByCourseID)
				admin.Put("/admin/questions/{questionId}", r.testHandler.UpdateQuestion)
				admin.Delete("/admin/questions/{questionId}", r.testHandler.DeleteQuestion)

				// AI generation
				admin.Post("/admin/generate/course", r.aiHandler.GenerateCourse)
				admin.Get("/admin/generate/jobs/{id}", r.aiHandler.GetJob)

				// Video generation
				admin.Post("/admin/courses/{id}/refresh-video", r.aiHandler.RefreshVideoStatus)
				admin.Post("/admin/videos/poll", r.aiHandler.PollPendingVideos)

				// Course Workflow (multi-agency)
				admin.Post("/admin/workflow/start", r.workflowHandler.StartResearch)
				admin.Get("/admin/workflow/{id}", r.workflowHandler.GetSession)
				admin.Put("/admin/workflow/{sessionId}/suggestions/{suggestionId}", r.workflowHandler.UpdateSuggestionStatus)
				admin.Post("/admin/workflow/{sessionId}/suggestions", r.workflowHandler.AddCustomTopic)
				admin.Post("/admin/workflow/{sessionId}/generate-more", r.workflowHandler.GenerateMoreSuggestions)
				admin.Post("/admin/workflow/{sessionId}/refine", r.workflowHandler.ProceedToRefinement)
				admin.Post("/admin/workflow/{sessionId}/scripts", r.workflowHandler.ProceedToScriptGeneration)
				admin.Post("/admin/workflow/{sessionId}/videos", r.workflowHandler.ProceedToVideoGeneration)

				// Refined topic management
				admin.Put("/admin/workflow/{sessionId}/topics/{topicId}", r.workflowHandler.UpdateRefinedTopic)
				admin.Post("/admin/workflow/{sessionId}/topics/{topicId}/regenerate", r.workflowHandler.RegenerateTopic)
				admin.Put("/admin/workflow/{sessionId}/topics/reorder", r.workflowHandler.ReorderRefinedTopics)

				// Lesson script management
				admin.Put("/admin/workflow/{sessionId}/lessons/{lessonId}", r.workflowHandler.UpdateLessonScript)
				admin.Post("/admin/workflow/{sessionId}/lessons/{lessonId}/regenerate", r.workflowHandler.RegenerateScript)

				// Presentation/Output type management
				admin.Put("/admin/workflow/{sessionId}/lessons/{lessonId}/output-type", r.workflowHandler.SetOutputType)
				admin.Post("/admin/workflow/{sessionId}/lessons/{lessonId}/presentation", r.workflowHandler.GeneratePresentation)
				admin.Get("/admin/workflow/{sessionId}/lessons/{lessonId}/presentation", r.workflowHandler.GetPresentation)
				admin.Post("/admin/workflow/{sessionId}/lessons/{lessonId}/regenerate-audio", r.workflowHandler.RegenerateAudio)

				// Question generation
				admin.Post("/admin/workflow/{sessionId}/questions", r.workflowHandler.ProceedToQuestionGeneration)
				admin.Get("/admin/workflow/{sessionId}/questions/preview", r.workflowHandler.PreviewQuestions)

				// Lesson management for completed courses
				admin.Put("/admin/courses/{courseId}/lessons/{lessonId}", r.workflowHandler.UpdateLessonScriptForCourse)
				admin.Post("/admin/courses/{courseId}/lessons/{lessonId}/regenerate", r.workflowHandler.RegenerateLessonScriptForCourse)
				admin.Post("/admin/courses/{courseId}/lessons/{lessonId}/regenerate-presentation", r.workflowHandler.RegeneratePresentationForCourse)
			})
		})
	})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.chi.ServeHTTP(w, req)
}
