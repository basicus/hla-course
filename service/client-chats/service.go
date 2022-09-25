package client_chats

import (
	"context"
	chat_api "github.com/basicus/hla-course/grpc/chats"
	"github.com/basicus/hla-course/log"
	"google.golang.org/grpc"
)

type Service struct {
	config Config
	conn   *grpc.ClientConn
	Client chat_api.ChatServiceClient
}

func New(config Config) *Service {
	return &Service{config: config}
}

func (s *Service) Run(ctx context.Context) error {
	logger := log.Ctx(ctx)

	logger.WithField("address", s.config.ServiceChats).Info("Start grpc client")
	defer func() {
		logger.Info("Stop listening")
	}()

	conn, err := grpc.Dial(s.config.ServiceChats, grpc.WithInsecure())
	if err != nil {
		return err
	}
	s.conn = conn

	s.Client = chat_api.NewChatServiceClient(conn)
	return nil
}

func (s *Service) Shutdown(_ context.Context) error {
	return s.conn.Close()
}
