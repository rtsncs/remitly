package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/rtsncs/remitly-swift-api/database"
	"github.com/rtsncs/remitly-swift-api/handler"
)

func Run() {
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e := echo.New()
	e.Logger.SetLevel(log.INFO)
	db, err := database.Connect(context.Background())
	if err != nil {
		e.Logger.Fatalf("Failed to connect to the database")
	}
	defer db.Close()
	h := handler.New(&db)
	e.Logger.Info("Connected to the database")
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	g := e.Group("/v1/swift-codes")
	g.GET("/:code", h.GetCode)
	g.GET("/country/:countryCode", h.GetByCountryCode)
	g.POST("", h.AddCode)
	g.DELETE("/:code", h.DeleteCode)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		if err := e.Start(host + ":" + port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatalf("Server error: %v\n", err)
		}
	}()

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	e.Logger.Info("Shutdown signal received")
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatalf("Error during shutdown: %v\n", err)
	} else {
		e.Logger.Info("Shutting down the server")
	}
}
