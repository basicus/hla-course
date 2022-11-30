package rest_chats

import (
	"context"
	"fmt"
	auth_api "github.com/basicus/hla-course/grpc/auth"
	counter_api "github.com/basicus/hla-course/grpc/counter"
	"github.com/basicus/hla-course/model"
	"github.com/google/uuid"
	"github.com/itimofeev/go-saga"
	"math/rand"
	"time"
)

// Шаг отправки сообщения. Функция
func (s *Service) saveMessage(ctx context.Context) (*sagaNewMessageData, error) {
	data := ctx.Value("data").(*sagaNewMessageData)
	d := *data
	s.log.Infof("saveMessage data %+v", d)
	if randomBool() {
		s.log.Infof("generating error on save message")
		return data, fmt.Errorf("generated error")
	}
	// Try to save message to chat
	messageSaved, err := s.storage.MessageSave(ctx, d.chatId, d.userFromId, d.date, d.message)
	if err != nil {
		return data, err
	}

	// Get username
	userName, err := s.authApi.UserName(ctx, &auth_api.UserNameRequest{UserId: messageSaved.UserFrom})
	if err != nil {
		return data, err
	}
	data.savedMessage = &messageSaved
	data.userName = &userName.UserName
	return data, nil
}

// Шаг сохранения. Компенсирующая функция. Выполняется если в цепочке генерируется ошибка.
func (s *Service) compSaveMessage(ctx context.Context, data *sagaNewMessageData) error {
	d := *data
	s.log.Infof("compSaveMessage data %+v", d)
	return nil
}

// Шаг увеличения счетчиков. Функция
func (s *Service) incrementCounter(ctx context.Context) error {
	data := ctx.Value("data").(*sagaNewMessageData)
	d := *data
	s.log.Infof("incrementCounter data %+v", d)
	if d.userName == nil || d.savedMessage == nil {
		return fmt.Errorf("empty data")
	}
	_, err := s.counterApi.NewMessage(ctx, &counter_api.CounterEventRequest{
		UserId:    data.userFromId,
		MessageId: data.savedMessage.Id,
		ChatId:    data.chatId,
	})
	if err != nil {
		return err
	}

	return nil
}

// Шаг увеличения счетчиков. Компенсирующая функция.
func (s *Service) compIncrementCounter(ctx context.Context) error {
	data := ctx.Value("data").(*sagaNewMessageData)
	d := *data
	s.log.Infof("compIncrementCounter data %+v", d)
	_, err := s.counterApi.CompensateNewMessage(ctx, &counter_api.CounterEventRequest{
		UserId:    data.userFromId,
		MessageId: data.savedMessage.Id,
		ChatId:    data.chatId,
	})
	if err != nil {
		return err
	}
	return nil
}

// Сага для последовательного сохранения сообщения и атомарного изменения счетчика сообщений
func (s *Service) newMessageCounterSaga(ctx context.Context, data sagaNewMessageData) (*sagaNewMessageData, error) {
	// Saga
	sagaId := uuid.New().String()
	ctxFunc := context.WithValue(context.Background(), "data", &data)

	newMessageSaga := saga.NewSaga("New message Saga")

	_ = newMessageSaga.AddStep(&saga.Step{
		Name:           "Save message",
		Func:           s.saveMessage,
		CompensateFunc: s.compSaveMessage,
	})

	_ = newMessageSaga.AddStep(&saga.Step{
		Name:           "Increment count",
		Func:           s.incrementCounter,
		CompensateFunc: s.compIncrementCounter,
	})

	coordinator := saga.NewCoordinator(ctxFunc, ctxFunc, newMessageSaga, s.ss)
	playResult := coordinator.Play()

	s.log.Infof("saga result %s %+v", sagaId, playResult)

	return &data, playResult.ExecutionError
}

type sagaNewMessageData struct {
	savedMessage       *model.Message
	chatId, userFromId int64
	date               time.Time
	message            string
	userName           *string
}

func randomBool() bool {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(2) == 1
}
