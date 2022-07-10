package monitoring

import (
	"context"
	"errors"
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/basicus/hla-course/log"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"net/http"
)

const (
	ServiceName = "social-network"
)

type Service struct {
	config     Config
	log        *logrus.Logger
	monitoring *fiber.App
	Prom       *fiberprometheus.FiberPrometheus
}

func New(config Config, log *logrus.Logger) (*Service, error) {

	monitoring := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	prometheus := fiberprometheus.New(ServiceName)
	prometheus.RegisterAt(monitoring, "/metrics")

	return &Service{
		config:     config,
		log:        log,
		monitoring: monitoring,
		Prom:       prometheus,
	}, nil
}

func (s *Service) Run(ctx context.Context) error {
	logger := log.Ctx(ctx)
	logger.WithField("address", s.config.Listen).Info("Start listening")
	defer func() {
		logger.Info("Stop listening")
	}()

	if err := s.monitoring.Listen(s.config.Listen); err != nil {
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

	if err := s.monitoring.Shutdown(); err != nil {
		logger.WithError(err).Error("Failed shutdown")
		return err
	}
	return nil
}
