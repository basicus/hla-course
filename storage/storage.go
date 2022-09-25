package storage

import (
	"context"
	"errors"
	"github.com/basicus/hla-course/model"
	"time"
)

var (
	ErrInvalidUserOrPassword = errors.New("invalid password or user not found")
)

type UserService interface {
	// GetById Получение информации о пользователе по id
	GetById(ctx context.Context, id int64) (model.User, error)
	// GetByLogin Поиск пользователя по логину
	GetByLogin(ctx context.Context, login string) (model.User, error)
	// GetUsers Поиск пользователей с фильтрацией по полям
	GetUsers(ctx context.Context, filter map[string]string, order map[string]string, offset, limit int) ([]model.User, error)
	// ValidateUser Проверка пароля пользователя
	ValidateUser(ctx context.Context, login string, password string) (bool, error)
	// CheckPasswordHash Сравнение хэша с паролем
	CheckPasswordHash(ctx context.Context, password, hash string) bool
	// Create Создать пользователя
	Create(ctx context.Context, user model.User) (model.User, error)
	// Update Обновить пользователя
	Update(ctx context.Context, user model.User, fieldsForUpdating map[string]struct{}) (model.User, error)
	// GetFriends Получение списка пользователей
	GetFriends(ctx context.Context, id int64) ([]model.User, error)
	// GetUserFollowers Получить список followers пользователей
	GetUserFollowers(ctx context.Context, id int64) ([]int64, error)
	// AddFriend Добавить пользователя
	AddFriend(ctx context.Context, user int64, friend int64) (bool, error)
	// DelFriend Удалить пользователя из друзей
	DelFriend(ctx context.Context, user int64, friend int64) (bool, error)
	// PublishPost Опубликовать запись
	PublishPost(ctx context.Context, user int64, title, message string) (model.Post, error)
	// GetFriendsPosts Получение ленты друзей
	GetFriendsPosts(ctx context.Context, id int64, limit int64) ([]model.Post, error)
	// GetPostsByUserId Получить список постов пользователя
	GetPostsByUserId(ctx context.Context, userId int64, limit, offset int64) ([]model.Post, error)
	// GetPostById Получить post по его Id
	GetPostById(ctx context.Context, postId int64) (model.Post, error)
	// GetUserName Получить имя пользователя
	GetUserName(ctx context.Context, userId int64) (string, error)
	// GetLogin Получить логин пользователя
	GetLogin(ctx context.Context, userId int64) (string, error)
}

type ChatsService interface {
	// GetChat Получить информацию о чате по id
	GetChat(ctx context.Context, chatId int64) (model.Chat, error)
	// UserGetChats Получить список чатов
	UserGetChats(ctx context.Context, userId int64) ([]model.Chat, error)
	// ChatCreate Создать новый чат
	ChatCreate(ctx context.Context, title string, participants ...int64) (model.Chat, error)
	// ChatDelete Удалить чат
	ChatDelete(ctx context.Context, chatId int64) (model.Chat, error)
	// ChatGetParticipants Получить список участников чата
	ChatGetParticipants(ctx context.Context, chatId int64) ([]int64, error)
	// ChatAddParticipants Добавить участников чата
	ChatAddParticipants(ctx context.Context, chatId int64, userIds []int64) error
	// ChatLeave Покинуть чат
	ChatLeave(ctx context.Context, chatId, userId int64) error
	// MessageSave Отправить сообщение в чат
	MessageSave(ctx context.Context, chatId, userFromId int64, date time.Time, message string) (model.Message, error)
	// ChatMessages Получение списка сообщений из чата
	ChatMessages(ctx context.Context, chatId int64, limit, offset int64) ([]model.Message, error)
	// MessageGet Получить сообщение по id
	MessageGet(ctx context.Context, id int64) (model.Message, error)
}
