package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strings"
	"time"

	"coffee-spa/config"
	"coffee-spa/controller"
	"coffee-spa/db"
	"coffee-spa/gemini"
	"coffee-spa/middleware"
	"coffee-spa/repository"
	"coffee-spa/router"
	"coffee-spa/usecase"
	"coffee-spa/usecase/port"
	"coffee-spa/validator"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// 依存関係を1箇所で組み立てるComposition Rootとして扱う。
func main() {
	// configを読み込む。
	c, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// DB接続を開く。
	d, err := db.Open(c)
	if err != nil {
		log.Fatal(err)
	}

	// 起動時にmigrationを。
	if err := db.Migrate(d); err != nil {
		log.Fatal(err)
	}

	// dev環境ではseedを投入。
	if c.GoEnv == "dev" {
		if err := db.SeedDev(
			d.G,
			c.SeedAdminEmail,
			c.SeedAdminPassword,
		); err != nil {
			log.Fatal(err)
		}
	}

	// Echo本体を生成する。
	e := echo.New()

	// Redisクライアントを生成する。
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr(),
	})

	// 起動時にRedisの疎通だけ確認する。
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal(err)
	}

	// repositoryを生成。
	repos := newRepositories(d.G, rdb)

	// validatorを生成。
	vals := newValidators()

	// 認証/共通serviceを生成。
	svcs := newAppServices(c)

	// usecaseを生成。
	ucs := newUsecases(d.S, repos, vals, svcs)

	// controllerを生成。
	ctls := newControllers(ucs)

	// routerへ全controllerを配線。
	router.New(
		e,
		ctls.healthCtl,
		ctls.authCtl,
		ctls.itemCtl,
		ctls.sourceCtl,
		ctls.beanCtl,
		ctls.recipeCtl,
		ctls.searchCtl,
		ctls.savedCtl,
		ctls.auditCtl,
		ctls.wsCtl,
		c.JWTSecret,
		repos.userRepo,
		c.FEURL,
		ucs.wsLimiter,
		ucs.rlUC,
	)

	log.Fatal(e.Start(":" + c.Port))
}

// main内で生成したrepositoryをまとめる。
type appRepositories struct {
	userRepo    port.UserRepository
	evRepo      port.EmailVerifyRepository
	pwRepo      port.PwResetRepository
	rtRepo      port.RtRepository
	sourceRepo  port.SourceRepository
	itemRepo    port.ItemRepository
	auditRepo   port.AuditRepository
	sessionRepo port.SessionRepository
	beanRepo    port.BeanRepository
	recipeRepo  port.RecipeRepository
	savedRepo   port.SavedRepository
	rlStore     port.RateLimitStore
}

// DB / Redisからrepositoryをまとめて生成。
// repository.GormDBではなく、実体の*gorm.DBを受ける。
func newRepositories(
	gdb *gorm.DB,
	rdb *redis.Client,
) appRepositories {
	return appRepositories{
		userRepo:    repository.NewUserRepository(gdb),
		evRepo:      repository.NewEvRepository(gdb),
		pwRepo:      repository.NewPwRepository(gdb),
		rtRepo:      repository.NewRtRepository(gdb),
		sourceRepo:  repository.NewSourceRepository(gdb),
		itemRepo:    repository.NewItemRepository(gdb),
		auditRepo:   repository.NewAuditRepository(gdb),
		sessionRepo: repository.NewSessionRepository(gdb),
		beanRepo:    repository.NewBeanRepository(gdb),
		recipeRepo:  repository.NewRecipeRepository(gdb),
		savedRepo:   repository.NewSavedRepository(gdb),
		rlStore:     repository.NewRateLimitStore(rdb),
	}
}

// usecaseへ注入するvalidator。
type appValidators struct {
	authVal   usecase.AuthVal
	itemVal   usecase.ItemVal
	sourceVal usecase.SourceVal
	searchVal usecase.SearchVal
	beanVal   usecase.BeanVal
	recipeVal usecase.RecipeVal
	savedVal  usecase.SavedVal
	auditVal  usecase.AuditVal
}

// 各validatorを生成。
func newValidators() appValidators {
	return appValidators{
		authVal:   validator.NewAuthValidator(),
		itemVal:   validator.NewItemValidator(),
		sourceVal: validator.NewSourceValidator(),
		searchVal: validator.NewSearchValidator(),
		beanVal:   validator.NewBeanValidator(),
		recipeVal: validator.NewRecipeValidator(),
		savedVal:  validator.NewSavedValidator(),
		auditVal:  validator.NewAuditValidator(),
	}
}

// usecase へ注入するservice。
type appServices struct {
	pwHasher usecase.PwHasher
	tokenSvc usecase.TokenSvc
	rtSvc    usecase.RefreshSvc
	clock    usecase.Clock
	idGen    usecase.IDGen
	mailer   usecase.Mailer

	geminiClient usecase.GeminiClient
}

// 共通serviceを生成。
func newAppServices(c config.Cfg) appServices {
	tk := usecase.NewJWTMaker(c.JWTSecret)

	return appServices{
		pwHasher:     usecase.NewBcryptHasher(),
		tokenSvc:     tk,
		rtSvc:        tk,
		clock:        usecase.NewRealClock(),
		idGen:        usecase.NewRandomIDGen(),
		mailer:       usecase.NewLogMailer(c.FEURL),
		geminiClient: newGeminiClient(),
	}
}

// 生成済みusecase。
type appUsecases struct {
	healthUC     usecase.HealthUsecase
	authUC       usecase.AuthUC
	itemUC       usecase.ItemUC
	sourceUC     usecase.SourceUC
	beanUC       usecase.BeanUC
	recipeUC     usecase.RecipeUC
	searchFlowUC usecase.SearchFlowUC
	sessionUC    usecase.SessionUC
	savedUC      usecase.SavedUC
	auditUC      usecase.AuditUC

	rlUC      usecase.RateLimiter
	wsLimiter middleware.WsRateLimiter
}

// repository / validator / service からusecase。
func newUsecases(
	sqlDB *sql.DB,
	repos appRepositories,
	vals appValidators,
	svcs appServices,
) appUsecases {
	// 認証前後API用のrate limitルールを定義。
	signupIPRule := usecase.RateRule{
		Rate:     0.1,
		Capacity: 3,
		Cost:     1,
	}
	loginIPRule := usecase.RateRule{
		Rate:     0.2,
		Capacity: 10,
		Cost:     1,
	}
	loginMailRule := usecase.RateRule{
		Rate:     0.1,
		Capacity: 5,
		Cost:     1,
	}
	refreshTokenRule := usecase.RateRule{
		Rate:     0.2,
		Capacity: 5,
		Cost:     1,
	}
	resendIPRule := usecase.RateRule{
		Rate:     0.05,
		Capacity: 2,
		Cost:     1,
	}
	resendMailRule := usecase.RateRule{
		Rate:     0.05,
		Capacity: 2,
		Cost:     1,
	}
	forgotIPRule := usecase.RateRule{
		Rate:     0.05,
		Capacity: 3,
		Cost:     1,
	}
	forgotMailRule := usecase.RateRule{
		Rate:     0.05,
		Capacity: 3,
		Cost:     1,
	}
	wsConnectRule := usecase.RateRule{
		Rate:     0.2,
		Capacity: 5,
		Cost:     1,
	}

	// 認証系とWS系で共通のrate limiter を生成。
	rlUC := usecase.NewRateLimitUC(
		repos.rlStore,
		svcs.clock,
		signupIPRule,
		loginIPRule,
		loginMailRule,
		refreshTokenRule,
		resendIPRule,
		resendMailRule,
		forgotIPRule,
		forgotMailRule,
		wsConnectRule,
	)

	// GeminiClient中心の新シグネチャで生成。
	searchFlowUC := usecase.NewSearchFlowUsecase(
		repos.sessionRepo,
		repos.beanRepo,
		repos.recipeRepo,
		repos.itemRepo,
		repos.auditRepo,
		vals.searchVal,
		usecase.NewCoffeeRanker(),
		svcs.geminiClient,
		svcs.clock,
		svcs.idGen,
		24*time.Hour,
	)

	out := appUsecases{
		healthUC: usecase.NewHealthUC(sqlDB),

		authUC: usecase.NewAuthUsecase(
			repos.userRepo,
			repos.evRepo,
			repos.pwRepo,
			repos.rtRepo,
			repos.auditRepo,
			vals.authVal,
			svcs.pwHasher,
			svcs.tokenSvc,
			svcs.rtSvc,
			svcs.clock,
			svcs.idGen,
			svcs.mailer,
			24*time.Hour,
			30*time.Minute,
			7*24*time.Hour,
		),

		itemUC: usecase.NewItemUsecase(
			repos.itemRepo,
			repos.sourceRepo,
			repos.auditRepo,
			vals.itemVal,
		),
		sourceUC: usecase.NewSourceUsecase(
			repos.sourceRepo,
			repos.auditRepo,
			vals.sourceVal,
		),
		beanUC: usecase.NewBeanUsecase(
			repos.beanRepo,
			repos.auditRepo,
			vals.beanVal,
		),
		recipeUC: usecase.NewRecipeUsecase(
			repos.recipeRepo,
			repos.beanRepo,
			repos.auditRepo,
			vals.recipeVal,
		),

		// SearchFlowUCを実際に生成。
		searchFlowUC: searchFlowUC,

		sessionUC: usecase.NewSessionUsecase(
			repos.sessionRepo,
			repos.auditRepo,
			vals.searchVal,
			svcs.clock,
		),
		savedUC: usecase.NewSavedUsecase(
			repos.savedRepo,
			repos.sessionRepo,
			repos.auditRepo,
			vals.savedVal,
		),
		auditUC: usecase.NewAuditUsecase(
			repos.auditRepo,
			vals.auditVal,
		),

		// AuthCtl用のrate limiter。
		rlUC: rlUC,

		// WsRateLimit middleware用のrate limiter。
		wsLimiter: rlUC,
	}

	return out
}

// controller。
type appControllers struct {
	healthCtl controller.HealthCtl
	authCtl   *controller.AuthCtl
	itemCtl   *controller.ItemCtl
	sourceCtl *controller.SourceCtl
	beanCtl   *controller.BeanCtl
	recipeCtl *controller.RecipeCtl
	searchCtl *controller.SearchCtl
	savedCtl  *controller.SavedCtl
	auditCtl  *controller.AuditCtl
	wsCtl     *controller.WsCtl
}

// usecaseからcontrollerを生成。
func newControllers(ucs appUsecases) appControllers {
	return appControllers{
		healthCtl: controller.NewHealthCtl(ucs.healthUC),
		authCtl:   controller.NewAuthCtl(ucs.authUC, ucs.rlUC),
		itemCtl:   controller.NewItemCtl(ucs.itemUC),
		sourceCtl: controller.NewSourceCtl(ucs.sourceUC),
		beanCtl:   controller.NewBeanCtl(ucs.beanUC),
		recipeCtl: controller.NewRecipeCtl(ucs.recipeUC),
		searchCtl: controller.NewSearchCtl(ucs.searchFlowUC, ucs.sessionUC),
		savedCtl:  controller.NewSavedCtl(ucs.savedUC),
		auditCtl:  controller.NewAuditCtl(ucs.auditUC),
		wsCtl:     controller.NewWsCtl(ucs.searchFlowUC, ucs.sessionUC),
	}
}

// Gemini本実装を生成し、失敗時はmockへフォールバックする。
func newGeminiClient() usecase.GeminiClient {
	svc, err := gemini.NewService()
	if err == nil {
		return svc
	}
	return gemini.NewMockClient()
}

// Redis接続先をenvから組み立てる。
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
