package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/remiehneppo/be-task-management/config"
	_ "github.com/remiehneppo/be-task-management/docs"
	"github.com/remiehneppo/be-task-management/internal/database"
	"github.com/remiehneppo/be-task-management/internal/handler"
	"github.com/remiehneppo/be-task-management/internal/logger"
	"github.com/remiehneppo/be-task-management/internal/middleware"
	"github.com/remiehneppo/be-task-management/internal/repository"
	"github.com/remiehneppo/be-task-management/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type App struct {
	api      *gin.Engine
	port     string
	database database.Database
	logger   *logger.Logger
	config   *config.AppConfig
}

func NewApp(cfg *config.AppConfig) *App {

	logger, err := logger.NewLogger(&cfg.Logger)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	api := gin.New()
	api.Use(gin.Recovery())
	api.Use(logger.GinLogger())

	// Initialize database
	db := database.NewMongoDatabase(cfg.MongoDB.URI, cfg.MongoDB.Database)
	// Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("Connecting to database...")
	if err := db.Connect(ctx); err != nil {
		logger.Fatal("error connect to database")
	}
	logger.Info("Database connected successfully")

	return &App{
		api:      api,
		port:     cfg.Port,
		database: db,
		logger:   logger,
		config:   cfg,
	}
}

func (a *App) Start() error {
	// Initialize Gin

	// Create server
	srv := &http.Server{
		Addr:    ":" + a.port,
		Handler: a.api,
	}

	// Channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	// Start server
	go func() {
		a.logger.Info("Server starting on port ", a.port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Channel for listening to OS signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking select waiting for server errors or shutdown signals
	select {
	case err := <-serverErrors:
		a.logger.Error("Server error: ", err)
		return err

	case <-shutdown:
		a.logger.Info("Starting graceful shutdown...")

		// Create context with timeout for shutdown operations
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Shutdown the server
		if err := srv.Shutdown(ctx); err != nil {
			a.logger.Error("Server shutdown error: ", err)

			// Force shutdown if graceful shutdown fails
			if err := srv.Close(); err != nil {
				a.logger.Error("Server forced close error: ", err)
				return err
			}
		}

		// Disconnect from database
		a.logger.Info("Disconnecting from database...")
		if err := a.database.Disconnect(ctx); err != nil {
			a.logger.Error("Database disconnect error: ", err)
			return err
		}

		a.logger.Info("Graceful shutdown completed")
	}

	return nil
}

func (a *App) RegisterHandler() {
	userRepo := repository.NewUserRepository(a.database)
	taskRepo := repository.NewTaskRepository(a.database)
	reportRepo := repository.NewReportRepository(a.database)

	jwtService := service.NewJWTService(
		a.config.JWT.Secret,
		a.config.JWT.Issuer,
		a.config.JWT.Expire,
	)

	aiService := service.NewOpenAIService(a.config.OpenAI)
	aiAssistantService := service.NewAIAssistantService(aiService)
	loginService := service.NewLoginService(jwtService, userRepo)
	userService := service.NewUserService(userRepo)
	taskService := service.NewTaskService(taskRepo, reportRepo, userRepo)

	aiAssistantHandler := handler.NewAIAssistantHandler(aiAssistantService)
	loginHandler := handler.NewLoginHandler(loginService, a.logger)
	userHandler := handler.NewUserHandler(userService, a.logger)
	taskHandler := handler.NewTaskHandler(taskService, a.logger)

	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Register routes

	a.api.Handle("GET", "/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	a.api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	a.api.POST("/api/v1/auth/login", loginHandler.Login)
	a.api.POST("/api/v1/auth/logout", authMiddleware.AuthBearerMiddleware(), loginHandler.Logout)
	a.api.POST("/api/v1/auth/refresh", authMiddleware.AuthBearerMiddleware(), loginHandler.Refresh)

	userGroup := a.api.Group("/api/v1/users")
	userGroup.Use(authMiddleware.AuthBearerMiddleware())
	userGroup.GET("/me", userHandler.GetUserInfo)
	userGroup.POST("/password", userHandler.UpdatePassword)
	userGroup.GET("/workspace", userHandler.GetUsersSameWorkspace)

	taskGroup := a.api.Group("/api/v1/tasks")
	taskGroup.Use(authMiddleware.AuthBearerMiddleware())

	taskGroup.GET("/assigned", taskHandler.GetTasksAssignedToUser)
	taskGroup.GET("/created", taskHandler.GetTasksCreatedByUser)
	taskGroup.GET("/{id}", taskHandler.GetTaskByID)
	taskGroup.POST("/create", taskHandler.CreateTask)
	taskGroup.POST("/update", taskHandler.UpdateTask)
	taskGroup.POST("/delete/{id}", taskHandler.DeleteTask)
	taskGroup.GET("/filter", taskHandler.FilterTasks)
	taskGroup.POST("/report/add", taskHandler.AddReportTask)
	taskGroup.POST("/report/delete", taskHandler.DeleteReport)
	taskGroup.POST("/report/update", taskHandler.UpdateReportTask)
	taskGroup.POST("/report/feedback", taskHandler.FeedbackReport)

	aiAssistantGroup := a.api.Group("/api/v1/assistant")
	aiAssistantGroup.Use(authMiddleware.AuthBearerMiddleware())
	aiAssistantGroup.POST("/chat", aiAssistantHandler.ChatWithAssistant)
	aiAssistantGroup.POST("/chat-stateless", aiAssistantHandler.ChatWithAssistantStateless)
}
