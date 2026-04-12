package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"coffee-spa/config"
	"coffee-spa/controller"
	"coffee-spa/db"
	"coffee-spa/repository"
	"coffee-spa/router"
	"coffee-spa/usecase"
	"coffee-spa/validator"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

func main() {
	c, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	d, err := db.Open(c)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Migrate(d); err != nil {
		log.Fatal(err)
	}

	if c.GoEnv == "dev" {
		if err := db.SeedDev(
			d.G,
			c.SeedAdminEmail,
			c.SeedAdminPassword,
		); err != nil {
			log.Fatal(err)
		}
	}

	e := echo.New()

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr(),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal(err)
	}

	userRepo := repository.NewUserRepository(d.G)
	evRepo := repository.NewEvRepository(d.G)
	pwRepo := repository.NewPwRepository(d.G)
	rtRepo := repository.NewRtRepository(d.G)
	sourceRepo := repository.NewSourceRepository(d.G)
	itemRepo := repository.NewItemRepository(d.G)
	auditRepo := repository.NewAuditRepository(d.G)

	authVal := validator.NewAuthValidator()
	itemVal := validator.NewItemValidator()
	sourceVal := validator.NewSourceValidator()

	ph := usecase.NewBcryptHasher()
	tk := usecase.NewJWTMaker(c.JWTSecret)
	clock := usecase.NewRealClock()
	idGen := usecase.NewRandomIDGen()
	mail := usecase.NewLogMailer(c.FEURL)

	healthUC := usecase.NewHealthUC(d.S)
	authUC := usecase.NewAuthUsecase(
		userRepo,
		evRepo,
		pwRepo,
		rtRepo,
		auditRepo,
		authVal,
		ph,
		tk,
		tk,
		clock,
		idGen,
		mail,
		24*time.Hour,
		30*time.Minute,
		7*24*time.Hour,
	)
	itemUC := usecase.NewItemUsecase(
		itemRepo,
		sourceRepo,
		auditRepo,
		itemVal,
	)
	sourceUC := usecase.NewSourceUsecase(
		sourceRepo,
		auditRepo,
		sourceVal,
	)

	healthCtl := controller.NewHealthCtl(healthUC)
	authCtl := controller.NewAuthCtl(authUC)
	itemCtl := controller.NewItemCtl(itemUC)
	sourceCtl := controller.NewSourceCtl(sourceUC)

	router.New(
		e,
		healthCtl,
		authCtl,
		itemCtl,
		sourceCtl,
		c.JWTSecret,
		userRepo,
		c.FEURL,
	)

	log.Fatal(e.Start(":" + c.Port))
}

func redisAddr() string {
	addr := strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	if addr != "" {
		return addr
	}

	host := strings.TrimSpace(os.Getenv("REDIS_HOST"))
	if host == "" {
		host = "localhost"
	}

	port := strings.TrimSpace(os.Getenv("REDIS_PORT"))
	if port == "" {
		port = "6379"
	}

	return host + ":" + port
}
