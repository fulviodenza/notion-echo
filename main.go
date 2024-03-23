package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	go botWithConfig.RunOauth2Endpoint()
	go botWithConfig.Start(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
