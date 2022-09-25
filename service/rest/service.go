package rest

import (
	"context"
	"errors"
	chat_api "github.com/basicus/hla-course/grpc/chats"
	"github.com/basicus/hla-course/log"
	"github.com/basicus/hla-course/service/monitoring"
	"github.com/basicus/hla-course/service/queue"
	"github.com/basicus/hla-course/service/rest/handlers"
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
	storage *storage.UserService
	auth    *storage.UserService
}

func New(config Config, log *logrus.Logger, prom *monitoring.Service, storage *storage.UserService, queue *queue.Service, auth *storage.UserService, chatApi chat_api.ChatServiceClient) (*Service, error) {

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	app.Use(requestid.New())
	app.Use(recover.New())
	// use monitoring
	if prom != nil && prom.Prom != nil {
		app.Use(prom.Prom.Middleware)
	}

	h := handlers.Handlers{
		Logger:      log,
		Storage:     *storage,
		AuthService: nil,
		Config:      handlers.Config{JwtSecret: config.JwtSecret, PostsLimit: config.PostsLimit},
		Queue:       queue,
		ChatApi:     chatApi,
	}
	if auth != nil {
		h.AuthService = *auth
	}

	app.Use(middleware.NewLogger(log))
	app.Post("/api/v1/register", h.Register)
	app.Post("/api/v1/login", h.Login)
	app.Get("/api/v1/users", h.UsersGet)
	app.Get("/api/v1/search", h.SearchProfile)
	app.Get("/api/v1/post/:id", h.GetPostById)
	protected := app.Group("/api/v1/user", middleware.Protected(config.JwtSecret))
	protected.Get("/feed", h.PersonalFeed)
	protected.Get("/chats", h.GetUserChats) // Получение списка чатов пользователя
	protected.Post("", h.UpdateProfile)
	protected.Get("/friends", h.GetFriends)
	protected.Get("/:id", h.UserInfo)
	protected.Get("", h.UserInfo)
	protected.Post("/:id/friend", h.AddFriend)
	protected.Delete("/:id/friend", h.DeleteFriend)
	protected.Post("/publish", h.PublishPost)

	// Функционал чатов (диалогов)

	protected.Get("/chat/:id", h.GetChatMessages)  // Получение списка сообщений из чата
	protected.Post("/chat", h.ChatCreate)          // Создать чат с пользователем
	protected.Post("/chat/:id", h.ChatPostMessage) // Отправить сообщение в чат

	//app.Post("/api/v1/logout", handlers.Logout)
	//app.Post("/api/v1/password_recover", handlers.PasswordRecover)

	return &Service{
		config:  config,
		log:     log,
		app:     app,
		storage: storage,
	}, nil
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
