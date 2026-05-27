package queue

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

func NewClient(ctx context.Context, url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn}, nil
}

type Client struct {
	conn *amqp.Connection
}

func (c *Client) Close() error {
	return c.conn.Close()
}

const deliveryLimit = 3

func (c *Client) DeclareQueueWithDLQ(queueName string) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	dlxName := queueName + ".dlx"
	deadName := queueName + ".dead"

	if err := ch.ExchangeDeclare(dlxName, "direct", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(deadName, true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(deadName, queueName, dlxName, false, nil); err != nil {
		return err
	}
	_, err = ch.QueueDeclare(queueName, true, false, false, false, amqp.Table{
		"x-queue-type":           "quorum",
		"x-dead-letter-exchange": dlxName,
		"x-delivery-limit":       int64(deliveryLimit),
	})
	return err
}

func (c *Client) Publish(ctx context.Context, queueName string, job proto.Message) error {
	body, err := proto.Marshal(job)
	if err != nil {
		return err
	}

	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	return ch.PublishWithContext(ctx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/protobuf",
			Body:        body,
		},
	)
}

type SpotifyGatewayDelivery struct {
	Job *ContentFetchJob
	msg amqp.Delivery
}

func (d SpotifyGatewayDelivery) Ack(multiple bool) error {
	return d.msg.Ack(multiple)
}

func (d SpotifyGatewayDelivery) Nack(multiple, requeue bool) error {
	return d.msg.Nack(multiple, requeue)
}

type ScanWorkerDelivery struct {
	Job *ContentScanJob
	msg amqp.Delivery
}

func (d ScanWorkerDelivery) Ack(multiple bool) error {
	return d.msg.Ack(multiple)
}

func (d ScanWorkerDelivery) Nack(multiple, requeue bool) error {
	return d.msg.Nack(multiple, requeue)
}

func (c *Client) ConsumeContentFetchJobs(ctx context.Context, queueName string) (<-chan SpotifyGatewayDelivery, error) {
	ch, err := c.conn.Channel()
	if err != nil {
		return nil, err
	}

	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		ch.Close()
		return nil, err
	}

	out := make(chan SpotifyGatewayDelivery)
	go func() {
		defer close(out)
		defer ch.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				var job ContentFetchJob
				if err := proto.Unmarshal(msg.Body, &job); err != nil {
					log.Error().Err(err).Msg("failed to unmarshal ScanJob")
					msg.Nack(false, false)
					continue
				}
				select {
				case <-ctx.Done():
					return
				case out <- SpotifyGatewayDelivery{Job: &job, msg: msg}:
				}
			}
		}
	}()

	return out, nil
}

func (c *Client) ConsumeContent(ctx context.Context, queueName string) (<-chan ScanWorkerDelivery, error) {
	ch, err := c.conn.Channel()
	if err != nil {
		return nil, err
	}

	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		ch.Close()
		return nil, err
	}

	out := make(chan ScanWorkerDelivery)
	go func() {
		defer close(out)
		defer ch.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				var job ContentScanJob
				if err := proto.Unmarshal(msg.Body, &job); err != nil {
					log.Error().Err(err).Msg("failed to unmarshal Content")
					msg.Nack(false, false)
					continue
				}
				select {
				case <-ctx.Done():
					return
				case out <- ScanWorkerDelivery{Job: &job, msg: msg}:
				}
			}
		}
	}()

	return out, nil
}
