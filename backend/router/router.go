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
	authCtl controller.AuthCtl,
	itemCtl controller.ItemCtl,
	srcCtl controller.SrcCtl,
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
	pub.POST("/auth/signup", authCtl.Signup)
	pub.POST("/auth/verify-email", authCtl.VerifyEmail)
	pub.POST("/auth/resend-verify", authCtl.ResendVerify)
	pub.POST("/auth/login", authCtl.Login)
	pub.POST("/auth/password/forgot", authCtl.ForgotPw)
	pub.POST("/auth/password/reset", authCtl.ResetPw)
	pub.GET("/items/top", itemCtl.Top)
	pub.GET("/items/:id", itemCtl.Detail)
	pub.GET("/items", itemCtl.List)
	pub.GET("/sources", srcCtl.List)

	//csrf required
	//refreshはcookie+csrfが必要
	csrf := e.Group("")
	csrf.Use(middleware.CSRF())
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
	admin.Use(middleware.AdminOnly())
	admin.POST("/items", itemCtl.Create)
	admin.POST("/sources", srcCtl.Create)
}
