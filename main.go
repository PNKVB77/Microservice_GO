package main

import (
	"context"
	"fmt"
	"mcrsrv_go/application"
	"os"
	"os/signal"
)

func main() {
	app := application.New()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := app.Start(ctx)
	if err != nil {
		fmt.Println("failed to start an app", err)
	}
}
