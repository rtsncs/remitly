package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/rtsncs/remitly-swift-api/database"
	"github.com/rtsncs/remitly-swift-api/handler"
)

func Start() {
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e := echo.New()
	e.Logger.SetLevel(log.INFO)
	db := database.Connect(context.Background())
	defer db.Close()
	h := handler.New(&db)
	e.Logger.Info("Connected to the database")

	g := e.Group("/v1/swift-codes")
	g.GET("/:code", h.GetCode)

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
