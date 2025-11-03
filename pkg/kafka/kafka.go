package kafka

import (
	"context"
	"errors"
	"strings"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type SmsKafkaMessage struct {
	SmsId     uint   `json:"sms_id"`
	To        string `json:"to"`
	Content   string `json:"content"`
	Provider  string `json:"provider"`
	UserId    uint   `json:"user_id"`
	ServiceId uint   `json:"service_id"`
}

type KafkaInterface interface {
	Publish(ctx context.Context, key string, value []byte) error
	ReadMessage(ctx context.Context) (*kafka.Message, error)
	UseReader(groupID string) error
	Close() error
}

type Client struct {
	brokers []string
	topic   string
	writer  *kafka.Writer
	reader  *kafka.Reader
}

func Init(brokers string, topic string) (KafkaInterface, error) {
	bs := splitBrokers(brokers)

	c := &Client{
		brokers: bs,
		topic:   topic,
		writer: &kafka.Writer{
			Addr:         kafka.TCP(bs...),
			Topic:        topic,
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireOne,
			Async:        false,
		},
	}
	return c, nil
}

func (c *Client) UseReader(groupID string) error {
	if groupID == "" {
		return errors.New("kafkax: groupID required")
	}
	c.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.brokers,
		GroupID:  groupID,
		Topic:    c.topic,
		MinBytes: 1,
		MaxBytes: 1 << 20,
		MaxWait:  time.Second,
	})
	return nil
}

func (c *Client) Close() error {
	var err error
	if c.reader != nil {
		if e := c.reader.Close(); e != nil && err == nil {
			err = e
		}
	}
	if c.writer != nil {
		if e := c.writer.Close(); e != nil && err == nil {
			err = e
		}
	}
	return err
}

func (c *Client) Publish(ctx context.Context, key string, value []byte) error {
	return c.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: value,
		Time:  time.Now(),
	})
}

func (c *Client) ReadMessage(ctx context.Context) (*kafka.Message, error) {
	if c.reader == nil {
		return nil, errors.New("kafka: reader not initialized; call UseReader(groupID)")
	}
	m, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func splitBrokers(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
