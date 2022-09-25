package rest_chats

import (
	"context"
	"errors"
	auth_api "github.com/basicus/hla-course/grpc/auth"
	"github.com/basicus/hla-course/log"
	"github.com/basicus/hla-course/service/monitoring"
	"github.com/basicus/hla-course/service/rest/middleware"
	"github.com/basicus/hla-course/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/sirupsen/logrus"
	"net/http"
)

type Service struct {
	config  Config
	log     *logrus.Logger
	app     *fiber.App
	storage storage.ChatsService
	authApi auth_api.AuthServiceClient
}

func New(config Config, log *logrus.Logger, prom *monitoring.Service, storage *storage.ChatsService, authApi auth_api.AuthServiceClient) (*Service, error) {

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	app.Use(recover.New())
	app.Use(requestid.New())
	// use monitoring
	if prom != nil && prom.Prom != nil {
		app.Use(prom.Prom.Middleware)
	}

	s := &Service{
		config:  config,
		log:     log,
		app:     app,
		storage: *storage,
		authApi: authApi,
	}
	// Функционал чатов (диалогов)
	app.Use(requestid.New())
	app.Use(middleware.NewLogger(log))
	protected := app.Group("/api/v1/user", s.Protected)
	protected.Get("/chats", s.GetUserChats)        // Получение списка чатов пользователя
	protected.Get("/chat/:id", s.GetChatMessages)  // Получение списка сообщений из чата
	protected.Post("/chat", s.ChatCreate)          // Создать чат с пользователем
	protected.Post("/chat/:id", s.ChatPostMessage) // Отправить сообщение в чат

	return s, nil
}

func (s *Service) Run(ctx context.Context) error {
	logger := log.Ctx(ctx)
	logger.WithField("address", s.config.Listen).Info("Start listening")
	defer func() {
		logger.Info("Stop listening")
	}()

	if err := s.app.Listen(s.config.Listen); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		logger.WithError(err).Error("Failed start listening")
		return err
	}

	return nil

}

func (s *Service) Shutdown(ctx context.Context) error {
	logger := log.Ctx(ctx)

	if err := s.app.Shutdown(); err != nil {
		logger.WithError(err).Error("Failed shutdown")
		return err
	}
	return nil
}
