package grpc_counter

import (
	"context"
	"errors"
	"fmt"
	counter_api "github.com/basicus/hla-course/grpc/counter"
	"github.com/basicus/hla-course/log"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"math/rand"
	"net"
	"net/http"
	"time"
)

type service struct {
	config Config
	srv    *grpc.Server
	log    *logrus.Logger
	counter_api.UnsafeCounterServiceServer
}

func New(config Config, loggersLogger *logrus.Logger) (*service, error) {
	return &service{
		config: config,
		srv: grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_logrus.UnaryServerInterceptor(loggersLogger.WithField("role", "grpc")),
		))),
		log: loggersLogger,
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

	counter_api.RegisterCounterServiceServer(s.srv, s)

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

// Внимание, в рамках ДЗ методы в реальности не выполняют заявленные функции имитируя поведение и генерируя ошибки.

// NewMessage Метод увеличивающий счетчик непрочитанных сообщений
func (s *service) NewMessage(_ context.Context, request *counter_api.CounterEventRequest) (*counter_api.CounterEventResponse, error) {
	s.log.Infof("Request %+v", request)
	var err error
	err = nil
	if randomBool() {
		err = fmt.Errorf("NewMessage random generated error")
	}
	return &counter_api.CounterEventResponse{}, err
}

// MessageRead Метод уменьшающий счетчик непрочитанных сообщений
func (s *service) MessageRead(_ context.Context, request *counter_api.CounterEventRequest) (*counter_api.CounterEventResponse, error) {
	s.log.Infof("Request %+v", request)
	var err error
	err = nil
	if randomBool() {
		err = fmt.Errorf("NewMessage random generated error")
	}
	return &counter_api.CounterEventResponse{}, err
}

// CompensateNewMessage Компенсирующий метод для отмены отправки нового сообщения
func (s *service) CompensateNewMessage(_ context.Context, request *counter_api.CounterEventRequest) (*counter_api.CounterEventResponse, error) {
	s.log.Infof("Request %+v", request)
	var err error
	err = nil
	return &counter_api.CounterEventResponse{}, err
}

// CompensateMessageRead Компенсирующий метод для уменьшения счетчика при прочтении сообщения
func (s *service) CompensateMessageRead(_ context.Context, request *counter_api.CounterEventRequest) (*counter_api.CounterEventResponse, error) {
	s.log.Infof("Request %+v", request)
	var err error
	err = nil
	return &counter_api.CounterEventResponse{}, err
}

func randomBool() bool {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(2) == 1
}
