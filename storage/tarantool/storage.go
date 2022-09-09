package tarantool

import (
	"context"
	"github.com/basicus/hla-course/model"
	"github.com/basicus/hla-course/storage"
	"github.com/sirupsen/logrus"
	"github.com/tarantool/go-tarantool"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type dbc struct {
	logger     *logrus.Logger
	connection *tarantool.Connection
	cfg        Config
}

func New(cfg Config, logger *logrus.Logger) (storage.UserService, error) {

	conn, err := tarantool.Connect(cfg.DSN, tarantool.Opts{
		User: cfg.UserName,
		Pass: cfg.Password,
	})

	if err != nil {
		return nil, err
	}
	_, err = conn.Ping()

	// Call bootstrap function
	if err != nil {
		return nil, err
	}

	return &dbc{
		logger:     logger.WithField("role", "tarantool").Logger,
		connection: conn,
		cfg:        cfg,
	}, nil
}

func (d *dbc) GetById(ctx context.Context, id int64) (model.User, error) {
	resp, err := d.connection.Select(d.cfg.Space, "primary", 0, 1, tarantool.IterEq, []interface{}{id})
	user := model.User{}
	if err != nil {
		return user, err
	}
	if resp.Code != tarantool.OkCode {
		d.logger.Errorf("Select failed: %s", resp.Error)
		return user, err
	}

	if len(resp.Data) == 0 {
		return user, err
	}

	tuples := resp.Tuples()
	for _, tuple := range tuples {
		for i, dd := range tuple {
			switch i {
			case 0:
				user.UserId = int64(dd.(uint64))
			case 1:
				user.Login = dd.(string)
			case 2:
				user.Email = dd.(string)
			case 3:
				user.Phone = dd.(string)
			case 4:
				user.PasswordHash = dd.(string)
			case 5:
				user.Name = dd.(string)
			case 6:
				user.Surname = dd.(string)
			}
		}
	}
	return user, err
}

func (d *dbc) GetByLogin(ctx context.Context, login string) (model.User, error) {
	resp, err := d.connection.Select(d.cfg.Space, "login", 0, 1, tarantool.IterEq, []interface{}{login})
	user := model.User{}
	if err != nil {
		return user, err
	}
	if resp.Code != tarantool.OkCode {
		d.logger.Errorf("Select failed: %s", resp.Error)
		return user, err
	}

	if len(resp.Data) == 0 {
		return user, err
	}

	tuples := resp.Tuples()
	for _, tuple := range tuples {
		for i, dd := range tuple {
			switch i {
			case 0:
				user.UserId = int64(dd.(uint64))
			case 1:
				user.Login = dd.(string)
			case 2:
				user.Email = dd.(string)
			case 3:
				user.Phone = dd.(string)
			case 4:
				user.PasswordHash = dd.(string)
			case 5:
				user.Name = dd.(string)
			case 6:
				user.Surname = dd.(string)
			}
		}
	}
	return user, err
}

func (d *dbc) GetUsers(_ context.Context, _ map[string]string, _ map[string]string, _, _ int) ([]model.User, error) {
	panic("implement me")
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

func (d *dbc) CheckPasswordHash(ctx context.Context, password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (d *dbc) Create(_ context.Context, _ model.User) (model.User, error) {
	panic("implement me")
}

func (d *dbc) Update(_ context.Context, _ model.User, _ map[string]struct{}) (model.User, error) {
	panic("implement me")
}

func (d *dbc) GetFriends(_ context.Context, _ int64) ([]model.User, error) {
	panic("implement me")
}

func (d *dbc) GetUserFollowers(_ context.Context, _ int64) ([]int64, error) {
	panic("implement me")
}

func (d *dbc) AddFriend(_ context.Context, _ int64, _ int64) (bool, error) {
	panic("implement me")
}

func (d *dbc) DelFriend(ctx context.Context, user int64, friend int64) (bool, error) {
	panic("implement me")
}

func (d *dbc) PublishPost(_ context.Context, _ int64, _, _ string) (model.Post, error) {
	panic("implement me")
}

func (d *dbc) GetFriendsPosts(_ context.Context, _ int64, _ int64) ([]model.Post, error) {
	panic("implement me")
}

func (d *dbc) GetPostsByUserId(_ context.Context, _ int64, _, _ int64) ([]model.Post, error) {
	panic("implement me")
}

func (d *dbc) GetPostById(_ context.Context, _ int64) (model.Post, error) {
	panic("implement me")
}

func (d *dbc) UserGetChats(_ context.Context, _ int64) ([]model.Chat, error) {
	panic("implement me")
}

func (d *dbc) ChatCreate(_ context.Context, _ string, _ ...int64) (model.Chat, error) {
	panic("implement me")
}

func (d *dbc) ChatDelete(_ context.Context, _ int64) (model.Chat, error) {
	panic("implement me")
}

func (d *dbc) ChatGetParticipants(_ context.Context, _ int64) ([]int64, error) {
	panic("implement me")
}

func (d *dbc) ChatAddParticipants(_ context.Context, _ int64, _ []int64) error {
	panic("implement me")
}

func (d *dbc) ChatLeave(_ context.Context, _, _ int64) error {
	panic("implement me")
}

func (d *dbc) MessageSave(_ context.Context, _, _ int64, _ time.Time, _ string) (model.Message, error) {
	panic("implement me")
}

func (d *dbc) ChatMessages(_ context.Context, _ int64, _, _ int64) ([]model.Message, error) {
	panic("implement me")
}

func (d *dbc) MessageGet(_ context.Context, _ int64) (model.Message, error) {
	panic("implement me")
}

func (d *dbc) GetUserName(ctx context.Context, userId int64) (string, error) {
	user, err := d.GetById(ctx, userId)
	if err != nil {
		return "", err
	}
	return user.Name + " " + user.Surname, nil
}

func (d *dbc) GetLogin(ctx context.Context, userId int64) (string, error) {
	user, err := d.GetById(ctx, userId)
	if err != nil {
		return "", err
	}
	return user.Login, nil
}

func (d *dbc) GetChat(_ context.Context, _ int64) (model.Chat, error) {
	panic("implement me")
}
