package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/config"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/database"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/factory"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/gateway"
	orderChangedHandler "github.com/Avito-courses/course-go-avito-domovonok/internal/handler/queues/order/changed"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/repository/postgres"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/service"
	orderChangedService "github.com/Avito-courses/course-go-avito-domovonok/internal/service/order/changed"
	"github.com/IBM/sarama"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, cfg.DB)
	if err != nil {
		log.Fatalln("Failed to initialize database:", err)
	}
	defer pool.Close()

	courierRepo := postgres.NewCourierRepository(pool)
	deliveryRepo := postgres.NewDeliveryRepository(pool)

	calculatorFactory := factory.NewDeliveryTimeCalculatorFactory()
	txManager := postgres.NewTransactionManager(pool)
	deliveryService := service.NewDeliveryService(deliveryRepo, courierRepo, calculatorFactory, txManager)

	orderGateway, err := gateway.NewOrderGateway(cfg.Order.ServiceHost)
	if err != nil {
		log.Fatalln("Failed to initialize order gateway:", err)
	}
	defer orderGateway.Close()

	orderService := orderChangedService.New(deliveryService, courierRepo, deliveryRepo, orderGateway)
	handler := orderChangedHandler.NewHandler(orderService)

	kafkaVersion, err := sarama.ParseKafkaVersion(cfg.Kafka.Version)
	if err != nil {
		log.Fatalln("Failed to parse Kafka version:", err)
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = kafkaVersion
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = true
	saramaConfig.Consumer.Offsets.AutoCommit.Interval = cfg.Kafka.AutoCommitInterval

	log.Printf("Kafka brokers: %v", cfg.Kafka.Brokers)

	kafkaClient, err := sarama.NewConsumerGroup(cfg.Kafka.Brokers, cfg.Kafka.ConsumerGroup, saramaConfig)
	if err != nil {
		log.Fatalln("unable to create kafka consumer group:", err)
	}
	defer kafkaClient.Close()

	go func() {
		for {
			err := kafkaClient.Consume(ctx, []string{cfg.Kafka.Topic}, handler)
			if err != nil {
				log.Printf("consume error: %v", err)
			}

			select {
			case <-ctx.Done():
				return
			}
		}
	}()

	log.Printf("Kafka worker started, consuming topic: %s", cfg.Kafka.Topic)

	waitGracefulShutdown(cancel)

	log.Println("Assign worker stopped gracefully.")
}

func waitGracefulShutdown(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var sig string
	select {
	case s := <-sigChan:
		sig = s.String()
	}

	cancel()

	log.Printf("Shutdown signal (%s)", sig)
}
