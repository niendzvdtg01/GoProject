package main

import (
	routing "backend/api"
	"backend/internal/cache"
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
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

func main() {
	// .env is a developer convenience; in container the env vars come from the
	// orchestrator. Load best-effort and never panic on a missing file.
	_ = godotenv.Load()
	_ = godotenv.Load("Backend/.env")

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

	// Cache wiring. Redis outage degrades to the noop cache so reads fall
	// straight through to the DB — never block startup on the cache.
	cacheCfg := config.NewCacheConfig()
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 2*time.Second)
	cacheStore, cacheErr := cache.NewRedisCache(pingCtx, cacheCfg.Addr)
	pingCancel()
	if cacheErr != nil {
		logger.Warn().Err(cacheErr).Str("addr", cacheCfg.Addr).Msg("redis unavailable; cache disabled")
		cacheStore = cache.NewNoopCache()
	} else {
		logger.Info().Str("addr", cacheCfg.Addr).Msg("redis cache connected")
	}
	teamMembersCache := cache.NewTeamMembers(cacheStore)
	assetCache := cache.NewAssetMetadata(cacheStore)
	aclCache := cache.NewACL(cacheStore)

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
		startConsumers(broker, auditLogRepository, teamMembersCache, assetCache, aclCache, &logger)
	}

	authMiddleware := middleware.NewAuthMiddleware(jwtSecret)
	authService := service.NewAuthService(userRepository, authMiddleware)
	userService := service.NewUserService(userRepository, authMiddleware, importTaskRepository)
	teamService := service.NewTeamManagementService(teamRepository, teamMemberRepository, userRepository).
		WithPublisher(publisher).
		WithCache(teamMembersCache)
	folderService := service.NewFolderService(folderRepository, userRepository, permissionRepository, teamMemberRepository).
		WithPublisher(publisher).
		WithCache(assetCache, aclCache)
	noteService := service.NewNoteService(noteRepository, folderRepository, userRepository, permissionRepository, teamMemberRepository).
		WithPublisher(publisher).
		WithCache(assetCache, aclCache)
	sharingService := service.NewSharing(noteRepository, folderRepository, userRepository, permissionRepository).
		WithPublisher(publisher).
		WithCache(aclCache)

	if err := utils.RegisterValidators(); err != nil {
		panic(err)
	}

	server := routing.SetupRouter(authMiddleware, authService, userService, userRepository, teamService, folderService, noteService, sharingService)
	server.Run(":8080")
}

// startConsumers boots the audit, notification, and cache-invalidator
// consumers. Each consumer runs in its own goroutine inside the broker; this
// function returns immediately. On SIGINT/SIGTERM the shared ctx is canceled
// so all consumer loops exit.
func startConsumers(broker rabbitmq.RabbitMQService, auditRepo *database.AuditLogRepository, teamCache *cache.TeamMembers, assetCache *cache.AssetMetadata, aclCache *cache.ACL, logger *zerolog.Logger) {
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
	consumers = append(consumers, consumer.NewCacheInvalidator(broker, teamCache, assetCache, aclCache, logger)...)
	for _, c := range consumers {
		if err := c.Start(ctx); err != nil {
			logger.Error().Err(err).Str("consumer", c.Name).Msg("failed to start consumer")
		}
	}
}
