package router

import (
	"coffee-spa/controller"
	"coffee-spa/middleware"
	"coffee-spa/repository"

	"github.com/labstack/echo/v4"
)

func New(
	e *echo.Echo,
	healthCtl controller.HealthCtl,
	authCtl *controller.AuthCtl,
	itemCtl *controller.ItemCtl,
	srcCtl *controller.SourceCtl,
	jwtSecret string,
	userRepo repository.UserRepository,
	feURL string,
) {
	//全体共通middleware
	e.Use(middleware.CORS(feURL))
	e.Use(middleware.SecurityHeaders())

	//health check
	e.GET("/health", healthCtl.Get)

	//public
	pub := e.Group("")
	pub.GET("/auth/csrf", authCtl.Csrf)
	pub.POST("/auth/signup", authCtl.Signup)
	pub.POST("/auth/verify-email", authCtl.VerifyEmail)
	pub.POST("/auth/login", authCtl.Login)
	pub.POST("/auth/password/forgot", authCtl.ForgotPw)
	pub.POST("/auth/password/reset", authCtl.ResetPw)
	pub.POST("/auth/verify-email/resend", authCtl.ResendVerify)
	pub.GET("/items/top", itemCtl.Top)
	pub.GET("/items/:id", itemCtl.Get)
	pub.GET("/items", itemCtl.List)
	pub.GET("/sources", srcCtl.List)

	//csrf required
	//refreshはcookie+csrfが必要
	csrf := e.Group("")
	csrf.Use(middleware.CsrfCheck())
	csrf.Use(middleware.RequireRefreshCookie())
	csrf.POST("/auth/refresh", authCtl.Refresh)

	//private
	priv := e.Group("")
	priv.Use(middleware.JWTAuth(jwtSecret))
	priv.Use(middleware.TokenVersion(userRepo))
	priv.GET("/me", authCtl.Me)
	priv.POST("/auth/logout", authCtl.Logout)

	//admin
	admin := e.Group("")
	admin.Use(middleware.JWTAuth(jwtSecret))
	admin.Use(middleware.TokenVersion(userRepo))
	admin.Use(middleware.RequireAdmin())
	admin.POST("/items", itemCtl.Create)
	admin.POST("/sources", srcCtl.Create)
}
