package main

import (
	"log"
	"time"

	"github.com/labstack/echo/v4"
	echoMw "github.com/labstack/echo/v4/middleware"
	"github.com/reduan2660/swapenv-server/internal/config"
	"github.com/reduan2660/swapenv-server/internal/db"
	"github.com/reduan2660/swapenv-server/internal/handlers"
	"github.com/reduan2660/swapenv-server/internal/middleware"
	"github.com/reduan2660/swapenv-server/internal/session"
)

func main() {
	cfg := config.Load()

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("DB connection failed: ", err)
	}

	if err := db.AutoMigrate(database); err != nil {
		log.Fatal("Migration Failed: ", err)
	}

	authHandler := &handlers.AuthHandler{
		DB:                 database,
		GithubClientID:     cfg.GithubClientId,
		GithubClientSecret: cfg.GithubClientSecret,
		JWTSecret:          cfg.JWTSecret,
	}

	sessionMgr := session.NewManager(5 * time.Minute)
	shareHandler := &handlers.ShareHandler{
		Sessions: sessionMgr,
	}

	e := echo.New()
	e.Use(echoMw.RequestLogger())
	e.Use(echoMw.Recover())

	e.GET("/health", handlers.Health)
	e.POST("/auth/device/", authHandler.DeviceCode)
	e.POST("/auth/poll/", authHandler.Poll)

	protected := e.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	protected.GET("/share", shareHandler.Share)
	protected.GET("/receive", shareHandler.Receive)

	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
