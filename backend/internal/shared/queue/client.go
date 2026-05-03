package queue

import (
	"context"
	"github.com/rs/zerolog/log"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

func NewClient(ctx context.Context, url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{
		conn: conn,
		ch:   ch,
	}, nil
}

type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func (c *Client) Close() error {
	if err := c.ch.Close(); err != nil {
		return err
	}
	return c.conn.Close()
}

func (c *Client) DeclareQueue(queueName string) error {
	_, err := c.ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,	   // arguments
	)
	return err
}

func (c *Client) Publish(ctx context.Context, queueName string, job proto.Message) error {
	body, err := proto.Marshal(job)
	if err != nil {
		return err
	}

	c.ch.PublishWithContext(ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/protobuf",
			Body: body,
		},
	)
	return nil
}

type ScanJobDelivery struct {
	Job *ScanJob
	Msg amqp.Delivery
}

type ContentDelivery struct {
	Job *Content
	Msg amqp.Delivery
}

func (c *Client) ConsumeScanJobs(ctx context.Context, queueName string) (<-chan ScanJobDelivery, error) {
	msgs, err := c.ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	ch := make(chan ScanJobDelivery)
	go func() {
		defer close(ch)
		for msg := range msgs {
			var job ScanJob
			if err := proto.Unmarshal(msg.Body, &job); err != nil {
				log.Error().Err(err).Msg("failed to unmarshal ScanJob")
				msg.Nack(false, false)
				continue
			}
			ch <- ScanJobDelivery{Job: &job, Msg: msg}
		}
	}()

	return ch, nil
}

func (c *Client) ConsumeContent(ctx context.Context, queueName string) (<-chan ContentDelivery, error) {
	msgs, err := c.ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	ch := make(chan ContentDelivery)
	go func() {
		defer close(ch)
		for msg := range msgs {
			var job Content
			if err := proto.Unmarshal(msg.Body, &job); err != nil {
				log.Error().Err(err).Msg("failed to unmarshal Content")
				msg.Nack(false, false)
				continue
			}
			ch <- ContentDelivery{Job: &job, Msg: msg}
		}
	}()

	return ch, nil
}
