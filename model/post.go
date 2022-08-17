package model

import "time"

// Post model
type Post struct {
	Id        int64     `json:"id" db:"id" fake:"skip"`
	UserId    int64     `json:"user_id" db:"user_id" fake:"skip"`
	Title     string    `json:"title" db:"title" fake:"{sentence}"`
	Message   string    `json:"message" db:"message" fake:"{sentence}"`
	CreatedAt time.Time `json:"created_at" db:"created_at" fake:"skip"`
	UpdateAt  time.Time `json:"updated_at" db:"updated_at" fake:"skip"`
	Deleted   bool      `json:"deleted" db:"deleted" fake:"skip"`
}

// PostPojo model
type PostPojo struct {
	UserId  int64  `fake:"skip"`
	Title   string `json:"title" db:"title" fake:"{sentence}"`
	Message string `json:"message" db:"message" fake:"{sentence}"`
}

// PostDTO Публикация (для отображения на клиенте)
type PostDTO struct {
	UserFrom string `json:"user_from"`
	Title    string `json:"title"`
	Message  string `json:"message"`
}

// WsEvent Используется для маршрутизации и отправки пользователю в websocket
type WsEvent struct {
	UserId  int64  `json:"user_id"`
	Message string `json:"message,omitempty"`
}
