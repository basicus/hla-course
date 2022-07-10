package storage

import (
	"context"
	"errors"
	"github.com/basicus/hla-course/model"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidUserOrPassword = errors.New("invalid password or user not found")
)

type UserService interface {
	// GetById Получение информации о пользователе по id
	GetById(ctx context.Context, id int64) (model.User, error)
	// GetByLogin Поиск пользователя по логину
	GetByLogin(ctx context.Context, login string) (model.User, error)
	// GetUsers Поиск пользователей с фильтрацией по полям
	GetUsers(ctx context.Context, filter map[string]string) ([]model.User, error)
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
	// AddFriend Добавить пользователя
	AddFriend(ctx context.Context, user int64, friend int64) (bool, error)
	// DelFriend Удалить пользователя из друзей
	DelFriend(ctx context.Context, user int64, friend int64) (bool, error)
}
