package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"coffee-spa/db"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// envRequiredは、必須環境変数を読み取る。
// 空の場合はseedを止める。
func envRequired(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		log.Fatalf("missing env: %s", key)
	}

	return value
}

// envDefaultは、任意環境変数を読み取る。
// 空の場合はdefaultValueを返す。
func envDefault(key string, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	return value
}

// buildPostgresDSNは、GitHub ActionsやローカルからPostgresへ接続するDSNを作る。
// Render外部接続ではsslmode=requireを使う。
func buildPostgresDSN() string {
	user := envRequired("POSTGRES_USER")
	password := envRequired("POSTGRES_PASSWORD")
	dbName := envRequired("POSTGRES_DB")
	host := envRequired("POSTGRES_HOST")
	port := envDefault("POSTGRES_PORT", "5432")
	sslMode := envDefault("POSTGRES_SSLMODE", "require")

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   net.JoinHostPort(host, port),
		Path:   "/" + dbName,
	}

	q := u.Query()
	q.Set("sslmode", sslMode)
	u.RawQuery = q.Encode()

	return u.String()
}

func main() {
	adminEmail := envRequired("SEED_ADMIN_EMAIL")
	adminPassword := envRequired("SEED_ADMIN_PASSWORD")

	dsn := buildPostgresDSN()

	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(fmt.Errorf("open postgres driver: %w", err))
	}
	defer sqlDB.Close()

	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatal(fmt.Errorf("ping postgres: %w", err))
	}

	g, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		log.Fatal(fmt.Errorf("open gorm: %w", err))
	}

	if err := db.SeedDev(g, adminEmail, adminPassword); err != nil {
		log.Fatal(fmt.Errorf("seed dev data: %w", err))
	}

	log.Println("seed completed")
}
