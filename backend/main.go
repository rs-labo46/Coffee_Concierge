package main

import (
	"coffee-spa/config"
	"coffee-spa/controller"
	"coffee-spa/db"
	"coffee-spa/gemini"
	"coffee-spa/repository"
	"coffee-spa/router"
	"coffee-spa/usecase"
	"coffee-spa/validator"
	"context"
	"log"
	"time"

	"github.com/labstack/echo/v4"
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
		if err := db.SeedDev(d.G, c.SeedAdminEmail, c.SeedAdminPassword); err != nil {
			log.Fatal(err)
		}
	}

	rdb := db.NewRedis(c)

	if err := db.PingRedis(context.Background(), rdb); err != nil {
		log.Fatal(err)
	}

	e := echo.New()

	clock := usecase.NewRealClock()
	idGen := usecase.NewRandomIDGen()
	tokenSvc := usecase.NewJWTMaker(c.JWTSecret)

	userRepo := repository.NewUserRepository(d.G)
	emailVerifyRepo := repository.NewEmailVerifyRepository(d.G)
	pwResetRepo := repository.NewPwResetRepository(d.G)
	rtRepo := repository.NewRtRepository(d.G)
	auditRepo := repository.NewAuditRepository(d.G)
	sourceRepo := repository.NewSourceRepository(d.G)
	itemRepo := repository.NewItemRepository(d.G)
	beanRepo := repository.NewBeanRepository(d.G)
	recipeRepo := repository.NewRecipeRepository(d.G)
	sessionRepo := repository.NewSessionRepository(d.G)
	savedRepo := repository.NewSavedRepository(d.G)
	rateLimitRepo := repository.NewRateLimitRepository(rdb)

	authVal := validator.NewAuthValidator()
	itemVal := validator.NewItemValidator()
	sourceVal := validator.NewSourceValidator()
	searchVal := validator.NewSearchValidator()
	beanVal := validator.NewBeanValidator()
	recipeVal := validator.NewRecipeValidator()
	savedVal := validator.NewSavedValidator()
	auditVal := validator.NewAuditValidator()

	rlUC := usecase.NewRateLimitUC(rateLimitRepo, clock)

	authUC := usecase.NewAuthUsecase(
		userRepo,
		emailVerifyRepo,
		pwResetRepo,
		rtRepo,
		auditRepo,
		authVal,
		usecase.NewBcryptHasher(),
		tokenSvc,
		tokenSvc,
		clock,
		idGen,
		usecase.NewLogMailer(c.FEURL),
		24*time.Hour,
		30*time.Minute,
		7*24*time.Hour,
	)

	itemUC := usecase.NewItemUsecase(itemRepo, sourceRepo, auditRepo, itemVal)
	sourceUC := usecase.NewSourceUsecase(sourceRepo, auditRepo, sourceVal)
	beanUC := usecase.NewBeanUsecase(beanRepo, auditRepo, beanVal)
	recipeUC := usecase.NewRecipeUsecase(recipeRepo, beanRepo, auditRepo, recipeVal)
	geminiClient, err := gemini.NewClient(c)
	if err != nil {
		log.Fatal(err)
	}

	if c.GeminiUseMock {
		log.Println("[GEMINI] using mock client")
	} else {
		log.Println("[GEMINI] using real client model=", c.GeminiModel)
	}

	searchFlowUC := usecase.NewSearchFlowUsecase(
		sessionRepo,
		beanRepo,
		recipeRepo,
		itemRepo,
		auditRepo,
		searchVal,
		usecase.NewCoffeeRanker(),
		geminiClient,
		clock,
		idGen,
		24*time.Hour,
	)

	sessionUC := usecase.NewSessionUsecase(sessionRepo, auditRepo, searchVal, clock)
	savedUC := usecase.NewSavedUsecase(savedRepo, sessionRepo, auditRepo, savedVal)
	auditUC := usecase.NewAuditUsecase(auditRepo, auditVal)
	healthUC := usecase.NewHealthUC(d.S)

	healthCtl := controller.NewHealthCtl(healthUC)
	authCtl := controller.NewAuthCtl(authUC, rlUC)
	itemCtl := controller.NewItemCtl(itemUC)
	sourceCtl := controller.NewSourceCtl(sourceUC)
	beanCtl := controller.NewBeanCtl(beanUC)
	recipeCtl := controller.NewRecipeCtl(recipeUC)
	searchCtl := controller.NewSearchCtl(searchFlowUC, sessionUC)
	savedCtl := controller.NewSavedCtl(savedUC)
	auditCtl := controller.NewAuditCtl(auditUC)
	wsCtl := controller.NewWsCtl(searchFlowUC, sessionUC)

	router.New(
		e,
		healthCtl,
		authCtl,
		itemCtl,
		sourceCtl,
		beanCtl,
		recipeCtl,
		searchCtl,
		savedCtl,
		auditCtl,
		wsCtl,
		c.JWTSecret,
		userRepo,
		c.FEURL,
		rlUC,
		rlUC,
	)

	log.Fatal(e.Start(":" + c.Port))
}
