package mysql

import (
	"context"
	"fmt"
	"github.com/basicus/hla-course/model"
	"github.com/jmoiron/sqlx"
	"time"
)

// UserGetChats Получить список чатов
func (d *dbc) UserGetChats(ctx context.Context, userId int64) ([]model.Chat, error) {
	var userChats []model.Chat
	// If RO connection is enabled use it
	connection := d.connection
	if d.roEnable {
		connection = d.connectionRo
	}
	// Get user chats
	chats, err := d.getUserChatIds(ctx, connection, userId)
	if err != nil {
		return nil, err
	}

	if chats != nil && len(chats) > 0 {
		userChats, err = d.getChatList(ctx, connection, chats)
		if err != nil {
			return nil, err
		}
	}
	return userChats, nil
}

// Получение id чатов пользователя
func (d *dbc) getUserChatIds(ctx context.Context, connection *sqlx.DB, userId int64) ([]int64, error) {
	var chats []model.ChatParticipant
	err := connection.SelectContext(ctx, &chats, "SELECT * from chat_participants where user_id=?", userId)
	if err != nil {
		return nil, err
	}

	var chatIds []int64

	for _, chat := range chats {
		chatIds = append(chatIds, chat.ChatId)
	}
	return chatIds, nil
}

func (d *dbc) getChatList(ctx context.Context, connection *sqlx.DB, chatIds []int64) ([]model.Chat, error) {
	var chats []model.Chat
	queryFriends, args, err := sqlx.In("SELECT * FROM chats WHERE chats.id IN (?)", chatIds)
	if err != nil {
		return nil, err
	}
	err = connection.SelectContext(ctx, &chats, d.connection.Rebind(queryFriends), args...)
	if err != nil {
		return nil, err
	}
	return chats, nil
}

// ChatCreate Создать новый чат
func (d *dbc) ChatCreate(ctx context.Context, title string, participants ...int64) (model.Chat, error) {
	sql := "insert into chats (title) values (?);"
	stmt, err := d.connection.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		return model.Chat{}, err
	}

	result, err := stmt.ExecContext(ctx, title)
	if err != nil {
		return model.Chat{}, err
	}
	chatId, err := result.LastInsertId()
	if err != nil {
		return model.Chat{}, err
	}
	chat, err := d.GetChat(ctx, chatId)
	if err != nil {
		return model.Chat{}, err
	}
	// Add participants
	err = d.ChatAddParticipants(ctx, chatId, participants)
	if err != nil {
		return model.Chat{}, err
	}

	return chat, nil
}

// ChatDelete Удалить чат
func (d *dbc) ChatDelete(ctx context.Context, chatId int64) (model.Chat, error) {
	panic("not implemented")
}

// ChatGetParticipants Получить список участников чата
func (d *dbc) ChatGetParticipants(ctx context.Context, chatId int64) ([]int64, error) {
	var chats []model.ChatParticipant
	// If RO connection is enabled use it
	connection := d.connection
	if d.roEnable {
		connection = d.connectionRo
	}
	err := connection.SelectContext(ctx, &chats, "SELECT * from chat_participants where chat_id=?", chatId)
	if err != nil {
		return nil, err
	}

	var participants []int64

	for _, chat := range chats {
		participants = append(participants, chat.UserId)
	}
	return participants, nil
}

// ChatAddParticipants Добавить участников чата
func (d *dbc) ChatAddParticipants(ctx context.Context, chatId int64, userIds []int64) error {
	participants, err := d.ChatGetParticipants(ctx, chatId)
	if err != nil {
		return err
	}
	alreadyParticipant := false

	for _, participant := range participants {
		for _, userId := range userIds {
			if participant == userId {
				alreadyParticipant = true
				break
			}
		}
		if alreadyParticipant {
			break
		}
	}
	c := 0
	if !alreadyParticipant {
		sql := "insert into chat_participants (chat_id,user_id) values (?, ?);"
		stmt, err := d.connection.Prepare(sql)
		defer stmt.Close()
		if err != nil {
			return err
		}
		for _, userId := range userIds {
			result, err := stmt.ExecContext(ctx, chatId, userId)
			if err != nil {
				return err
			}
			affected, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if affected > 0 {
				c++
			}
		}
		if c > 0 {
			return nil
		}
	}
	return nil
}

// ChatLeave Покинуть чат
func (d *dbc) ChatLeave(ctx context.Context, chatId, userId int64) error {
	sql := "delete from chat_participants where chat_id = ? and user_id =? ;"
	stmt, err := d.connection.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		return err
	}

	result, err := stmt.ExecContext(ctx, chatId, userId)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}
	return fmt.Errorf("user is no participant")
}

// MessageSave Отправить сообщение в чат
func (d *dbc) MessageSave(ctx context.Context, chatId, userFromId int64, date time.Time, message string) (model.Message, error) {
	sql := "insert into messages (chat_id, user_from, send_at, message) values (?,?,?,?);"
	stmt, err := d.connection.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		return model.Message{}, err
	}

	result, err := stmt.ExecContext(ctx, chatId, userFromId, date, message)
	if err != nil {
		return model.Message{}, err
	}
	messageId, err := result.LastInsertId()
	if err != nil {
		return model.Message{}, err
	}
	messageSaved, err := d.MessageGet(ctx, messageId)
	if err != nil {
		return model.Message{}, err
	}

	return messageSaved, nil
}

// ChatMessages Получение списка сообщений из чата
func (d *dbc) ChatMessages(ctx context.Context, chatId int64, limit, offset int64) ([]model.Message, error) {
	var messages []model.Message
	// If RO connection is enabled use it
	connection := d.connection
	if d.roEnable {
		connection = d.connectionRo
	}
	err := connection.SelectContext(ctx, &messages, "SELECT * from messages where chat_id=?", chatId)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (d *dbc) MessageGet(ctx context.Context, id int64) (model.Message, error) {
	var message model.Message
	err := d.connection.GetContext(ctx, &message, "SELECT * from messages where id=?", id)
	if err != nil {
		return model.Message{}, err
	}
	return message, nil
}

// GetChat Получить информацию о чате по id
func (d *dbc) GetChat(ctx context.Context, chatId int64) (model.Chat, error) {
	var chat model.Chat
	err := d.connection.GetContext(ctx, &chat, "SELECT * from chats where id=?", chatId)
	if err != nil {
		return model.Chat{}, err
	}

	return chat, nil
}
