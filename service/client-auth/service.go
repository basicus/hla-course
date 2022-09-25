package client_auth

import (
	"context"
	auth_api "github.com/basicus/hla-course/grpc/auth"
	"github.com/basicus/hla-course/log"
	"google.golang.org/grpc"
)

type Service struct {
	config Config
	conn   *grpc.ClientConn
	Client auth_api.AuthServiceClient
}

func New(config Config) *Service {
	return &Service{config: config}
}

func (s *Service) Run(ctx context.Context) error {
	logger := log.Ctx(ctx)

	logger.WithField("address", s.config.ServiceAuth).Info("Start grpc client")
	defer func() {
		logger.Info("Stop serving requests")
	}()

	conn, err := grpc.Dial(s.config.ServiceAuth, grpc.WithInsecure())
	if err != nil {
		return err
	}
	s.conn = conn
	s.Client = auth_api.NewAuthServiceClient(conn)
	return nil
}

func (s *Service) Shutdown(_ context.Context) error {
	return s.conn.Close()
}
