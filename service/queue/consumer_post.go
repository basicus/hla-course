package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adjust/rmq/v4"
	"github.com/basicus/hla-course/log"
	"github.com/basicus/hla-course/model"
	"github.com/basicus/hla-course/storage"
	"github.com/sirupsen/logrus"
	"time"
)

type ConsumerPost struct {
	name    string
	count   int
	before  time.Time
	logger  *logrus.Logger
	storage storage.UserService
	publish func(ctx context.Context, userId int64, shardId string, event model.Event) error
	queue   *TaskQueue
}

func NewConsumerPost(tag string, logger *logrus.Logger, service *storage.UserService, queue *TaskQueue, publish func(ctx context.Context, userId int64, shardId string, event model.Event) error) *ConsumerPost {
	return &ConsumerPost{
		name:    fmt.Sprintf("consumer-%s", tag),
		count:   0,
		before:  time.Now(),
		logger:  logger,
		storage: *service,
		queue:   queue,
		publish: publish,
	}
}

func (c *ConsumerPost) Consume(delivery rmq.Delivery) {
	var task TaskPost

	if err := json.Unmarshal([]byte(delivery.Payload()), &task); err != nil {
		if err := delivery.Reject(); err != nil {
			c.logger.WithField("consumer", c.name).Errorf("cant unmarshall task with post id  %d for user_id %d: %e", task.Post.Id, task.Post.UserId, err)
		}
		return
	}

	c.logger.WithField("consumer", c.name).Infof("consume post_id %d for user_id %d", task.Post.Id, task.Post.UserId)
	c.count++

	// Выполняем поиск всех подписчиков и планируем их к обновлению ленты
	fields := logrus.Fields{
		"consumer": c.name,
		"post_id":  task.Post.Id,
	}
	ctx := log.WithContext(context.Background(), c.logger.WithFields(fields))

	followers, err := c.storage.GetUserFollowers(ctx, task.Post.UserId)

	if err != nil {
		// Cant get user followers
		c.logger.WithFields(fields).Infof("consume post_id %d error for user_id %d  when get follower list: %e", task.Post.Id, task.Post.UserId, err)
		_ = delivery.Reject()
	}

	var userName string
	userName, err = c.storage.GetUserName(ctx, task.Post.UserId)
	if err != nil {
		userName = ""
	}

	// Ставим на обновление кэши пользователей
	for _, followerId := range followers {
		// Queue follower id for update feed
		err := c.queue.AddTaskUpdateUserIdFeed(followerId)
		if err != nil {
			c.logger.WithFields(fields).Errorf("error on queue for update feed for follower_id %d: %s", followerId, err)
		} else {
			c.logger.WithFields(fields).Infof("successfully queued for update feed for follower_id %d", followerId)
		}
		// Пытаемся отправить событие (не самое оптимальное место)
		userInfo, err := c.storage.GetById(ctx, followerId)
		if err == nil {
			// User found
			eventPost := model.EventPost{
				Data: model.PostDTO{
					UserFrom: userName,
					Title:    task.Post.Title,
					Message:  task.Post.Message,
				},
			}
			err := c.publish(ctx, userInfo.UserId, userInfo.ShardId, &eventPost)
			if err != nil {
				c.logger.Errorf("error when publish post info: %s", err.Error())
			}
		}
	}

	if err := delivery.Ack(); err != nil {
		c.logger.WithFields(fields).Errorf("post error ack queue update followers post_id %d for user_id %d: %e", task.Post.Id, task.Post.UserId, err)

	} else {
		c.logger.WithFields(fields).Infof("acked task post_id %d for user_id %d", task.Post.Id, task.Post.UserId)
	}

	// Сообщает по скорости обработки запросов
	if c.count%consumerReportBatchSize == 0 {
		duration := time.Now().Sub(c.before)
		c.before = time.Now()
		perSecond := time.Second / (duration / consumerReportBatchSize)
		c.logger.WithField("consumer", c.name).Infof("consumed %d %d r/s", c.count, perSecond)
	}
}
