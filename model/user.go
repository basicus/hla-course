package model

import "errors"

// User model
type User struct {
	UserId       int64        `json:"id" db:"user_id" fake:"skip"`
	Login        string       `json:"login" db:"login" fake:"{regex:[abcdefghijklmnopqrstuvwxyz0123456789]{10}}"`
	Email        string       `json:"email" db:"email" fake:"{email}"`
	Phone        string       `json:"phone" db:"phone" fake:"{phone}"`
	Name         string       `json:"name" db:"name" fake:"{firstname}"`
	Surname      string       `json:"surname" db:"surname" fake:"{lastname}"`
	Age          int          `json:"age" db:"age" fake:"{number:1,100}"`
	Sex          string       `json:"sex" db:"sex" fake:"{randomstring:[male,female]}""`
	Country      string       `json:"country" db:"country" fake:"{country}"`
	City         string       `json:"city" db:"city" fake:"{city}"`
	Interests    string       `json:"interests" db:"interests" fake:"{sentence}"`
	Password     string       `json:"password,omitempty" db:"-" fake:"{password}"`
	PasswordHash string       `json:"-" db:"password" fake:"skip"`
	ShardId      string       `json:"-" db:"shard_id"`
	Settings     UserSettings `json:"settings,omitempty" db:"-"`
}

type UserSettings struct {
	WSConnect string `json:"ws_connect_uri"`
}

func (u *User) ValidateRegister() error {
	if u.Name == "" {
		return errors.New("name is empty")
	}
	if u.Password == "" {
		return errors.New("password must be set")
	}
	if u.Age <= 0 && u.Age >= 105 {
		return errors.New("age must be between 1 and 104 years")
	}
	if u.Login == "" {
		return errors.New("login is required")
	}
	if !(u.Sex == "male" || u.Sex == "female") {
		return errors.New("sex must be male or female")
	}

	return nil
}

func (u *User) ValidateUpdate() (map[string]struct{}, error) {
	upd := make(map[string]struct{})
	if u.Name != "" {
		upd["name"] = struct{}{}
	}
	if u.Age > 0 && u.Age < 105 {
		upd["age"] = struct{}{}
	}
	if u.Country != "" {
		upd["country"] = struct{}{}
	}

	if u.Sex != "male" && u.Sex != "female" {
		upd["sex"] = struct{}{}
	}

	if u.Surname != "" {
		upd["surname"] = struct{}{}
	}

	if u.City != "" {
		upd["city"] = struct{}{}
	}

	if u.Country != "" {
		upd["country"] = struct{}{}
	}

	if u.Phone != "" {
		upd["phone"] = struct{}{}
	}

	if u.Email != "" {
		upd["email"] = struct{}{}
	}
	return upd, nil
}
