package model

// Friend Друг друга
type Friend struct {
	UserId   int64 `json:"user_id" db:"user_id"`
	FriendId int64 `json:"friend_id" db:"friend_id"`
}
