package main

import (
	"context"
	"fmt"
	"os"

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

	ctx := context.Background()

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

	analyticsService := domain.NewAnalyticsService(producer, hdfsClient, cfg)

	if err := analyticsService.Run(ctx); err != nil {
		if _, ok := err.(domain.ErrAppStopped); ok {
			logger.Get().Info("analytic consumer stopped")
			return
		}
		panic(err)
	}
}
