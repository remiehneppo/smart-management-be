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
	"github.com/redis/go-redis/v9"
	"github.com/remiehneppo/be-task-management/config"
	_ "github.com/remiehneppo/be-task-management/docs"
	"github.com/remiehneppo/be-task-management/internal/database"
	"github.com/remiehneppo/be-task-management/internal/handler"
	"github.com/remiehneppo/be-task-management/internal/logger"
	"github.com/remiehneppo/be-task-management/internal/middleware"
	"github.com/remiehneppo/be-task-management/internal/repository"
	"github.com/remiehneppo/be-task-management/internal/service"
	"github.com/remiehneppo/be-task-management/internal/worker"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/auth"
)

type App struct {
	api         *gin.Engine
	port        string
	database    database.Database
	redisClient *redis.Client
	vectorDb    *weaviate.Client
	logger      *logger.Logger
	worker      *worker.Worker
	config      *config.AppConfig
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

	vectorDbCfg := weaviate.Config{
		Host:   cfg.Weaviate.Host,
		Scheme: cfg.Weaviate.Scheme,
	}
	if cfg.Weaviate.APIKey != "" {
		vectorDbCfg.AuthConfig = auth.ApiKey{
			Value: cfg.Weaviate.APIKey,
		}
		vectorDbCfg.Headers = map[string]string{
			"X-Weaviate-Api-Key":     cfg.Weaviate.APIKey,
			"X-Weaviate-Cluster-Url": fmt.Sprintf("%s://%s", cfg.Weaviate.Scheme, cfg.Weaviate.Host),
		}
	}
	for _, header := range cfg.Weaviate.Header {
		vectorDbCfg.Headers[header.Key] = header.Value
	}
	weaviateClient, err := weaviate.NewClient(vectorDbCfg)
	if err != nil {
		logger.Fatal("error connect to vector database")
	}
	live, err := weaviateClient.Misc().LiveChecker().Do(context.Background())
	if err != nil {
		panic(err)
	}
	logger.Info("Weaviate live status: ", live)
	redisOpts := &redis.Options{
		Addr: cfg.Redis.URL,
	}
	if cfg.Redis.Username != "" && cfg.Redis.Password != "" {
		redisOpts.Username = cfg.Redis.Username
		redisOpts.Password = cfg.Redis.Password
	}
	redisClient := redis.NewClient(redisOpts)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("error connect to redis database")
	}
	logger.Info("redis connected successfully")

	return &App{
		api:         api,
		port:        cfg.Port,
		database:    db,
		logger:      logger,
		config:      cfg,
		vectorDb:    weaviateClient,
		redisClient: redisClient,
		worker:      worker.NewWorker(logger),
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

	// Start worker
	a.worker.Start()
	a.logger.Info("Worker started")

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
	fileMetadataRepo := repository.NewFileMetadataRepository(a.database)
	pendingDocumentRepo := repository.NewPendingDocumentRepository(a.database)
	documentClass := repository.DefaultDocumentClass
	documentClass.Vectorizer = a.config.Weaviate.Text2Vec.Module
	moduleConfig := make(map[string]interface{})
	if a.config.Weaviate.Text2Vec.Model != "" {
		moduleConfig[a.config.Weaviate.Text2Vec.Module] = map[string]interface{}{
			"model":       a.config.Weaviate.Text2Vec.Model,
			"apiEndpoint": a.config.Weaviate.Text2Vec.APIEndpoint,
		}
	}
	documentClass.ModuleConfig = moduleConfig
	documentVectorRepo := repository.NewDocumentVectorRepository(
		context.Background(),
		a.vectorDb,
		documentClass,
		100,
	)

	jwtService := service.NewJWTService(
		a.config.JWT.Secret,
		a.config.JWT.Issuer,
		a.config.JWT.Expire,
	)

	lockService := service.NewLockService(a.redisClient)
	var aiService service.AIService
	if !a.config.UseAI {
		aiService = service.NewNoAIService()
	} else {
		aiService = service.NewOpenAIService(a.config.OpenAI)
	}
	aiAssistantService := service.NewAIAssistantService(aiService)
	loginService := service.NewLoginService(jwtService, userRepo)
	userService := service.NewUserService(userRepo)
	taskService := service.NewTaskService(taskRepo, reportRepo, userRepo)
	fileService := service.NewFileService(
		a.config.FileUpload.UploadDir,
		a.config.FileUpload.MaxSize,
		fileMetadataRepo,
	)
	pdfService := service.NewPDFService(service.DefaultDocumentServiceConfig)
	ragService := service.NewRAGService(
		aiService,
		a.config.RAG.SystemPrompt,
	)
	documentService := service.NewDocumentService(
		aiService,
		ragService,
		fileService,
		pdfService,
		documentVectorRepo,
		pendingDocumentRepo,
		[]string{".pdf"},
		lockService,
	)

	aiAssistantHandler := handler.NewAIAssistantHandler(aiAssistantService)
	loginHandler := handler.NewLoginHandler(loginService, a.logger)
	userHandler := handler.NewUserHandler(userService, a.logger)
	taskHandler := handler.NewTaskHandler(taskService, a.logger)
	documentHandler := handler.NewDocumentHandler(documentService)

	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	a.worker.RegisterIntervalJob(
		60,
		documentService.ProcessDocumentJob(),
	)

	a.api.Use(middleware.CorsMiddleware)
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

	a.api.POST("/api/v1/documents/demo-load-text", documentHandler.DemoloadText)
	documentGroup := a.api.Group("/api/v1/documents")
	documentGroup.Use(authMiddleware.AuthBearerMiddleware())
	documentGroup.POST("/upload", documentHandler.UploadPDF)
	documentGroup.POST("/search", documentHandler.SearchDocument)
	documentGroup.POST("/ask-ai", documentHandler.AskAI)
	documentGroup.POST("/batch-upload", documentHandler.BatchUploadPDFAsync)
	documentGroup.GET("/view", documentHandler.ViewDocument)

	// Middleware

}
