package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"eshkere/internal/analytics"
	analyticsch "eshkere/internal/analytics/clickhouse"
	"eshkere/internal/config"

	kafkago "github.com/segmentio/kafka-go"
)

const (
	defaultBatchSize    = 500
	defaultFlushTimeout = 2 * time.Second
)

type eventWriter interface {
	InsertAdEvents(ctx context.Context, events []analytics.AdEvent) error
	Close() error
}

func main() {
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.ReadConfig(*configPath)
	if err != nil {
		log.Fatalf("read config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	writer, err := analyticsch.NewWriter(ctx, analyticsch.Config{
		Addr:     cfg.ClickHouse.Addr,
		Database: cfg.ClickHouse.Database,
		Username: cfg.ClickHouse.Username,
		Password: cfg.ClickHouse.Password,
	})
	if err != nil {
		log.Fatalf("clickhouse writer: %v", err)
	}
	defer writer.Close()

	brokers := cfg.Kafka.BrokerList()
	if err := ensureTopic(ctx, brokers, cfg.Kafka.AdEventsTopic); err != nil {
		log.Fatalf("ensure kafka topic: %v", err)
	}

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:  brokers,
		Topic:    cfg.Kafka.AdEventsTopic,
		GroupID:  consumerGroup(cfg.Kafka.ConsumerGroup),
		MinBytes: 1,
		MaxBytes: 10e6,
		Dialer: &kafkago.Dialer{
			Timeout:   5 * time.Second,
			DualStack: true,
			Resolver:  net.DefaultResolver,
		},
	})
	defer reader.Close()

	log.Printf("analytics consumer started: topic=%s group=%s", cfg.Kafka.AdEventsTopic, consumerGroup(cfg.Kafka.ConsumerGroup))
	if err := run(ctx, reader, writer, defaultBatchSize, defaultFlushTimeout); err != nil && ctx.Err() == nil {
		log.Fatalf("run consumer: %v", err)
	}
}

func ensureTopic(ctx context.Context, brokers []string, topic string) error {
	if len(brokers) == 0 {
		return errors.New("kafka brokers cannot be empty")
	}
	if topic == "" {
		return errors.New("kafka topic cannot be empty")
	}

	dialer := &kafkago.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("dial kafka broker: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("get kafka controller: %w", err)
	}

	controllerAddr := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
	controllerConn, err := dialer.DialContext(ctx, "tcp", controllerAddr)
	if err != nil {
		return fmt.Errorf("dial kafka controller: %w", err)
	}
	defer controllerConn.Close()

	if err = controllerConn.CreateTopics(kafkago.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}); err != nil {
		return fmt.Errorf("create kafka topic %s: %w", topic, err)
	}

	log.Printf("kafka topic is ready: %s", topic)
	return nil
}

func consumerGroup(group string) string {
	if group != "" {
		return group
	}
	return "analytics-consumer"
}

func run(ctx context.Context, reader *kafkago.Reader, writer eventWriter, batchSize int, flushTimeout time.Duration) error {
	var (
		events   []analytics.AdEvent
		messages []kafkago.Message
		timer    = time.NewTimer(flushTimeout)
	)
	defer timer.Stop()

	flush := func() error {
		defer resetTimer(timer, flushTimeout)
		if len(events) == 0 {
			return nil
		}
		flushCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := writer.InsertAdEvents(flushCtx, events); err != nil {
			return err
		}
		if err := reader.CommitMessages(flushCtx, messages...); err != nil {
			return err
		}
		log.Printf("flushed ad events: count=%d", len(events))
		events = events[:0]
		messages = messages[:0]
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return flush()
		case <-timer.C:
			if err := flush(); err != nil {
				return err
			}
		default:
			fetchCtx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
			msg, err := reader.FetchMessage(fetchCtx)
			cancel()
			if err != nil {
				if ctx.Err() != nil {
					return flush()
				}
				if errors.Is(err, context.DeadlineExceeded) {
					continue
				}
				return err
			}

			var event analytics.AdEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("skip invalid ad event: topic=%s partition=%d offset=%d err=%v", msg.Topic, msg.Partition, msg.Offset, err)
				if err := reader.CommitMessages(ctx, msg); err != nil {
					return err
				}
				continue
			}

			events = append(events, event)
			messages = append(messages, msg)
			if len(events) >= batchSize {
				if err := flush(); err != nil {
					return err
				}
			}
		}
	}
}

func resetTimer(timer *time.Timer, timeout time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(timeout)
}
