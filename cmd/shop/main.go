// Package main содержит точку входа для продюсера товаров в Kafka.
package main

import (
	"context"
	"errors"
	"os/signal"
	"syscall"

	"github.com/zYoma/yandex_kafka7/internal/application"
	"github.com/zYoma/yandex_kafka7/internal/application/config"
	"github.com/zYoma/yandex_kafka7/internal/infra/clients/kafka"
	"github.com/zYoma/yandex_kafka7/internal/logger"
)

// main запускает продюсер товаров, отправляет товары из файла в Kafka топик.
func main() {

	// получаем конфигурацию продюсера
	config, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	// слушаем сигналы
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// получаем сериализатор
	serializer, err := kafka.GetSerializer(config)
	if err != nil {
		panic(err)
	}

	// внедряем конкретную реализацию
	producerClient, err := kafka.NewKafkaProducer(serializer, config)
	if err != nil {
		panic(err)
	}
	defer producerClient.Stop()

	// создаем приложение
	producer := application.NewProducer(config, producerClient)

	// запускаем продюсер
	if err := producer.Run(ctx); err != nil {
		if errors.Is(err, application.ErrAppStopped) {
			logger.Get().Info("producer stopped")
			return
		}
		panic(err)
	}

}
