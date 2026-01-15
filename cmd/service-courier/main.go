package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/config"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/database"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/factory"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/gateway"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/handler"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/limiter"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/logger"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/metrics"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/repository/postgres"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/router"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/service"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/worker"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	zapLogger := zap.Must(zap.NewProduction())

	appLogger := logger.NewZapLogger(zapLogger)
	defer func() {
		if err := appLogger.Sync(); err != nil {
			appLogger.Error("Unable to sync logger", logger.Error(err))
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	prom := metrics.NewPrometheusMetrics()
	metrics.StartSystemMetricsCollector(ctx, prom)

	pool, err := database.NewPool(ctx, cfg.DB, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize database:", logger.Error(err))
	}
	defer pool.Close()

	courierRepo := postgres.NewCourierRepository(pool)
	courierService := service.NewCourierService(courierRepo)
	courierHandler := handler.NewCourierHandler(courierService, appLogger)

	deliveryRepo := postgres.NewDeliveryRepository(pool)
	calculatorFactory := factory.NewDeliveryTimeCalculatorFactory()
	txManager := postgres.NewTransactionManager(pool)
	deliveryService := service.NewDeliveryService(deliveryRepo, courierRepo, calculatorFactory, txManager, appLogger)
	deliveryHandler := handler.NewDeliveryHandler(deliveryService, appLogger)

	expirationWorker := worker.NewDeliveryExpirationWorker(deliveryService, cfg.DeliveryCheckInterval, appLogger)
	go expirationWorker.Start(ctx)

	orderGateway, err := gateway.NewOrderGateway(cfg.Order.ServiceHost, cfg.Order.MaxRetries, cfg.Order.RetryDelay, prom)
	if err != nil {
		appLogger.Fatal("Failed to initialize order gateway:", logger.Error(err))
	}
	defer func() {
		if err := orderGateway.Close(); err != nil {
			appLogger.Error("Failed to close order gateway:", logger.Error(err))
		}
	}()

	orderWorker := worker.NewOrderWorker(orderGateway, deliveryService, cfg.Order.CheckInterval, appLogger)
	go orderWorker.Run(ctx)

	rateLimiter := limiter.NewTokenBucket(cfg.RateLimit.Capacity, cfg.RateLimit.RefillRate)

	httpHandler := router.New(courierHandler, deliveryHandler, appLogger, prom, rateLimiter)

	srv := &http.Server{
		Addr:    net.JoinHostPort("", cfg.Port),
		Handler: httpHandler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Fatal("Server error:", logger.Error(err))
		}
	}()
	appLogger.Info("Server listening on", logger.Any("addr", srv.Addr))

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	appLogger.Info("Shutting down service-courier...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Fatal("Graceful shutdown failed:", logger.Error(err))
	}
	appLogger.Info("Service stopped successfully")
}
