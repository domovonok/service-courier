package changed

import (
	"context"
	"encoding/json"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/dto/queues/order/changed"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/logger"
	"github.com/IBM/sarama"
)

type Service interface {
	ProcessMessage(ctx context.Context, message *changed.Message) error
}

type Handler struct {
	service Service
	log     logger.Logger
}

func NewHandler(service Service, log logger.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

func (h *Handler) Setup(sarama.ConsumerGroupSession) error {
	h.log.Info("Kafka consumer group session started")
	return nil
}

func (h *Handler) Cleanup(sarama.ConsumerGroupSession) error {
	h.log.Info("Kafka consumer group session ended")
	return nil
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case kafkaMessage := <-claim.Messages():
			if kafkaMessage == nil {
				return nil
			}

			h.log.Info(
				"Received message from Kafka",
				logger.Any("topic", kafkaMessage.Topic),
				logger.Any("partition", kafkaMessage.Partition),
				logger.Any("offset", kafkaMessage.Offset),
			)

			var message changed.Message
			if err := json.Unmarshal(kafkaMessage.Value, &message); err != nil {
				h.log.Error(
					"Failed to unmarshal message",
					logger.Error(err),
					logger.Any("value", string(kafkaMessage.Value)),
				)
				session.MarkMessage(kafkaMessage, "")
				continue
			}

			if err := h.service.ProcessMessage(session.Context(), &message); err != nil {
				h.log.Error("Failed to process message", logger.Error(err))
				continue
			}

			session.MarkMessage(kafkaMessage, "")
		case <-session.Context().Done():
			return nil
		}
	}
}
