package mysql

import (
	"context"
	"github.com/basicus/hla-course/model"
	"github.com/basicus/hla-course/storage"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

const (
	SqlN     = " "
	SqlAnd   = " AND "
	SqlOr    = " OR "
	SqlLike  = " LIKE ?"
	SqlEq    = " = ?"
	SqlWhere = " WHERE "
	SqlOrder = " ORDER BY "
	SqlComma = " , "
)

var fieldsOrder = []string{"user_id", "name", "surname", "age", "country", "city", "interests"}

func (d *dbc) GetById(ctx context.Context, id int64) (model.User, error) {
	var user model.User
	err := d.connection.GetContext(ctx, &user, "SELECT * from users where user_id=?", id)
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (d *dbc) GetByLogin(ctx context.Context, login string) (model.User, error) {
	var user model.User
	err := d.connection.GetContext(ctx, &user, "SELECT * from users where login=?", login)
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (d *dbc) GetUsers(ctx context.Context, filter map[string]string, order map[string]string, offset, limit int) ([]model.User, error) {
	var users []model.User
	var sb strings.Builder
	var args []interface{}
	c := 0
	sb.WriteString("SELECT * from users ")
	if len(filter) > 0 {
		sb.WriteString(SqlWhere)
	}
	for _, field := range fieldsOrder {
		v, ok := filter[field]
		if ok {
			if c == 0 {
				sb.WriteString(SqlN)
			} else {
				sb.WriteString(SqlAnd)
			}
			sb.WriteString(field)
			sb.WriteString(SqlLike)
			args = append(args, v)
			c++
		}
	}
	c = 0
	if len(order) > 0 {
		sb.WriteString(SqlOrder)
	}

	for _, field := range fieldsOrder {
		v, ok := order[field]
		if ok {
			if c == 0 {
				sb.WriteString(SqlN)
			} else {
				sb.WriteString(SqlComma)
			}
			sb.WriteString(field)
			sb.WriteString(SqlN)
			sb.WriteString(v)
			c++
		}
	}

	if offset > 0 {
		sb.WriteString(" OFFSET ? ")
		args = append(args, offset)
	}
	if limit > 0 {
		sb.WriteString(" LIMIT ? ")
		args = append(args, limit)
	}
	// If RO connection is enabled use it
	connection := d.connection
	if d.roEnable {
		connection = d.connectionRo
	}
	err := connection.SelectContext(ctx, &users, sb.String(), args...)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (d *dbc) ValidateUser(ctx context.Context, login string, password string) (bool, error) {
	user, err := d.GetByLogin(ctx, login)
	if err != nil {
		return false, err
	}

	if user.PasswordHash != "" {
		ch := d.CheckPasswordHash(ctx, password, user.PasswordHash)
		if ch {
			return true, nil
		}
	}
	return false, storage.ErrInvalidUserOrPassword
}

func (d *dbc) CheckPasswordHash(_ context.Context, password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (d *dbc) Create(ctx context.Context, user model.User) (model.User, error) {
	sql := "insert into users (login, email, phone, password, name, surname, age, sex, country, city, interests) " +
		"values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);"
	stmt, err := d.connection.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		return model.User{}, err
	}
	user.PasswordHash, _ = hashPassword(user.Password)
	result, err := stmt.ExecContext(ctx, user.Login, user.Email, user.Phone, user.PasswordHash, user.Name, user.Surname,
		user.Age, user.Sex, user.Country, user.City, user.Interests)
	if err != nil {
		return model.User{}, err
	}
	userId, err := result.LastInsertId()
	if err != nil {
		return model.User{}, err
	}
	userDb, err := d.GetById(ctx, userId)
	if err != nil {
		return model.User{}, err
	}
	return userDb, nil
}

func (d *dbc) Update(ctx context.Context, user model.User, fieldsForUpdating map[string]struct{}) (model.User, error) {
	var sb strings.Builder
	sb.WriteString("update users set ")
	if fieldsForUpdating == nil {
		sb.WriteString(" phone=:phone, name=:name, surname=:surname, age=:age, sex=:sex, country=:country, " +
			"city=:city, interests=:interests")
	} else {
		count := len(fieldsForUpdating)
		for field := range fieldsForUpdating {
			sb.WriteString(field + "=:" + field)
			if count == 1 {
				sb.WriteString(" ")
			} else {
				sb.WriteString(", ")
				count--
			}
		}
		sb.WriteString("where user_id=:user_id")
	}

	_, err := d.connection.NamedExecContext(ctx, sb.String(), user)
	if err != nil {
		return model.User{}, err
	}
	userDb, err := d.GetById(ctx, user.UserId)
	if err != nil {
		return model.User{}, err
	}
	return userDb, nil
}

func (d *dbc) GetFriends(ctx context.Context, id int64) ([]model.User, error) {
	var friends []model.Friend
	var friendsUsers []model.User
	// If RO connection is enabled use it
	connection := d.connection
	if d.roEnable {
		connection = d.connectionRo
	}
	err := connection.SelectContext(ctx, &friends, "SELECT * from user_friend where user_id=?", id)
	if err != nil {
		return nil, err
	}

	var friendsId []int64

	for _, friend := range friends {
		friendsId = append(friendsId, friend.FriendId)
	}
	queryFriends, args, err := sqlx.In("SELECT * FROM users WHERE users.user_id IN (?)", friendsId)
	if err != nil {
		return nil, err
	}
	err = connection.SelectContext(ctx, &friendsUsers, d.connection.Rebind(queryFriends), args...)
	if err != nil {
		return nil, err
	}

	return friendsUsers, nil

}

func (d *dbc) AddFriend(ctx context.Context, user int64, friend int64) (bool, error) {
	sql := "insert into user_friend (user_id, friend_id) " +
		"values (?, ?);"
	stmt, err := d.connection.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		return false, err
	} // TODO Check for duplicates

	result, err := stmt.ExecContext(ctx, user, friend)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if affected > 0 {
		return true, err
	}
	return false, err

}

func (d *dbc) DelFriend(ctx context.Context, user int64, friend int64) (bool, error) {
	sql := "delete from user_friend where user_id = ? and friend_id = ? ;"

	stmt, err := d.connection.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		return false, err
	}

	result, err := stmt.ExecContext(ctx, user, friend)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if affected > 0 {
		return true, err
	}
	return false, err
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 1)
	return string(bytes), err
}