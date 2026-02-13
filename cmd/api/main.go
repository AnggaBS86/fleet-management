package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"leet-management/internal/api"
	"leet-management/internal/config"
	"leet-management/internal/db"
	"leet-management/internal/mqtt"
	"leet-management/internal/queue"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, err := db.New(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("db init failed: %v", err)
	}
	defer store.Close()

	publisher, err := queue.NewPublisher(cfg.RabbitURL, cfg.GeofenceEventExchange, cfg.GeofenceQueue, cfg.GeofenceRoutingKey)
	if err != nil {
		log.Fatalf("rabbit init failed: %v", err)
	}
	defer publisher.Close()

	subscriber, err := mqtt.NewSubscriber(cfg.MQTTBroker, cfg.MQTTClientID, cfg.MQTTUsername, cfg.MQTTPassword, store, publisher, cfg.GeofenceLat, cfg.GeofenceLon, cfg.GeofenceRadiusMeters)
	if err != nil {
		log.Fatalf("mqtt init failed: %v", err)
	}
	defer subscriber.Close()

	if err := subscriber.Start(ctx, "/fleet/vehicle/+/location"); err != nil {
		log.Fatalf("mqtt subscribe failed: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	apiServer := api.NewServer(store)
	apiServer.RegisterRoutes(r)

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("api listening on :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
