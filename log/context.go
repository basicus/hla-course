package log

import (
	"context"
	"github.com/sirupsen/logrus"
)

type logKey struct{}

var key logKey

func WithContext(ctx context.Context, log *logrus.Entry) context.Context {
	return context.WithValue(ctx, key, log)
}

func Ctx(ctx context.Context) *logrus.Entry {
	if entity, ok := ctx.Value(key).(*logrus.Entry); ok {
		return entity
	}
	return logrus.NewEntry(logrus.StandardLogger())
}
