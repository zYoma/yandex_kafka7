package main

import (
	"context"
	"errors"
	"os/signal"
	"syscall"

	"github.com/zYoma/yandex_kafka7/internal/application"
	"github.com/zYoma/yandex_kafka7/internal/application/config"
	"github.com/zYoma/yandex_kafka7/internal/infra/clients/hdfs"
	"github.com/zYoma/yandex_kafka7/internal/infra/clients/kafka"
	"github.com/zYoma/yandex_kafka7/internal/logger"
)

func main() {
	config, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	deserializer, err := kafka.GetDeserializer(config)
	if err != nil {
		panic(err)
	}

	consumerClient, err := kafka.NewKafkaConsumer(deserializer, config)
	if err != nil {
		panic(err)
	}

	hdfsClient, err := hdfs.NewHDFSClient(config)
	if err != nil {
		panic(err)
	}

	consumerClient.SetHDFSClient(hdfsClient)

	consumer := application.NewConsumerWithHDFS(config, consumerClient, hdfsClient)

	if err := consumer.Run(ctx); err != nil {
		if errors.Is(err, application.ErrAppStopped) {
			logger.Get().Info("consumer stopped")
			return
		}
		panic(err)
	}
}
