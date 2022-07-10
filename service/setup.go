package service

import (
	"context"
	"github.com/basicus/hla-course/log"
	"github.com/oklog/run"
)

func Setup(ctx context.Context, srv Service, name string, g *run.Group) error {
	logger := log.Ctx(ctx).WithField("service", name)
	ctx = log.WithContext(ctx, logger)

	g.Add(func() error {
		logger.Info("Running the service...")
		defer logger.Info("the service is stopped")
		if err := ctx.Err(); err != nil {
			return nil
		}
		return srv.Run(ctx)
	}, func(error) {
		logger.Info("Shutdown the service...")
		defer logger.Info("the service is shutdown")
		cCtx := log.WithContext(context.Background(), logger)
		if err := srv.Shutdown(cCtx); err != nil {
			logger.WithError(err).Error("Cannot shutdown the service properly")
		}
	})

	return nil
}
