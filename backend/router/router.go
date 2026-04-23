package router

import (
	"coffee-spa/controller"
	"coffee-spa/middleware"

	"github.com/labstack/echo/v4"
)

// routerの責務はどのURLに、どのcontrollerとmiddleware を当てるかを定義。
func New(
	e *echo.Echo,
	healthCtl controller.HealthCtl,
	authCtl *controller.AuthCtl,
	itemCtl *controller.ItemCtl,
	srcCtl *controller.SourceCtl,
	beanCtl *controller.BeanCtl,
	recipeCtl *controller.RecipeCtl,
	searchCtl *controller.SearchCtl,
	savedCtl *controller.SavedCtl,
	auditCtl *controller.AuditCtl,
	wsCtl *controller.WsCtl,
	jwtSecret string,
	userRepo middleware.TokenVersionReader,
	feURL string,
	wsLimiter middleware.WsRateLimiter,
) {
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.CORS(feURL))
	e.Use(middleware.SecurityHeaders())

	// ヘルスチェック
	e.GET("/health", healthCtl.Get)

	// 未認証でも通すgroup。
	public := e.Group("")
	registerPublicAuthRoutes(public, authCtl)
	registerPublicContentRoutes(public, itemCtl, srcCtl, beanCtl, recipeCtl)
	registerPublicSearchRoutes(public, searchCtl)

	// CSRFとrefresh cookie前提のため、別groupで保護。
	csrf := e.Group("")
	csrf.Use(middleware.CsrfCheck())
	csrf.Use(middleware.RequireRefreshCookie())
	csrf.POST("/auth/refresh", authCtl.Refresh)

	// JWT認証済みユーザー専用group。
	private := e.Group("")
	private.Use(middleware.JWTAuth(jwtSecret))
	private.Use(middleware.TokenVersion(userRepo))
	registerPrivateRoutes(private, authCtl, searchCtl, savedCtl)

	// privateに加えて管理者権限が必要なgroup。
	admin := e.Group("/admin")
	admin.Use(middleware.JWTAuth(jwtSecret))
	admin.Use(middleware.TokenVersion(userRepo))
	admin.Use(middleware.RequireAdmin())
	registerAdminRoutes(admin, itemCtl, srcCtl, beanCtl, recipeCtl, auditCtl)

	// guestの継続対話用。
	// HTTPと別にupgrade/sessionKey/WS専用rate limit。
	wsGuest := e.Group("/ws")
	wsGuest.Use(middleware.WsUpgradeCheck())
	wsGuest.Use(middleware.SessionKeyQueryRequired())
	wsGuest.Use(middleware.WsRateLimit(wsLimiter))
	wsGuest.GET("/guest/search/sessions/:id", wsCtl.Connect)

	// 認証ユーザーの継続対話用。
	wsPrivate := e.Group("/ws")
	wsPrivate.Use(middleware.WsUpgradeCheck())
	wsPrivate.Use(middleware.JWTAuth(jwtSecret))
	wsPrivate.Use(middleware.TokenVersion(userRepo))
	wsPrivate.Use(middleware.WsRateLimit(wsLimiter))
	wsPrivate.GET("/search/sessions/:id", wsCtl.Connect)
}

// 未認証で使う認証系routeをまとめて登録。
func registerPublicAuthRoutes(
	g *echo.Group,
	authCtl *controller.AuthCtl,
) {
	g.GET("/auth/csrf", authCtl.Csrf)
	g.POST("/auth/signup", authCtl.Signup)
	g.POST("/auth/verify-email", authCtl.VerifyEmail)
	g.POST("/auth/login", authCtl.Login)
	g.POST("/auth/password/forgot", authCtl.ForgotPw)
	g.POST("/auth/password/reset", authCtl.ResetPw)
	g.POST("/auth/verify-email/resend", authCtl.ResendVerify)
}

// 公開コンテンツ取得系routeをまとめる。
// Item/Source/Bean/Recipeは閲覧系だけ publicへ。
func registerPublicContentRoutes(
	g *echo.Group,
	itemCtl *controller.ItemCtl,
	srcCtl *controller.SourceCtl,
	beanCtl *controller.BeanCtl,
	recipeCtl *controller.RecipeCtl,
) {
	g.GET("/items/top", itemCtl.Top)
	g.GET("/items/:id", itemCtl.Get)
	g.GET("/items", itemCtl.List)

	g.GET("/sources", srcCtl.List)
	g.GET("/sources/:id", srcCtl.Get)

	g.GET("/beans", beanCtl.List)
	g.GET("/beans/:id", beanCtl.Get)

	g.GET("/recipes", recipeCtl.List)
	g.GET("/recipes/:id", recipeCtl.Get)
}

// guestを含む検索開始系routeを登録。
// session開始と初回条件設定はHTTPで行う設計に固定。
func registerPublicSearchRoutes(
	g *echo.Group,
	searchCtl *controller.SearchCtl,
) {
	g.POST("/search/sessions", searchCtl.StartSession)
	g.POST("/search/sessions/:id/pref", middleware.SessionKeyHeaderRequired()(searchCtl.SetPref))
	g.PATCH("/search/sessions/:id/pref", middleware.SessionKeyHeaderRequired()(searchCtl.PatchPref))
	g.GET("/search/guest/sessions/:id", middleware.SessionKeyHeaderRequired()(searchCtl.GetGuestSession))
}

// 認証済みユーザー向け routeを登録。
// 履歴一覧、詳細取得、保存済み提案はuser ownership 前提のためprivateに置く。
func registerPrivateRoutes(
	g *echo.Group,
	authCtl *controller.AuthCtl,
	searchCtl *controller.SearchCtl,
	savedCtl *controller.SavedCtl,
) {
	g.GET("/me", authCtl.Me)
	g.POST("/auth/logout", authCtl.Logout)

	g.GET("/search/sessions", searchCtl.ListHistory)
	g.GET("/search/sessions/:id", searchCtl.GetSession)
	g.POST("/search/sessions/:id/close", searchCtl.CloseSession)

	g.POST("/saved-suggestions", savedCtl.Save)
	g.GET("/saved-suggestions", savedCtl.List)
	g.DELETE("/saved-suggestions/:suggestionId", savedCtl.Delete)
}

// 管理者だけが使う更新系routeを登録する。
// create/update/audit閲覧など、公開できない操作。
func registerAdminRoutes(
	g *echo.Group,
	itemCtl *controller.ItemCtl,
	srcCtl *controller.SourceCtl,
	beanCtl *controller.BeanCtl,
	recipeCtl *controller.RecipeCtl,
	auditCtl *controller.AuditCtl,
) {
	g.POST("/items", itemCtl.Create)
	g.POST("/sources", srcCtl.Create)

	g.POST("/beans", beanCtl.Create)
	g.PATCH("/beans/:id", beanCtl.Update)

	g.POST("/recipes", recipeCtl.Create)
	g.PATCH("/recipes/:id", recipeCtl.Update)

	g.GET("/audit-logs", auditCtl.List)
}
