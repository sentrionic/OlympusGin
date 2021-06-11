package database

import (
	"fmt"
	"github.com/sentrionic/OlympusGin/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/url"
)

type Connection interface {
	Get() *gorm.DB
}

type databaseConnection struct {
	DB *gorm.DB
}

func NewDatabaseConnection(c *config.Config) Connection {
	cfg := c.Get()
	user := cfg.GetString("db.username")
	password := cfg.GetString("db.password")
	database := cfg.GetString("db.database")
	host := cfg.GetString("db.host")
	port := cfg.GetInt("db.port")

	var enableLogging logger.Interface
	if cfg.GetBool("db.log") {
		enableLogging = logger.Default
	}

	dsn := url.URL{
		User:     url.UserPassword(user, password),
		Scheme:   "postgres",
		Host:     fmt.Sprintf("%s:%d", host, port),
		Path:     database,
		RawQuery: (&url.Values{"sslmode": []string{"disable"}}).Encode(),
	}

	db, err := gorm.Open(postgres.Open(dsn.String()), &gorm.Config{
		Logger: enableLogging,
	})
	if err != nil {
		panic("database connection failed")
	}

	if cfg.GetBool("db.sync") {
		if err := synchronize(db); err != nil {
			fmt.Printf("database synchronization failed: %v", err)
		}
	}

	return &databaseConnection{DB: db}
}

func (d *databaseConnection) Get() *gorm.DB {
	return d.DB
}
