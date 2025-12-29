package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/reduan2660/swapenv-server/internal/config"
	"github.com/reduan2660/swapenv-server/internal/db"
	"github.com/reduan2660/swapenv-server/internal/handlers"
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

	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.GET("/health", handlers.Health)

	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
