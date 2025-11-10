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

	"github.com/Avito-courses/course-go-avito-domovonok/internal/router"
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
	httpHandler := router.New(courierHandler)

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

	poolConfig.MaxConns = 20
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	const maxRetries = 5
	retryDelay := time.Second

	for i := 1; i <= maxRetries; i++ {
		if err := pool.Ping(ctx); err == nil {
			log.Println("Successfully connected to database")
			return pool, nil
		} else {
			log.Printf("Database ping attempt %d/%d failed: %v", i, maxRetries, err)
			if i < maxRetries {
				log.Printf("Retrying in %v...", retryDelay)
				time.Sleep(retryDelay)
			}
		}
	}

	pool.Close()
	return nil, fmt.Errorf("unable to ping database")
}
