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
	"coffee-spa/policy"
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

	pwPol := policy.NewPwPol()
	emailPol := policy.NewEmailPol()
	kindPol := policy.NewKindPol()
	urlPol := policy.NewURLPol()
	pagePol := policy.NewPagePol()

	authVal := validator.NewAuthValidator(emailPol, pwPol)
	itemVal := validator.NewItemValidator(kindPol, urlPol, pagePol)
	sourceVal := validator.NewSourceValidator(urlPol)

	ph := usecase.NewBcryptHasher()
	tk := usecase.NewJWTMaker(c.JWTSecret)
	mail := usecase.NewLogMailer(c.FEURL)

	rlStore := repository.NewRateLimitStore(rdb)
	rl := usecase.NewRateLimitUC(
		rlStore,
		usecase.RateRule{Limit: 1, Window: 5 * time.Second},
		usecase.RateRule{Limit: 1, Window: 5 * time.Second},
		usecase.RateRule{Limit: 1, Window: 2 * time.Second},
		usecase.RateRule{Limit: 1, Window: 2 * time.Second},
		usecase.RateRule{Limit: 1, Window: 4 * time.Second},
		usecase.RateRule{Limit: 1, Window: 2 * time.Second},
		usecase.RateRule{Limit: 1, Window: 4 * time.Second},
	)

	healthUC := usecase.NewHealthUC(d.S)
	authUC := usecase.NewAuthUC(
		userRepo,
		evRepo,
		pwRepo,
		rtRepo,
		auditRepo,
		authVal,
		ph,
		tk,
		mail,
		rl,
	)
	itemUC := usecase.NewItemUC(
		itemRepo,
		auditRepo,
		itemVal,
	)
	sourceUC := usecase.NewSourceUC(
		sourceRepo,
		auditRepo,
		sourceVal,
	)

	healthCtl := controller.NewHealthCtl(healthUC)
	authCtl := controller.NewAuthCtl(authUC)
	itemCtl := controller.NewItemCtl(itemUC)
	srcCtl := controller.NewSrcCtl(sourceUC)

	router.New(
		e,
		healthCtl,
		authCtl,
		itemCtl,
		srcCtl,
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
