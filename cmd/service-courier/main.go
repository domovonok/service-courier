package main

import (
	"context"
	"errors"
	"log"
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
	"github.com/Avito-courses/course-go-avito-domovonok/internal/logger"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/middleware"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/repository/postgres"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/router"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/service"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/worker"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalln("Failed to initialize logger:", err)
	}
	defer zapLogger.Sync()

	appLogger := logger.NewZapLogger(zapLogger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	middleware.StartSystemMetricsCollector(ctx)

	pool, err := database.NewPool(ctx, cfg.DB)
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

	orderGateway, err := gateway.NewOrderGateway(cfg.Order.ServiceHost)
	if err != nil {
		log.Fatalln("Failed to initialize order gateway:", err)
	}
	defer orderGateway.Close()

	orderWorker := worker.NewOrderWorker(orderGateway, deliveryService, cfg.Order.CheckInterval)
	go orderWorker.Run(ctx)

	httpHandler := router.New(courierHandler, deliveryHandler, appLogger)

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
