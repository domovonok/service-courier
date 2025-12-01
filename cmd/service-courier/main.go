package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/factory"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/router"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/worker"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/config"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/handler"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/repository/postgres"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/service"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := initDB(ctx, cfg.DB)
	if err != nil {
		log.Fatalln("Failed to initialize database:", err)
	}
	defer pool.Close()

	courierRepo := postgres.NewCourierRepository(pool)
	courierService := service.NewCourierService(courierRepo)
	courierHandler := handler.NewCourierHandler(courierService)

	deliveryRepo := postgres.NewDeliveryRepository(pool)
	calculatorFactory := factory.NewDeliveryTimeCalculatorFactory()
	txManager := postgres.NewTransactionManager(pool)
	deliveryService := service.NewDeliveryService(deliveryRepo, courierRepo, calculatorFactory, txManager)
	deliveryHandler := handler.NewDeliveryHandler(deliveryService)

	expirationWorker := worker.NewDeliveryExpirationWorker(deliveryService, cfg.DeliveryCheckInterval)
	go expirationWorker.Start(ctx)

	httpHandler := router.New(courierHandler, deliveryHandler)

	srv := &http.Server{
		Addr:    net.JoinHostPort("", cfg.Port),
		Handler: httpHandler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalln("Server error:", err)
		}
	}()
	log.Println("Server listening on", srv.Addr)

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Shutting down service-courier...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalln("Graceful shutdown failed:", err)
	}
	log.Println("Service stopped successfully")
}

func initDB(ctx context.Context, cfg config.DBConfig) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.PgUser, cfg.PgPassword, cfg.PgHost, cfg.PgPort, cfg.PgDB)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("invalid connection string: %w", err)
	}

	poolConfig.MaxConns = cfg.Pool.MaxConns
	poolConfig.MinConns = cfg.Pool.MinConns
	poolConfig.MinIdleConns = cfg.Pool.MinIdleConns
	poolConfig.HealthCheckPeriod = cfg.Pool.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	for i := 1; i <= cfg.Pool.PingMaxRetries; i++ {
		if err := pool.Ping(ctx); err == nil {
			log.Println("Successfully connected to database")
			return pool, nil
		} else {
			log.Printf("Database ping attempt %d/%d failed: %v", i, cfg.Pool.PingMaxRetries, err)
			if i < cfg.Pool.PingMaxRetries {
				log.Printf("Retrying in %v...", cfg.Pool.PingRetryDelay)
				time.Sleep(cfg.Pool.PingRetryDelay)
			}
		}
	}

	pool.Close()
	return nil, fmt.Errorf("unable to ping database")
}
