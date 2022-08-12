package model

import "time"

// Chat Чат общение между двумя и более участниками
type Chat struct {
	Id        int64     `db:"id" json:"id,omitempty"`
	Title     string    `db:"title" json:"title,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at,omitempty"`
	Closed    bool      `db:"closed" json:"closed,omitempty"`
	// Добавить автора и права
}

type ChatCreateDTO struct {
	Title string  `json:"title,omitempty"`
	Users []int64 `json:"users"`
}

// ChatParticipant Список участников чата
type ChatParticipant struct {
	Id     int64 `json:"id,omitempty" db:"id"`
	ChatId int64 `json:"chat_id" db:"chat_id"`
	UserId int64 `json:"user_id" db:"user_id"`
	Status int8  `json:"status" db:"status"`
}

// Message Сообщение, которое сохраняется в БД
type Message struct {
	Id       int64     `json:"id,omitempty" db:"id"`
	ChatId   int64     `json:"chat_id" db:"chat_id"`
	UserFrom int64     `json:"user_from" db:"user_from"`
	SendAt   time.Time `json:"send_at" db:"send_at"`
	Message  string    `json:"message" db:"message"`
}

type MessageSendDTO struct {
	ChatId  int64  `json:"chat_id,omitempty" `
	Message string `json:"message"`
}

// ChatParticipantsDTO Участник чата (для отображения Клиенту)
type ChatParticipantsDTO struct {
	UserName string `json:"user_name"`
	Status   string `json:"status"`
}

// MessageDTO Сообщение (для отображения Клиентом)
type MessageDTO struct {
	Id       int64     `json:"id"`
	UserFrom string    `json:"user_from"`
	Date     time.Time `json:"date"`
	Message  string    `json:"message"`
}

// MessageList Список сообщений чата
type MessageList struct {
	Chat Chat         `json:"chat"`
	List []MessageDTO `json:"list"`
}
