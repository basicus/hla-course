package queue

import (
	"encoding/json"
	"github.com/adjust/rmq/v4"
	"github.com/basicus/hla-course/model"
	"sync"
	"time"
)

// TaskQueue Очередь Задач
type TaskQueue struct {
	rmq.Queue
	done int64
	m    sync.Mutex
}

// TaskPost Задача на обновление поста
type TaskPost struct {
	Post      model.Post
	QueueDate time.Time `json:"queue_date"`
}

type TaskUpdateUserIdFeed struct {
	UserId int64
}

func (t *TaskQueue) AddTaskPost(post model.Post) error {
	taskBytes, err := json.Marshal(TaskPost{
		Post:      post,
		QueueDate: time.Now(),
	})
	if err != nil {
		return err
	}

	err = t.Queue.PublishBytes(taskBytes)
	if err != nil {
		return err
	}
	return nil
}

func (t *TaskQueue) AddTaskUpdateUserIdFeed(userId int64) error {
	taskBytes, err := json.Marshal(TaskUpdateUserIdFeed{
		UserId: userId,
	})
	if err != nil {
		return err
	}

	err = t.Queue.PublishBytes(taskBytes)
	if err != nil {
		return err
	}
	return nil
}

func (t *TaskQueue) TaskDone() {
	defer t.m.Unlock()
	t.m.Lock()
	t.done++
}
