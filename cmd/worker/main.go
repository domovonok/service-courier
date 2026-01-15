package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/config"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/database"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/factory"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/gateway"
	orderChangedHandler "github.com/Avito-courses/course-go-avito-domovonok/internal/handler/queues/order/changed"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/kafka"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/logger"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/metrics"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/repository/postgres"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/service"
	orderChangedService "github.com/Avito-courses/course-go-avito-domovonok/internal/service/order/changed"
	"github.com/IBM/sarama"
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

	sarama.Logger = zap.NewStdLog(zapLogger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	prom := metrics.NewPrometheusMetrics()

	pool, err := database.NewPool(ctx, cfg.DB, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize database", logger.Error(err))
	}
	defer pool.Close()

	courierRepo := postgres.NewCourierRepository(pool)
	deliveryRepo := postgres.NewDeliveryRepository(pool)

	calculatorFactory := factory.NewDeliveryTimeCalculatorFactory()
	txManager := postgres.NewTransactionManager(pool)
	deliveryService := service.NewDeliveryService(deliveryRepo, courierRepo, calculatorFactory, txManager, appLogger)

	orderGateway, err := gateway.NewOrderGateway(cfg.Order.ServiceHost, cfg.Order.MaxRetries, cfg.Order.RetryDelay, prom)
	if err != nil {
		appLogger.Fatal("Failed to initialize order gateway", logger.Error(err))
	}
	defer func() {
		if err := orderGateway.Close(); err != nil {
			appLogger.Error("Failed to close order gateway:", logger.Error(err))
		}
	}()

	handlerFactory := orderChangedService.NewHandlerFactory(deliveryService, courierRepo, deliveryRepo, orderGateway, appLogger)
	orderService := orderChangedService.New(handlerFactory, appLogger)
	handler := orderChangedHandler.NewHandler(orderService, appLogger)

	consumer, err := kafka.NewKafkaConsumer(cfg.Kafka, handler, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize kafka consumer", logger.Error(err))
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			appLogger.Error("Failed to close kafka consumer:", logger.Error(err))
		}
	}()

	consumer.Start(ctx)

	appLogger.Info("Kafka worker started, consuming topic:", logger.Any("topic", cfg.Kafka.Topic))

	<-ctx.Done()

	appLogger.Info("Assign worker stopped gracefully")
}
