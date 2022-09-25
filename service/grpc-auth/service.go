package grpc_auth

import (
	"context"
	"errors"
	authapi "github.com/basicus/hla-course/grpc/auth"
	"github.com/basicus/hla-course/log"
	"github.com/basicus/hla-course/storage"
	"github.com/golang-jwt/jwt/v4"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"net/http"
)

type service struct {
	config  Config
	log     *logrus.Logger
	storage storage.UserService
	srv     *grpc.Server
	authapi.UnsafeAuthServiceServer
	key string
}

func (s *service) CheckSession(_ context.Context, request *authapi.CheckSessionRequest) (*authapi.CheckSessionResponse, error) {
	s.log.Info("Request Check session Info")
	token, err := jwt.Parse(request.JwtToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.key), nil
	})
	if err != nil {
		return nil, err
	}
	if token.Valid {
		claims := token.Claims.(jwt.MapClaims)
		userId := int64(claims["user_id"].(float64))

		return &authapi.CheckSessionResponse{
			Ok:       true,
			UserId:   userId,
			JwtToken: request.GetJwtToken(),
		}, nil
	} else {
		return &authapi.CheckSessionResponse{
			Ok: false}, nil
	}

}

func (s *service) UserName(ctx context.Context, request *authapi.UserNameRequest) (*authapi.UserNameResponse, error) {
	s.log.Info("Request Get username")

	username, err := s.storage.GetUserName(ctx, request.GetUserId())
	if err != nil {
		return nil, err
	}
	return &authapi.UserNameResponse{UserName: username}, nil
}

func (s *service) UserShard(ctx context.Context, request *authapi.UserShardRequest) (*authapi.UserShardResponse, error) {
	user, err := s.storage.GetById(ctx, request.GetUserId())
	if err != nil {
		return nil, err
	}
	return &authapi.UserShardResponse{ShardId: user.ShardId}, nil
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

	authapi.RegisterAuthServiceServer(s.srv, s)

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

func New(config Config, storage *storage.UserService, loggersLogger *logrus.Logger) (*service, error) {
	return &service{
		config:  config,
		storage: *storage,
		srv: grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_logrus.UnaryServerInterceptor(loggersLogger.WithField("role", "grpc")),
		))),
		log: loggersLogger,
		key: config.JwtSecret,
	}, nil
}
