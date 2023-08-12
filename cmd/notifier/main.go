package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sku4/ad-notifier/configs"
	"github.com/sku4/ad-notifier/internal/handler/rest"
	"github.com/sku4/ad-notifier/internal/handler/watcher"
	"github.com/sku4/ad-notifier/internal/repository"
	"github.com/sku4/ad-notifier/internal/sender"
	"github.com/sku4/ad-notifier/internal/service"
	"github.com/sku4/ad-parser/pkg/logger"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/pool"
)

func main() {
	// init config
	log := logger.Get()
	cfg, err := configs.Init()
	if err != nil {
		log.Fatalf("error init config: %s", err)
	}

	// init tarantool
	conn, err := pool.Connect(cfg.Tarantool.Servers, tarantool.Opts{
		Timeout:   cfg.Tarantool.Timeout,
		Reconnect: cfg.Tarantool.ReconnectInterval,
		RequiredProtocolInfo: tarantool.ProtocolInfo{
			Features: []tarantool.ProtocolFeature{tarantool.WatchersFeature},
		},
	})
	if err != nil {
		log.Fatalf("error tarantool connection refused: %s", err)
	}
	defer func() {
		errs := conn.Close()
		for _, e := range errs {
			log.Errorf("error close connection pool: %s", e)
		}
	}()

	// init context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	ctx = configs.Set(ctx, cfg)

	senders, errs := sender.NewSender(ctx)
	if errs != nil {
		log.Fatalf("error init senders: %s", errors.Join(errs...))
	}

	repos := repository.NewRepository(conn)
	services := service.NewService(repos, senders)
	watchers := watcher.NewWatcher(ctx, services)
	handlers := rest.NewHandler()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	err = watchers.Register()
	if err != nil {
		log.Fatalf("error watchers register: %s", err)
	}

	// run rest server
	routes := handlers.InitRoutes()
	restServer := rest.NewServer(ctx, routes)
	go func() {
		log.Info(fmt.Sprintf("Rest server is running on: %d", cfg.Rest.Port))
		if errRest := restServer.ListenAndServe(); errRest != nil {
			log.Infof("Rest server %s", errRest)
			quit <- nil
		}
	}()

	log.Infof("App Started")

	// graceful shutdown
	log.Infof("Got signal %v, attempting graceful shutdown", <-quit)
	cancel()
	log.Info("Context is stopped")

	err = restServer.Shutdown(ctx)
	if err != nil {
		log.Errorf("error rest server shutdown: %s", err)
	}

	err = services.Notifier.Shutdown()
	if err != nil {
		log.Errorf("error notifier shutdown: %s", err)
	} else {
		log.Info("Notifier stopped")
	}

	watchers.Unregister()
	errs = conn.CloseGraceful()
	for _, e := range errs {
		log.Errorf("error close graceful connection pool: %s", e)
	}

	log.Info("App Shutting Down")
}
