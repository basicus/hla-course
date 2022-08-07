package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adjust/rmq/v4"
	"github.com/basicus/hla-course/log"
	"github.com/basicus/hla-course/storage"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type ConsumerUserFeed struct {
	name    string
	count   int
	before  time.Time
	logger  *logrus.Logger
	storage storage.UserService
	redis   *redis.Client
}

func NewConsumerUserFeed(tag string, logger *logrus.Logger, service *storage.UserService, redis *redis.Client) *ConsumerUserFeed {
	return &ConsumerUserFeed{
		name:    fmt.Sprintf("consumer-%s", tag),
		count:   0,
		before:  time.Now(),
		logger:  logger,
		storage: *service,
		redis:   redis,
	}
}

func (c *ConsumerUserFeed) Consume(delivery rmq.Delivery) {
	var task TaskUpdateUserIdFeed

	if err := json.Unmarshal([]byte(delivery.Payload()), &task); err != nil {
		if err := delivery.Reject(); err != nil {
			c.logger.WithField("consumer", c.name).Errorf("cant unmarshall task with user_id %d: %e", task.UserId, err)
		}
		return
	}

	c.logger.WithField("consumer", c.name).Infof("consume update feed for user_id %d", task.UserId)
	c.count++

	// Выполняем поиск всех подписчиков и добавляем их в очередь обновления ленты
	fields := logrus.Fields{
		"consumer": c.name,
		"user_id":  task.UserId,
	}
	ctx := log.WithContext(context.Background(), c.logger.WithFields(fields))

	// Обновляем кэш пользователя
	friendsPosts, err := c.storage.GetFriendsPosts(ctx, task.UserId, 1000)
	if err != nil {
		return
	}
	friendPostsJson, err := json.Marshal(friendsPosts)
	if err != nil {
		c.logger.WithField("consumer", c.name).Errorf("error on marshalling friends posts %e", err)
	}

	_, err = c.redis.Set(ctx, "user_feed"+strconv.FormatInt(task.UserId, 10), friendPostsJson, 0).Result()
	if err != nil {
		c.logger.WithField("consumer", c.name).Errorf("error on set cached value for feed user_id %d  %e", task.UserId, err)
	}

	if err := delivery.Ack(); err != nil {
		c.logger.WithFields(fields).Errorf("post error ack queue update feed for user_id %d: %e", task.UserId, err)

	} else {
		c.logger.WithFields(fields).Infof("acked task update feed for user_id %d", task.UserId)
	}

	// Сообщает о скорости обработки запросов
	if c.count%consumerReportBatchSize == 0 {
		duration := time.Now().Sub(c.before)
		c.before = time.Now()
		perSecond := time.Second / (duration / consumerReportBatchSize)
		c.logger.WithField("consumer", c.name).Infof("consumed %d %d r/s", c.count, perSecond)
	}
}
