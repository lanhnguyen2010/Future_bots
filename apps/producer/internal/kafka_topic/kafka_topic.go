package kafka_topic

import (
	"fmt"
	"strings"

	"github.com/segmentio/kafka-go"
)

// KafkaTopic represents minimal configuration required to ensure a topic exists.
type KafkaTopic struct {
	Broker            string
	Topic             string
	Partitions        int
	ReplicationFactor int
}

// Create attempts to create the topic on the provided broker controller.
// If the topic already exists the function returns the corresponding error so
// callers can decide whether to ignore it.
func (kt *KafkaTopic) Create() error {
	if kt.Topic == "" {
		return fmt.Errorf("topic name is required")
	}
	if kt.Broker == "" {
		return fmt.Errorf("broker address is required")
	}

	partitions := kt.Partitions
	if partitions <= 0 {
		partitions = 1
	}
	replication := kt.ReplicationFactor
	if replication <= 0 {
		replication = 1
	}

	conn, err := kafka.Dial("tcp", kt.Broker)
	if err != nil {
		return fmt.Errorf("connect broker: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("controller lookup: %w", err)
	}

	controllerAddr := fmt.Sprintf("%s:%d", controller.Host, controller.Port)
	ctrlConn, err := kafka.Dial("tcp", controllerAddr)
	if err != nil {
		return fmt.Errorf("connect controller: %w", err)
	}
	defer ctrlConn.Close()

	topicConfig := kafka.TopicConfig{
		Topic:             kt.Topic,
		NumPartitions:     partitions,
		ReplicationFactor: replication,
	}

	if err := ctrlConn.CreateTopics(topicConfig); err != nil {
		return fmt.Errorf("create topic: %w", err)
	}
	return nil
}

// IsAlreadyExists reports whether the returned error indicates that the topic
// already exists.
func IsAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(strings.ToLower(msg), "already exists")
}
