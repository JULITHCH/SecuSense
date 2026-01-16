package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/secusense/backend/config"
	httpDelivery "github.com/secusense/backend/internal/delivery/http"
	"github.com/secusense/backend/internal/delivery/http/handler"
	"github.com/secusense/backend/internal/delivery/http/middleware"
	"github.com/secusense/backend/internal/repository/postgres"
	"github.com/secusense/backend/internal/usecase/ai"
	"github.com/secusense/backend/internal/usecase/auth"
	"github.com/secusense/backend/internal/usecase/certificate"
	"github.com/secusense/backend/internal/usecase/course"
	"github.com/secusense/backend/internal/usecase/enrollment"
	"github.com/secusense/backend/internal/usecase/test"
	"github.com/secusense/backend/infrastructure/database"
	"github.com/secusense/backend/infrastructure/ollama"
	"github.com/secusense/backend/infrastructure/synthesia"
	"github.com/secusense/backend/pkg/jwt"
	"github.com/secusense/backend/pkg/pdf"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.NewPostgresConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db)
	courseRepo := postgres.NewCourseRepository(db)
	courseContentRepo := postgres.NewCourseContentRepository(db)
	enrollmentRepo := postgres.NewEnrollmentRepository(db)
	testRepo := postgres.NewTestRepository(db)
	questionRepo := postgres.NewQuestionRepository(db)
	attemptRepo := postgres.NewTestAttemptRepository(db)
	answerRepo := postgres.NewUserAnswerRepository(db)
	certRepo := postgres.NewCertificateRepository(db)
	aiJobRepo := postgres.NewAIGenerationJobRepository(db)

	// Initialize JWT manager
	jwtManager := jwt.NewManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessExpiresIn,
		cfg.JWT.RefreshExpiresIn,
		cfg.JWT.Issuer,
	)

	// Initialize external clients
	ollamaClient := ollama.NewClient(cfg.Ollama)
	synthesiaClient := synthesia.NewClient(cfg.Synthesia)

	// Initialize PDF generator
	pdfGen := pdf.NewCertificateGenerator("https://secusense.example.com")

	// Initialize use cases
	authUC := auth.NewUseCase(userRepo, refreshTokenRepo, jwtManager)
	courseUC := course.NewUseCase(courseRepo, courseContentRepo)
	enrollmentUC := enrollment.NewUseCase(enrollmentRepo, courseRepo)
	testUC := test.NewUseCase(testRepo, questionRepo, attemptRepo, answerRepo, enrollmentRepo, courseRepo)
	certUC := certificate.NewUseCase(certRepo, attemptRepo, pdfGen)
	aiUC := ai.NewUseCase(aiJobRepo, courseRepo, courseContentRepo, testRepo, questionRepo, ollamaClient, synthesiaClient)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUC)
	courseHandler := handler.NewCourseHandler(courseUC)
	enrollmentHandler := handler.NewEnrollmentHandler(enrollmentUC)
	testHandler := handler.NewTestHandler(testUC)
	certHandler := handler.NewCertificateHandler(certUC)
	aiHandler := handler.NewAIHandler(aiUC)

	// Initialize router
	router := httpDelivery.NewRouter(
		httpDelivery.RouterConfig{
			AllowOrigins: cfg.Server.AllowOrigins,
		},
		authMiddleware,
		authHandler,
		courseHandler,
		enrollmentHandler,
		testHandler,
		certHandler,
		aiHandler,
	)

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
