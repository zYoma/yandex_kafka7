package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/zYoma/yandex_kafka7/internal/application/config"
	"github.com/zYoma/yandex_kafka7/internal/domain"
	"github.com/zYoma/yandex_kafka7/internal/infra/clients/hdfs"
	kafkacustom "github.com/zYoma/yandex_kafka7/internal/infra/clients/kafka"
	"github.com/zYoma/yandex_kafka7/internal/logger"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		panic(err)
	}

	serializer, err := kafkacustom.GetSerializer(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating serializer: %v\n", err)
		panic(err)
	}

	producer, err := kafkacustom.NewKafkaProducer(serializer, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating producer: %v\n", err)
		panic(err)
	}
	defer producer.Stop()

	hdfsClient, err := hdfs.NewHDFSClient(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating HDFS client: %v\n", err)
		panic(err)
	}
	defer hdfsClient.Close()

	recProcessor := domain.NewRecommendationProcessor(producer, cfg.GetRecommendationsTopic())

	consumer, err := kafkacustom.NewKafkaConsumer(nil, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating consumer: %v\n", err)
		panic(err)
	}
	defer consumer.Consumer.Close()

	consumer.SetHDFSClient(hdfsClient)
	consumer.SetMessageProcessor(recProcessor)

	err = consumer.Consumer.SubscribeTopics([]string{cfg.GetRequestsTopic()}, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to subscribe to topic: %v\n", err)
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Get().Sugar().Infof("Received signal: %v", sig)
		cancel()
	}()

	if err := consumer.StartBatchMessage(ctx); err != nil {
		panic(err)
	}
}
