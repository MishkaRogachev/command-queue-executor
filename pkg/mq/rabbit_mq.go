package mq

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	serverQueue  string
	replyQueue   string
	pending      sync.Map
	cancelFunc   context.CancelFunc
	wg           sync.WaitGroup
	connStr      string
	retryCount   int
	retryBackoff time.Duration
	timeout      time.Duration

	closed bool // new: track if closed
}

func NewRabbitMQ(
	connStr, serverQueue string,
	timeout time.Duration,
	retryCount int,
	retryBackoff time.Duration,
) (*RabbitMQ, error) {
	mq := &RabbitMQ{
		connStr:      connStr,
		serverQueue:  serverQueue,
		timeout:      timeout,
		retryCount:   retryCount,
		retryBackoff: retryBackoff,
	}
	if err := mq.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	if err := mq.setupReplyQueue(); err != nil {
		_ = mq.Close()
		return nil, fmt.Errorf("failed to setup reply queue: %w", err)
	}
	return mq, nil
}

func (mq *RabbitMQ) connect() error {
	conn, err := amqp.Dial(mq.connStr)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return err
	}
	mq.conn = conn
	mq.channel = ch

	_, cancel := context.WithCancel(context.Background())
	mq.cancelFunc = cancel

	log.Println("[RabbitMQ] connected")
	return nil
}

func (mq *RabbitMQ) setupReplyQueue() error {
	name := fmt.Sprintf("rpc.reply.%d", rand.Int63())
	q, err := mq.channel.QueueDeclare(name, false, true, true, false, nil)
	if err != nil {
		return err
	}
	mq.replyQueue = q.Name
	deliveries, err := mq.channel.Consume(q.Name, "", true, true, false, false, nil)
	if err != nil {
		return err
	}
	mq.wg.Add(1)
	go mq.handleReplies(deliveries)
	return nil
}

func (mq *RabbitMQ) handleReplies(deliveries <-chan amqp.Delivery) {
	defer mq.wg.Done()
	for d := range deliveries {
		corrID := d.CorrelationId
		if corrID == "" {
			continue
		}
		val, ok := mq.pending.LoadAndDelete(corrID)
		if !ok {
			continue
		}
		ch := val.(chan Message)
		ch <- Message{Data: string(d.Body)}
		close(ch)
	}
}

func (mq *RabbitMQ) Request(msg Message) (<-chan Message, error) {
	// If we've already closed, fail immediately
	if mq.closed {
		return nil, ErrConnectionClosed
	}

	if mq.channel == nil {
		return nil, errors.New("channel is not open")
	}

	corrID := uuid.NewString()
	responseChan := make(chan Message, 1)
	mq.pending.Store(corrID, responseChan)

	if msg.ReplyTo == "" {
		msg.ReplyTo = mq.replyQueue
	}

	pub := func() error {
		return mq.channel.PublishWithContext(
			context.Background(),
			"",
			mq.serverQueue,
			false,
			false,
			amqp.Publishing{
				ContentType:   "text/plain",
				CorrelationId: corrID,
				ReplyTo:       msg.ReplyTo,
				Body:          []byte(msg.Data),
			},
		)
	}
	var err error
	for i := 0; i < mq.retryCount; i++ {
		if err = pub(); err == nil {
			break
		}
		log.Printf("[RabbitMQ] publish fail (attempt %d): %v", i+1, err)
		time.Sleep(mq.retryBackoff)
	}
	if err != nil {
		mq.pending.Delete(corrID)
		close(responseChan)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), mq.timeout)
	go func(id string) {
		<-ctx.Done()
		if _, loaded := mq.pending.LoadAndDelete(id); loaded {
			close(responseChan)
		}
		cancel()
	}(corrID)

	return responseChan, nil
}

func (mq *RabbitMQ) RegisterHandler(handler func(Message) Message) error {
	if mq.channel == nil {
		return errors.New("channel is not open")
	}
	_, err := mq.channel.QueueDeclare(mq.serverQueue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("queue declare: %w", err)
	}
	deliveries, err := mq.channel.Consume(mq.serverQueue, "", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}
	mq.wg.Add(1)
	go mq.handleRequests(deliveries, handler)
	return nil
}

func (mq *RabbitMQ) handleRequests(deliveries <-chan amqp.Delivery, handler func(Message) Message) {
	defer mq.wg.Done()
	for d := range deliveries {
		if d.ReplyTo == "" || d.CorrelationId == "" {
			continue
		}
		req := Message{Data: string(d.Body), ReplyTo: d.ReplyTo}
		resp := handler(req)
		mq.publishResponse(resp, d.CorrelationId, d.ReplyTo)
	}
}

func (mq *RabbitMQ) publishResponse(resp Message, corrID, queue string) {
	for i := 0; i < mq.retryCount; i++ {
		err := mq.channel.PublishWithContext(
			context.Background(),
			"",
			queue,
			false,
			false,
			amqp.Publishing{
				ContentType:   "text/plain",
				CorrelationId: corrID,
				Body:          []byte(resp.Data),
			},
		)
		if err == nil {
			return
		}
		log.Printf("[RabbitMQ] response publish fail (attempt %d): %v", i+1, err)
		time.Sleep(mq.retryBackoff)
	}
}

func (mq *RabbitMQ) Close() error {
	mq.closed = true
	mq.cancelFunc()

	var errs []error

	if mq.channel != nil {
		if err := mq.channel.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	mq.wg.Wait()

	if mq.conn != nil {
		if err := mq.conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}
	return nil
}
