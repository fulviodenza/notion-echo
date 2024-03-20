package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo"
	"github.com/notion-echo/bot"
)

func main() {
	var err error
	ctx := context.Background()

	botWithConfig, err := bot.NewBotWithConfig()
	if err != nil {
		log.Fatalf("got error: %v", err)
	}

	err = botWithConfig.TelegramClient.Run(false)
	if err != nil {
		log.Fatalf("got error: %v", err)
	}
	go botWithConfig.Start(ctx)
	go startOauth2Endpoint()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

func startOauth2Endpoint() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.Logger.Fatal(e.Start(":8080"))
}
