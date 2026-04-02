package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"react-go-admin/app/internal/core"
	httprouter "react-go-admin/app/internal/http/router"
	"react-go-admin/app/internal/initialize"
	"react-go-admin/app/internal/migrate"
	"react-go-admin/app/internal/seed"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	runtime, err := initialize.InitRuntime()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer runtime.Close()

	if !runtime.Config.DisableAutoMigrate {
		if err := migrate.Up(ctx, runtime.DB); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := seed.Run(ctx, runtime.Container); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		runtime.Logger.Info("auto-migrate disabled, skipping bootstrap")
	}

	engine := httprouter.New(runtime.Container)
	address := fmt.Sprintf("%s:%d", runtime.Config.Host, runtime.Config.Port)
	runtime.Logger.Info("starting server", "addr", address)

	if err := core.RunServer(ctx, runtime.Logger, address, engine); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
