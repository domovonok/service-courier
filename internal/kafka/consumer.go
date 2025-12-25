package kafka

import (
	"context"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/config"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/logger"
	"github.com/IBM/sarama"
)

type Consumer struct {
	log     logger.Logger
	client  sarama.ConsumerGroup
	topics  []string
	handler sarama.ConsumerGroupHandler
}

func NewKafkaConsumer(cfg config.KafkaConfig, handler sarama.ConsumerGroupHandler, log logger.Logger) (*Consumer, error) {
	saramaCfg, err := newSaramaConfig(cfg.Sarama)
	if err != nil {
		return nil, err
	}

	client, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.ConsumerGroup, saramaCfg)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		log:     log,
		client:  client,
		topics:  []string{cfg.Topic},
		handler: handler,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for {
			if err := c.client.Consume(ctx, c.topics, c.handler); err != nil {
				c.log.Error("Consume error", logger.Error(err))
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()
}

func (c *Consumer) Close() error {
	return c.client.Close()
}

func newSaramaConfig(cfg config.SaramaConfig) (*sarama.Config, error) {
	kafkaVersion, err := sarama.ParseKafkaVersion(cfg.Version)
	if err != nil {
		return nil, err
	}

	saramaCfg := sarama.NewConfig()
	saramaCfg.Version = kafkaVersion
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaCfg.Consumer.Offsets.AutoCommit.Enable = true
	saramaCfg.Consumer.Offsets.AutoCommit.Interval = cfg.AutoCommitInterval

	return saramaCfg, nil
}
