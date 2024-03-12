package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/sergeizaitcev/gophkeeper/internal/router"
	"github.com/sergeizaitcev/gophkeeper/internal/router/mongodb"
)

var (
	flagAddr     string // Адрес http-сервера.
	flagCertFile string // Путь к файлу server.crt.
	flagKeyFile  string // Путь к файлу server.key.
	flagDatabase string // Подключение к mongodb.
	flagDebug    bool   // Режим дебага.
)

func main() {
	var level slog.Level

	flag.StringVar(&flagAddr, "http", ":8433", "http server address")
	flag.StringVar(&flagCertFile, "cert", "server.crt", "path to cert file")
	flag.StringVar(&flagKeyFile, "key", "server.key", "path to key file")
	flag.StringVar(&flagDatabase, "database", "mongodb://127.0.0.1:27017", "connect to mongodb")
	flag.BoolVar(&flagDebug, "debug", false, "debug mode")

	flag.Parse()

	if flagDebug {
		level = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			AddSource: flagDebug,
			Level:     level,
		},
	)))

	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(flagDatabase))
	if err != nil {
		return err
	}
	defer func() {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer disconnectCancel()
		_ = client.Disconnect(disconnectCtx)
	}()

	pingCtx, pingCancel := context.WithTimeout(ctx, 3*time.Second)
	defer pingCancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		return err
	}

	storage := mongodb.New(client.Database("gophkeeper"))
	if err := storage.MigrateIndex(ctx); err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              flagAddr,
		Handler:           router.New(storage, slog.Default()),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Minute,
		IdleTimeout:       time.Minute,
	}

	errc := make(chan error, 1)
	go func() {
		slog.Info("server running on port " + flagAddr)
		errc <- srv.ListenAndServeTLS(flagCertFile, flagKeyFile)
	}()

	select {
	case <-ctx.Done():
		fmt.Println()
	case err := <-errc:
		return err
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer shutdownCancel()

	return srv.Shutdown(shutdownCtx)
}
