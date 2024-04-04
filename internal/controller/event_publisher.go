package controller

import (
	"log"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v2/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
)

type EventPublisher struct {
	publisher *amqp.Publisher
}

var queueSuffix = "ctrls"

func NewEventPublisher(amqpURI string) (*EventPublisher, error) {
	amqpConfig := amqp.NewDurablePubSubConfig(amqpURI, amqp.GenerateQueueNameTopicNameWithSuffix(queueSuffix))
	amqpPublisher, err := amqp.NewPublisher(amqpConfig, watermill.NewStdLogger(true, true))
	if err != nil {
		log.Fatalf("Connection to amqp failed: %v", err)
	}
	return &EventPublisher{amqpPublisher}, nil
}

func (e *EventPublisher) PublishMessage(event []byte, queue string) {
	uuid := watermill.NewUUID()
	msg := message.NewMessage(uuid, event)
	if err := e.publisher.Publish(queue, msg); err != nil {
		log.Printf("Publish failed: %v. Msg uuid: %s", err, uuid)
	}
}
