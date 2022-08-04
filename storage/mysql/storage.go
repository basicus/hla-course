package mysql

import (
	"github.com/basicus/hla-course/migrations"
	"github.com/basicus/hla-course/storage"
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

	conn, err := sqlx.Open("mysql", cfg.DSN)
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
		connRo, err = sqlx.Open("mysql", cfg.DSNro)
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
