package producer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// Header represents a Kafka record header.
type Header struct {
	Key   string
	Value []byte
}

// Message encapsulates the data required to publish a Kafka record.
type Message struct {
	Topic   string
	Key     []byte
	Value   []byte
	Headers []Header
}

// Writer defines the subset of kafka.Writer used by the service.
type Writer interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

// WriterFactory creates a kafka writer for a specific topic.
type WriterFactory func(topic string) (Writer, error)

// Service coordinates writing Kafka messages.
type Service struct {
	mu           sync.Mutex
	writers      map[string]Writer
	newWriter    WriterFactory
	defaultTopic string
	now          func() time.Time
}

// ErrNoTopic is returned when neither the payload nor configuration provide a topic.
var ErrNoTopic = errors.New("topic is required")

// NewService creates a service using the provided writer factory.
func NewService(defaultTopic string, factory WriterFactory) *Service {
	if factory == nil {
		factory = func(string) (Writer, error) { return nil, errors.New("writer factory not configured") }
	}
	return &Service{
		writers:      make(map[string]Writer),
		newWriter:    factory,
		defaultTopic: defaultTopic,
		now:          time.Now,
	}
}

// WithNow overrides the time provider for testing purposes.
func (s *Service) WithNow(now func() time.Time) *Service {
	if now != nil {
		s.now = now
	}
	return s
}

// Produce publishes the provided message to Kafka.
func (s *Service) Produce(ctx context.Context, msg Message) error {
	topic := msg.Topic
	if topic == "" {
		topic = s.defaultTopic
	}
	if topic == "" {
		return ErrNoTopic
	}

	writer, err := s.writerForTopic(topic)
	if err != nil {
		return err
	}

	headers := make([]kafka.Header, 0, len(msg.Headers))
	for _, h := range msg.Headers {
		headers = append(headers, kafka.Header{Key: h.Key, Value: h.Value})
	}

	message := kafka.Message{
		Topic:   topic,
		Key:     msg.Key,
		Value:   msg.Value,
		Headers: headers,
		Time:    s.now(),
	}

	if err := writer.WriteMessages(ctx, message); err != nil {
		return fmt.Errorf("write kafka message: %w", err)
	}

	return nil
}

// Close releases all writer resources.
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var firstErr error
	for topic, writer := range s.writers {
		if err := writer.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("close writer for topic %s: %w", topic, err)
		}
		delete(s.writers, topic)
	}
	return firstErr
}

func (s *Service) writerForTopic(topic string) (Writer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if writer, ok := s.writers[topic]; ok {
		return writer, nil
	}
	writer, err := s.newWriter(topic)
	if err != nil {
		return nil, fmt.Errorf("create writer for topic %s: %w", topic, err)
	}
	s.writers[topic] = writer
	return writer, nil
}

// NewKafkaWriterFactory builds a writer factory using the provided broker list.
func NewKafkaWriterFactory(brokers []string) WriterFactory {
	b := append([]string(nil), brokers...)
	return func(topic string) (Writer, error) {
		if topic == "" {
			return nil, ErrNoTopic
		}
		w := &kafka.Writer{
			Addr:                   kafka.TCP(b...),
			Topic:                  topic,
			AllowAutoTopicCreation: true,
			RequiredAcks:           kafka.RequireAll,
			Balancer:               &kafka.Hash{},
		}
		return w, nil
	}
}
