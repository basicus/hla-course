package main

import (
	"context"
	"github.com/basicus/hla-course/log"
	"github.com/basicus/hla-course/model"
	"github.com/basicus/hla-course/storage/mysql"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/joeshaw/envdecode"
	"github.com/sirupsen/logrus"
)

type config struct {
	Logger log.Config
	Db     mysql.Config
}

const GenerateCount = 1000

// Add users to database with random fake data. Add 4 friends to them

func main() {
	var cfg config
	if err := envdecode.StrictDecode(&cfg); err != nil {
		logrus.WithError(err).Fatal("Cannot decode config envs")
	}

	logger := log.New(cfg.Logger)

	dbc, err := mysql.New(cfg.Db, logger)
	if err != nil {
		logger.WithError(err).Fatal("Cannot access to database")
	}
	faker := gofakeit.New(0)
	var userIds []int64

	for i := 0; i < GenerateCount; i++ {
		var user model.User
		err := faker.Struct(&user)
		user.Password = "123" // Comment to set random password
		if err != nil {
			continue
		}
		logger.Infof("generated %+v", user)
		create, err := dbc.Create(context.Background(), user)
		if err != nil {
			logger.WithError(err).Error("cant create user")
			continue
		}
		userIds = append(userIds, create.UserId)
		logger.Infof("user id %d hash %s", create.UserId, create.PasswordHash)
	}

	// Add/remove friends
	logger.Infof("Generating user friends")
	for i := 0; i < len(userIds); i++ {
		userId := userIds[i]
		// Add three friends
		friends := generateFriends(faker, userIds, 5)
		for j := 0; j < len(friends); j++ {
			_, err := dbc.AddFriend(context.Background(), userId, friends[j])
			if err != nil {
				logger.WithError(err).Error("cant add friend")
			}
		}
		// Delete second Friend :(
		_, err := dbc.DelFriend(context.Background(), userId, friends[2])
		if err != nil {
			logger.WithError(err).Error("cant delete friend")
		}

		// Get user friends
		//users, err := dbc.GetFriends(context.Background(), userId)
		//if err != nil {
		//	logger.WithError(err).Error("cant get users")
		//}
		//logger.Infof("User %d friends %+v", userId, users)
	}

}

func generateFriends(faker *gofakeit.Faker, userIds []int64, l int) []int64 {
	perm := faker.Rand.Perm(len(userIds) - 1)
	var friends []int64
	for i := 0; i < l; i++ {
		friends = append(friends, userIds[perm[i]])
	}
	return friends
}
