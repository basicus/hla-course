package service

import "context"

type Service interface {
	Run(ctx context.Context) error
	Shutdown(ctx context.Context) error
}
