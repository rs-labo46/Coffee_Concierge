package db

import (
	"coffee-spa/config"
	"database/sql"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	G *gorm.DB // GORM本体
	S *sql.DB
}

func Open(c config.Cfg) (DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Tokyo",
		c.PgHost, c.PgPort, c.PgUser, c.PgPass, c.PgDB,
	)

	g, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}) // GORM接続
	if err != nil {
		return DB{}, err
	}

	s, err := g.DB()
	if err != nil {
		return DB{}, err
	}

	return DB{G: g, S: s}, nil
}
