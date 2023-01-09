package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gospt/internal/api"
	"gospt/internal/auth"
	"gospt/internal/config"
	"gospt/internal/gctx"
)

func init() {
	configPath, _ := os.UserConfigDir()
	configDir := filepath.Join(configPath, "gospt")
	config.LoadConfig(configDir)
}

func main() {
	var err error
	log.New(os.Stdout, "LOG:", 0)
	ctx := gctx.NewContext(context.Background())
	client, err := auth.GetClient(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	currentUser, err := client.CurrentUser(ctx)
	if err != nil {
		panic(err.Error())
	}
	ctx.UserId = currentUser.ID
	err = api.Run(ctx, client, os.Args[1:])
	if err != nil {
		fmt.Println(err)
	}
}
