package client_counter

import (
	"context"
	counter_api "github.com/basicus/hla-course/grpc/counter"
	"github.com/basicus/hla-course/log"
	"google.golang.org/grpc"
)

type Service struct {
	config Config
	conn   *grpc.ClientConn
	Client counter_api.CounterServiceClient
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

	s.Client = counter_api.NewCounterServiceClient(conn)
	return nil
}

func (s *Service) Shutdown(_ context.Context) error {
	return s.conn.Close()
}
