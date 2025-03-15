package tweet

import (
	"context"
	"fmt"
	"log"

	rabbit "github.com/devpablocristo/monorepo/pkg/brokers/rabbitmq/amqp091/producer"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/tweet/usecases/domain"
)

type broker struct {
	producer   rabbit.Producer
	routingKey string // Se utiliza como clave de enrutamiento en el exchange.
}

// NewRabbitMQTweetPublisher crea una nueva instancia del adapter para publicar eventos de tweet.
func NewBroker(producer rabbit.Producer, routingKey string) Broker {
	return &broker{
		producer:   producer,
		routingKey: routingKey,
	}
}

// PublishTweetCreated publica un evento de tweet creado.
// Se serializa el tweet (utilizando, por ejemplo, JSON) y se envía al exchange configurado.
func (b *broker) PublishTweetCreated(ctx context.Context, tweet *domain.Tweet) error {
	// El método Produce del producer se encarga de serializar y enviar el mensaje.
	// Se utiliza la routingKey definida para direccionar el mensaje.
	corrID, err := b.producer.Produce(ctx, b.routingKey, "", "", tweet)
	if err != nil {
		return fmt.Errorf("failed to publish tweet event: %w", err)
	}
	log.Printf("Tweet event published with correlation id: %s", corrID)
	return nil
}
