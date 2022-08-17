package queue

import (
	"context"
	"errors"
	"fmt"
	"github.com/adjust/rmq/v4"
	"github.com/basicus/hla-course/log"
	"github.com/basicus/hla-course/model"
	"github.com/basicus/hla-course/storage"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	RedisTimelineTag = "timeline-update"
	queueNamePosts   = "post"
	queueNameFeed    = "feed"

	prefetchLimit = 10
	pollPeriod    = 1 * time.Second
)

var (
	ErrUnknownQueue   = errors.New("unknown queue")
	ErrNoOrEmptyQueue = errors.New("empty or unknown queue")
)

type Service struct {
	connection  rmq.Connection
	redisClient *redis.Client
	log         *logrus.Logger
	storage     *storage.UserService
	queues      map[string]*TaskQueue
	err         chan error
	config      Config
	counters    map[string]*qStatCounters
	publish     func(ctx context.Context, userId int64, shardId string, event model.Event) error
}

// New Инициализация сервиса
func New(config Config, storage *storage.UserService, logger *logrus.Logger, publish func(ctx context.Context, userId int64, shardId string, event model.Event) error) (*Service, error) {
	e := make(chan error, 10)
	redisClient := redis.NewClient(
		&redis.Options{
			Addr:      config.Address,
			OnConnect: nil,
			Username:  config.UserName,
			Password:  config.Password,
			DB:        config.Database,
			PoolSize:  config.PoolSize,
		},
	)
	connection, err := rmq.OpenConnectionWithRedisClient(RedisTimelineTag, redisClient, e)
	if err != nil {
		return nil, err
	}
	logger.Info("Created new instance Queue service")
	return &Service{
		connection:  connection,
		err:         e,
		log:         logger,
		redisClient: redisClient,
		storage:     storage,
		queues:      make(map[string]*TaskQueue),
		config:      config,
		publish:     publish,
	}, nil
}

// NewPost Обновление лент у пользователей
func (s *Service) NewPost(ctx context.Context, post model.Post) error {
	logger := log.Ctx(ctx)
	logger.Infof("request new post for user_id %d", post.UserId)
	queue := queueNamePosts

	// Создать очередь
	taskQueue, err := s.getQueue(queue)
	if err != nil {
		return err
	}
	err = taskQueue.AddTaskPost(post)
	if err != nil {
		logger.WithError(err).Error("error on adding post to queue")
		return err
	}

	logger.Infof("add task for update feed of followers user_id %d is success",
		post.UserId)
	return nil
}

func (s *Service) UpdateFeed(ctx context.Context, userId int64) error {
	logger := log.Ctx(ctx)
	logger.Infof("request update feed for user_id %d", userId)
	queue := queueNameFeed

	// Создать очередь
	taskQueue, err := s.getQueue(queue)
	if err != nil {
		return err
	}
	err = taskQueue.AddTaskUpdateUserIdFeed(userId)
	if err != nil {
		logger.WithError(err).Error("error on adding queue update feed to queue")
		return err
	}

	logger.Infof("add task for update feed of user_id %d is success",
		userId)
	return nil
}

// StartConsumers Запустить консьюмеры
func (s *Service) StartConsumers(_ context.Context) error {
	taskPostsQueue, err := s.getQueue(queueNamePosts)
	if err != nil {
		return err
	}

	err = taskPostsQueue.StartConsuming(prefetchLimit, pollPeriod)
	if err != nil {
		return err
	}
	taskFeedsQueue, err := s.getQueue(queueNameFeed)
	if err != nil {
		return err
	}

	// Запуск консьюмеров
	for i := 0; i < s.config.NumberConsumersForQueue; i++ {
		name := fmt.Sprintf("consumer-%s", queueNamePosts)
		s.log.Infof("adding consumer %d name %s", i, name)
		if _, err := taskPostsQueue.AddConsumer(name, NewConsumerPost(fmt.Sprintf("%s-%d", name, i), s.log, s.storage, taskFeedsQueue, s.publish)); err != nil {
			return err
		}
	}

	err = taskFeedsQueue.StartConsuming(prefetchLimit, pollPeriod)
	if err != nil {
		return err
	}

	for i := 0; i < s.config.NumberConsumersForQueue; i++ {
		name := fmt.Sprintf("consumer-%s", queueNameFeed)
		s.log.Infof("adding consumer %d name %s", i, name)
		if _, err := taskFeedsQueue.AddConsumer(name, NewConsumerUserFeed(fmt.Sprintf("%s-%d", name, i), s.log, s.storage, s.redisClient)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) getQueueStats(queue string) (rmq.QueueStat, error) {
	stats, err := s.connection.CollectStats([]string{queue})
	if err != nil {
		return rmq.QueueStat{}, err
	}
	if len(stats.QueueStats) > 1 {
		return rmq.QueueStat{}, ErrUnknownQueue
	}
	if len(stats.QueueStats) == 0 {
		return rmq.QueueStat{}, ErrNoOrEmptyQueue
	}

	return stats.QueueStats[queue], nil

}

// getQueue Создает или возвращает очередь задач
func (s *Service) getQueue(queue string) (*TaskQueue, error) {
	taskQ, ok := s.queues[queue]
	if !ok {
		openQueue, err := s.connection.OpenQueue(queue)
		if err != nil {
			return nil, err
		}
		task := &TaskQueue{Queue: openQueue}
		s.queues[queue] = task
		return task, nil
	}
	return taskQ, nil
}

func (s *Service) removeQueue(queue string) error {

	_, ok := s.queues[queue]
	if !ok {
		s.log.Infof("remove queue %s not found", queue)
		return nil
	}
	delete(s.queues, queue)
	s.log.Infof("queue %s deleted", queue)
	return nil
}

// Run Запуск сервиса
func (s *Service) Run(ctx context.Context) error {
	logger := log.Ctx(ctx)
	logger.Info("started queue service")

	go s.QueueCleaner(s.config, logger)

	logger.Info("starting consumers")
	err := s.StartConsumers(ctx)
	if err != nil {
		return err
	}

	for err := range s.err {
		switch err := err.(type) {
		case *rmq.HeartbeatError:
			if err.Count == rmq.HeartbeatErrorLimit {
				logger.WithError(err).Error("heartbeat error (limit)")
			} else {
				logger.WithError(err).Error("heartbeat error")
			}
		case *rmq.ConsumeError:
			logger.WithError(err).Error("consume error")
		case *rmq.DeliveryError:
			logger.WithError(err).Error("delivery error")
		default:
			logger.WithError(err).Error("other error")
		}
	}

	return nil
}

// Shutdown Graceful shutdown сервиса
func (s *Service) Shutdown(ctx context.Context) error {
	logger := log.Ctx(ctx)
	defer func() {
		logger.Info("Stop queue service")
	}()

	<-s.connection.StopAllConsuming() // Ожидаем завершения работы всех консьюмеров
	return nil
}

func (s *Service) QueueCleaner(config Config, log *logrus.Entry) {
	cleaner := rmq.NewCleaner(s.connection)

	for range time.Tick(config.CleanPeriod) {
		returned, err := cleaner.Clean()
		if err != nil {
			log.Errorf("failed to clean: %s", err)
			continue
		}
		log.Infof("cleaned unacked %d", returned)
	}
}
func (s *Service) GetRedisClient() *redis.Client {
	return s.redisClient
}
