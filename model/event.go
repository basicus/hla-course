package model

import "encoding/json"

const (
	EventTypeInvite  string = "invite"
	EventTypePost    string = "post"
	EventTypeMessage string = "message"
)

// Модели данных необходимых для отправки в очередь

type Event interface {
	String() string
	GetType() string
}

// EventPost Событие Публикация
type EventPost struct {
	EventType string `json:"event"`
	PostDTO   `json:"data"`
}

func (e *EventPost) String() string {
	e.EventType = EventTypePost
	bytes, err := json.Marshal(e)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func (e *EventPost) GetType() string {
	return EventTypePost
}

// EventChatMessage Событие Сообщение в чате
type EventChatMessage struct {
	EventType  string `json:"event"`
	Chat       Chat   `json:"chat"`
	MessageDTO `json:"data"`
}

func (e *EventChatMessage) String() string {
	e.EventType = EventTypeMessage
	bytes, err := json.Marshal(e)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func (e *EventChatMessage) GetType() string {
	return EventTypeMessage
}

// EventChatInvite Событие приглашение в чат
type EventChatInvite struct {
	EventType string `json:"event"`
	Chat      Chat   `json:"chat"`
}

func (e *EventChatInvite) String() string {
	e.EventType = EventTypeInvite
	bytes, err := json.Marshal(e)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func (e *EventChatInvite) GetType() string {
	return EventTypeInvite
}
