package main

import (
	routing "backend/api"
	"backend/internal/config"
	"backend/internal/consumer"
	"backend/internal/middleware"
	"backend/internal/rabbitmq"
	database "backend/internal/repository"
	"backend/internal/service"
	"backend/package/event"
	"backend/package/utils"
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

func main() {
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("Backend/.env"); err != nil {
			panic("Error loading .env file")
		}
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	dbConfig := config.NewDBConfig()
	if err := database.ConnectDB(dbConfig.GetDSN()); err != nil {
		panic(err)
	}
	defer database.CloseDB()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		panic("JWT_SECRET is required")
	}

	userRepository := database.NewUserRepository(database.DB)
	teamRepository := database.NewTeamRepository(database.DB)
	teamMemberRepository := database.NewTeamMemberRepository(database.DB)
	folderRepository := database.NewFolderRepository(database.DB)
	noteRepository := database.NewNoteRepository(database.DB)
	permissionRepository := database.NewPermissionRepository(database.DB)
	importTaskRepository := database.NewImportTaskRepository(database.DB)
	auditLogRepository := database.NewAuditLogRepository(database.DB)

	// RabbitMQ wiring. Broker outage must not block API startup: degrade to the
	// noop publisher so the rest of the app keeps working until the broker is back.
	rmqCfg := config.NewRabbitMQConfig()
	broker, brokerErr := rabbitmq.NewRabbitMQService(rmqCfg.URL, &logger)
	var publisher event.Publisher = event.NewNoopPublisher()
	if brokerErr != nil {
		logger.Warn().Err(brokerErr).Msg("rabbitmq disabled; events will be dropped")
	} else {
		defer broker.Close()
		p, err := event.NewPublisher(broker, &logger)
		if err != nil {
			logger.Warn().Err(err).Msg("event publisher init failed; events will be dropped")
		} else {
			publisher = p
		}
		startConsumers(broker, auditLogRepository, &logger)
	}

	authMiddleware := middleware.NewAuthMiddleware(jwtSecret)
	authService := service.NewAuthService(userRepository, authMiddleware)
	userService := service.NewUserService(userRepository, authMiddleware, importTaskRepository)
	teamService := service.NewTeamManagementService(teamRepository, teamMemberRepository, userRepository).WithPublisher(publisher)
	folderService := service.NewFolderService(folderRepository, userRepository, permissionRepository, teamMemberRepository).WithPublisher(publisher)
	noteService := service.NewNoteService(noteRepository, folderRepository, userRepository, permissionRepository, teamMemberRepository).WithPublisher(publisher)
	sharingService := service.NewSharing(noteRepository, folderRepository, userRepository, permissionRepository).WithPublisher(publisher)

	if err := utils.RegisterValidators(); err != nil {
		panic(err)
	}

	server := routing.SetupRouter(authMiddleware, authService, userService, userRepository, teamService, folderService, noteService, sharingService)
	server.Run(":8080")
}

// startConsumers boots the audit + notification consumers. Each consumer runs
// in its own goroutine inside the broker; this function returns immediately.
// On SIGINT/SIGTERM the shared ctx is canceled so all consumer loops exit.
func startConsumers(broker rabbitmq.RabbitMQService, auditRepo *database.AuditLogRepository, logger *zerolog.Logger) {
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		logger.Info().Msg("shutting down consumers")
		cancel()
	}()

	consumers := append(
		consumer.NewAuditConsumer(broker, auditRepo, logger),
		consumer.NewNotificationConsumer(broker, logger)...,
	)
	for _, c := range consumers {
		if err := c.Start(ctx); err != nil {
			logger.Error().Err(err).Str("consumer", c.Name).Msg("failed to start consumer")
		}
	}
}
