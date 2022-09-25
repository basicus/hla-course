package grpc_chats

import (
	"context"
	"errors"
	auth_api "github.com/basicus/hla-course/grpc/auth"
	chatsapi "github.com/basicus/hla-course/grpc/chats"
	"github.com/basicus/hla-course/log"
	client_auth "github.com/basicus/hla-course/service/client-auth"
	"github.com/basicus/hla-course/storage"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net"
	"net/http"
)

type service struct {
	config  Config
	log     *logrus.Logger
	storage storage.ChatsService
	auth    *client_auth.Service
	srv     *grpc.Server
	chatsapi.UnsafeChatServiceServer
}

func New(config Config, storage *storage.ChatsService, auth *client_auth.Service, loggersLogger *logrus.Logger) (*service, error) {
	return &service{
		config:  config,
		storage: *storage,
		srv: grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_logrus.UnaryServerInterceptor(loggersLogger.WithField("role", "grpc")),
		))),
		log:  loggersLogger,
		auth: auth,
	}, nil
}

func (s *service) Run(ctx context.Context) error {
	logger := log.Ctx(ctx)

	logger.WithField("address", s.config.Address).Info("Start listening")
	defer func() {
		logger.Info("Stop listening")
	}()

	lis, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		logger.WithError(err).Error("failed to create grpc listener")
		return err
	}

	chatsapi.RegisterChatServiceServer(s.srv, s)

	if err := s.srv.Serve(lis); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		logger.WithError(err).Error("failed to start listening")
		return err
	}

	return nil
}

func (s *service) Shutdown(_ context.Context) error {
	s.srv.GracefulStop()
	return nil
}

func (s *service) Get(ctx context.Context, request *chatsapi.GetChatRequest) (*chatsapi.GetChatResponse, error) {
	s.log.Infof("Request Get Chat Info request_id %s", request.RequestId)

	chat, err := s.storage.GetChat(ctx, request.GetChatId())
	if err != nil {
		return nil, err
	}

	users, err := s.storage.ChatGetParticipants(ctx, request.GetChatId())
	if err != nil {
		return nil, err
	}

	return &chatsapi.GetChatResponse{
		Chat: &chatsapi.ChatInfo{
			ChatId:    chat.Id,
			Title:     chat.Title,
			CreatedAt: timestamppb.New(chat.CreatedAt),
			Closed:    chat.Closed,
		},
		Users: users,
	}, nil
}

func (s *service) ListChats(ctx context.Context, request *chatsapi.ListUserChatsRequest) (*chatsapi.ListUserChatsResponse, error) {
	s.log.Infof("Request List Chats request_id %s", request.RequestId)

	chats, err := s.storage.UserGetChats(ctx, request.GetUserId())
	if err != nil {
		return nil, err
	}

	// Convert to grpc response
	chatResponse := make([]*chatsapi.ChatInfo, len(chats))
	for i, chat := range chats {
		chatResponse[i] = &chatsapi.ChatInfo{
			ChatId:    chat.Id,
			Title:     chat.Title,
			CreatedAt: timestamppb.New(chat.CreatedAt),
			Closed:    chat.Closed,
		}
	}
	return &chatsapi.ListUserChatsResponse{Chats: chatResponse}, nil
}

func (s *service) Messages(ctx context.Context, request *chatsapi.ChatMessagesRequest) (*chatsapi.ChatMessagesResponse, error) {
	s.log.Infof("Request Chat Messages request_id %s", request.RequestId)

	chat, err := s.storage.GetChat(ctx, request.GetChatId())
	if err != nil {
		return nil, err
	}

	messages, err := s.storage.ChatMessages(ctx, request.GetChatId(), 0, 0)
	if err != nil {
		return nil, err
	}

	userNames := make(map[int64]string)

	// Convert to grpc response
	messageResponse := make([]*chatsapi.ChatMessage, len(messages))
	for i, msg := range messages {
		// Check if we already have username in map
		_, ok := userNames[msg.UserFrom]
		if !ok {
			usrName, err := s.auth.Client.UserName(ctx, &auth_api.UserNameRequest{UserId: msg.UserFrom})
			if err != nil {
				return nil, err
			}
			userNames[msg.UserFrom] = usrName.UserName
		}
		messageResponse[i] = &chatsapi.ChatMessage{
			MessageId: msg.Id,
			UserFrom:  userNames[msg.UserFrom],
			Date:      timestamppb.New(msg.SendAt),
			Message:   msg.Message,
		}
	}

	return &chatsapi.ChatMessagesResponse{
		Chat: &chatsapi.ChatInfo{
			ChatId:    chat.Id,
			Title:     chat.Title,
			CreatedAt: timestamppb.New(chat.CreatedAt),
			Closed:    chat.Closed,
		},
		Messages: messageResponse,
	}, nil
}

func (s *service) CreateChat(ctx context.Context, request *chatsapi.CreateChatRequest) (*chatsapi.CreateChatResponse, error) {
	s.log.Infof("Request CreateChat request_id %s", request.RequestId)
	chat, err := s.storage.ChatCreate(ctx, request.Title, request.GetUsers()...)
	if err != nil {
		return nil, err
	}

	return &chatsapi.CreateChatResponse{Chat: &chatsapi.ChatInfo{
		ChatId:    chat.Id,
		Title:     chat.Title,
		CreatedAt: timestamppb.New(chat.CreatedAt),
		Closed:    chat.Closed,
	}}, nil
}

func (s *service) PostMessage(ctx context.Context, request *chatsapi.PostMessageRequest) (*chatsapi.PostMessageResponse, error) {
	s.log.Infof("Request PostMessage request_id %s", request.RequestId)
	message, err := s.storage.MessageSave(ctx, request.ChatId, request.GetUserId(), request.Date.AsTime(), request.GetMessage())
	if err != nil {
		return nil, err
	}
	userfrom, err := s.auth.Client.UserName(ctx, &auth_api.UserNameRequest{UserId: message.UserFrom})
	if err != nil {
		return nil, err
	}

	return &chatsapi.PostMessageResponse{Message: &chatsapi.ChatMessage{
		MessageId: message.Id,
		UserFrom:  userfrom.UserName,
		Date:      timestamppb.New(message.SendAt),
		Message:   message.Message,
	}}, nil
}
