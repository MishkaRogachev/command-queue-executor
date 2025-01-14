package mq

import (
	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn     *amqp091.Connection
	channel  *amqp091.Channel
	queueMap map[string]<-chan Message
	exchange string
}

func NewRabbitMQ(url, exchange string) (*RabbitMQ, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	if err := ch.ExchangeDeclare(
		exchange,
		"direct",
		true,  // durable
		false, // autoDelete
		false, // internal
		false, // noWait
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &RabbitMQ{
		conn:     conn,
		channel:  ch,
		queueMap: make(map[string]<-chan Message),
		exchange: exchange,
	}, nil
}

func (r *RabbitMQ) Publish(msg Message) error {
	return r.channel.Publish(
		r.exchange,
		msg.Topic,
		false,
		false,
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg.Value),
		},
	)
}

func (r *RabbitMQ) Subscribe(topic string) (<-chan Message, error) {
	queue, err := r.channel.QueueDeclare(
		topic,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,
	)
	if err != nil {
		return nil, err
	}

	if err := r.channel.QueueBind(
		queue.Name,
		topic,
		r.exchange,
		false,
		nil,
	); err != nil {
		return nil, err
	}

	deliveries, err := r.channel.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	msgChan := make(chan Message)
	go func() {
		for d := range deliveries {
			msgChan <- Message{Topic: topic, Value: string(d.Body)}
		}
		close(msgChan)
	}()

	r.queueMap[topic] = msgChan
	return msgChan, nil
}

func (r *RabbitMQ) Unsubscribe(topic string) error {
	if _, exists := r.queueMap[topic]; !exists {
		return ErrNotSubscribedToTopic
	}

	delete(r.queueMap, topic)
	return nil
}

func (r *RabbitMQ) Close() error {
	if err := r.channel.Close(); err != nil {
		return err
	}
	return r.conn.Close()
}
