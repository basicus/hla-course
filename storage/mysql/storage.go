package mysql

import (
	"github.com/basicus/hla-course/migrations"
	"github.com/basicus/hla-course/storage"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type dbc struct {
	logger       *logrus.Logger
	connection   *sqlx.DB
	connectionRo *sqlx.DB
	redis        *redis.Client
	roEnable     bool
}

func New(cfg Config, logger *logrus.Logger) (storage.UserService, error) {

	// Migrations
	source := bindata.Resource(migrations.AssetNames(), migrations.Asset)
	d, err := bindata.WithInstance(source)
	if err != nil {
		return nil, err
	}
	m, err := migrate.NewWithSourceInstance("go-bindata", d, "mysql://"+cfg.DSN)
	if err != nil {
		return nil, err
	}
	if err = m.Up(); err != nil {
		if err != migrate.ErrNoChange {
			return nil, err
		}
	}

	conn, err := sqlx.Open("mysql", cfg.DSN+"?parseTime=true")
	if err != nil {
		return nil, err
	}
	err = conn.Ping()
	conn.SetMaxOpenConns(cfg.MaxOpenConnections)
	if err != nil {
		return nil, err
	}

	var roEnable bool
	var connRo *sqlx.DB
	if cfg.DSNro != "" && !cfg.RoDisable {
		roEnable = true
		connRo, err = sqlx.Open("mysql", cfg.DSNro+"?parseTime=true")
		if err != nil {
			return nil, err
		}
		err = connRo.Ping()
		connRo.SetMaxOpenConns(cfg.MaxOpenConnections)
		if err != nil {
			return nil, err
		}
	}
	logger.WithField("role", "storage").Logger.Infof("Using readonly connection pooling: %t", roEnable)
	return &dbc{
		logger:       logger.WithField("role", "storage").Logger,
		connection:   conn,
		connectionRo: connRo,
		roEnable:     roEnable,
	}, nil
}

func NewChats(cfg Config, logger *logrus.Logger) (storage.ChatsService, error) {

	// Migrations
	source := bindata.Resource(migrations.AssetNames(), migrations.Asset)
	d, err := bindata.WithInstance(source)
	if err != nil {
		return nil, err
	}
	m, err := migrate.NewWithSourceInstance("go-bindata", d, "mysql://"+cfg.DSN)
	if err != nil {
		return nil, err
	}
	if err = m.Up(); err != nil {
		if err != migrate.ErrNoChange {
			return nil, err
		}
	}

	conn, err := sqlx.Open("mysql", cfg.DSN+"?parseTime=true")
	if err != nil {
		return nil, err
	}
	err = conn.Ping()
	conn.SetMaxOpenConns(cfg.MaxOpenConnections)
	if err != nil {
		return nil, err
	}

	var roEnable bool

	logger.WithField("role", "storage-chats").Logger.Infof("Using readonly connection pooling: %t", roEnable)
	return &dbc{
		logger:     logger.WithField("role", "storage-chats").Logger,
		connection: conn,
		roEnable:   roEnable,
	}, nil
}
