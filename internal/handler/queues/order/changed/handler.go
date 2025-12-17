package changed

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/dto/queues/order/changed"
	"github.com/IBM/sarama"
)

type Service interface {
	ProcessMessage(ctx context.Context, message *changed.Message) error
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Setup(sarama.ConsumerGroupSession) error {
	log.Println("Kafka consumer group session started")
	return nil
}

func (h *Handler) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("Kafka consumer group session ended")
	return nil
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case kafkaMessage := <-claim.Messages():
			if kafkaMessage == nil {
				return nil
			}

			log.Printf("Received message from Kafka: topic=%s, partition=%d, offset=%d",
				kafkaMessage.Topic, kafkaMessage.Partition, kafkaMessage.Offset)

			var message changed.Message
			if err := json.Unmarshal(kafkaMessage.Value, &message); err != nil {
				log.Printf("Failed to unmarshal message: %v, value: %s", err, string(kafkaMessage.Value))
				session.MarkMessage(kafkaMessage, "")
				continue
			}

			if err := h.service.ProcessMessage(session.Context(), &message); err != nil {
				log.Printf("Failed to process message: %v", err)
				continue
			}

			session.MarkMessage(kafkaMessage, "")

		case <-session.Context().Done():
			return nil
		}
	}
}
