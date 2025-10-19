package main

import (
	"bufio"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/future-bots/platform/config"
	"github.com/future-bots/producer/internal/kafka_topic"
	"github.com/future-bots/producer/internal/producer"
	stocksparser "github.com/future-bots/producer/internal/stock_parser"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	dataFile := config.EnvOrDefault("PRODUCER_DATA_FILE", filepath.Join("apps", "producer", "internal", "data", "stocks-ssi-ps_s_20251013.txt"))
	topic := config.EnvOrDefault("PRODUCER_TOPIC", "ssi_ps")
	brokersEnv := config.EnvOrDefault("PRODUCER_KAFKA_BROKERS", "localhost:9092")
	brokers := splitAndClean(brokersEnv)
	if len(brokers) == 0 {
		logger.Error("no kafka brokers configured", "env", brokersEnv)
		os.Exit(1)
	}

	topicCfg := kafka_topic.KafkaTopic{
		Broker:            brokers[0],
		Topic:             topic,
		Partitions:        config.IntFromEnv("PRODUCER_TOPIC_PARTITIONS", 1),
		ReplicationFactor: config.IntFromEnv("PRODUCER_TOPIC_REPLICATION", 1),
	}

	if err := topicCfg.Create(); err != nil {
		if kafka_topic.IsAlreadyExists(err) {
			logger.Info("kafka topic already exists", "topic", topic)
		} else {
			logger.Error("failed to ensure topic", "topic", topic, "error", err)
			os.Exit(1)
		}
	} else {
		logger.Info("kafka topic created", "topic", topic)
	}

	file, err := os.Open(dataFile)
	if err != nil {
		logger.Error("failed to open data file", "file", dataFile, "error", err)
		os.Exit(1)
	}
	defer file.Close()

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      brokers,
		Topic:        topic,
		RequiredAcks: int(kafka.RequireAll),
		Balancer:     &kafka.Hash{},
		BatchTimeout: 200 * time.Millisecond,
	})
	defer writer.Close()

	parser := stocksparser.HoseStockParser{BaseParser: stocksparser.BaseParser{Columns: stocksparser.HoseColumns}}
	scanner := bufio.NewScanner(file)
	produced := 0

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			logger.Info("ingest interrupted", "produced", produced)
			return
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parsed, err := parser.Parse(line)
		if err != nil || parsed == nil {
			logger.Warn("failed to parse line", "error", err)
			continue
		}

		snapshot, err := producer.MapToSsiPsSnapshot(parsed)
		if err != nil {
			logger.Warn("failed to map snapshot", "error", err)
			continue
		}

		payload, err := proto.Marshal(snapshot)
		if err != nil {
			logger.Warn("failed to marshal snapshot", "code", snapshot.GetCode(), "error", err)
			continue
		}

		msg := kafka.Message{
			Key:   []byte(snapshot.GetCode()),
			Value: payload,
		}
		if ts := snapshot.GetTimestamp(); ts != nil {
			msg.Time = ts.AsTime()
		}

		if err := writer.WriteMessages(ctx, msg); err != nil {
			logger.Error("failed to write message", "code", snapshot.GetCode(), "error", err)
			os.Exit(1)
		}
		produced++
	}

	if err := scanner.Err(); err != nil {
		logger.Error("scanner error", "error", err)
		os.Exit(1)
	}

	logger.Info("ingest completed", "produced", produced, "file", dataFile, "topic", topic)
}

func splitAndClean(csv string) []string {
	parts := strings.Split(csv, ",")
	cleaned := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			cleaned = append(cleaned, v)
		}
	}
	return cleaned
}
