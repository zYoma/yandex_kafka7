package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/zYoma/yandex_kafka7/internal/application/config"
	"github.com/zYoma/yandex_kafka7/internal/domain"
	"github.com/zYoma/yandex_kafka7/internal/infra/clients/kafka"
)

func main() {
	if err := godotenv.Load(".env.analytic_client"); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: .env.analytic_client file not found: %v\n", err)
	}

	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	serializer, err := kafka.GetSerializer(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating serializer: %v\n", err)
		os.Exit(1)
	}

	producer, err := kafka.NewKafkaProducer(serializer, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating producer: %v\n", err)
		os.Exit(1)
	}
	defer producer.Stop()

	fmt.Println("Sending recommendations to Kafka topic 'recommendations'...")

	recommendations := []domain.Recommendation{
		{
			ProductID:    "12345",
			ProductName:  "Тестовый товар 1",
			Reason:       "Популярный товар в категории",
			Confidence:   0.95,
			Alternatives: []string{"54321", "98765"},
			Timestamp:    time.Now().Format(time.RFC3339),
		},
		{
			ProductID:    "54321",
			ProductName:  "Тестовый товар 2",
			Reason:       "Рекомендован на основе истории просмотров",
			Confidence:   0.88,
			Alternatives: []string{"12345", "98765"},
			Timestamp:    time.Now().Format(time.RFC3339),
		},
		{
			ProductID:    "98765",
			ProductName:  "Новый разрешенный товар",
			Reason:       "Новинка недели",
			Confidence:   0.92,
			Alternatives: []string{"12345", "54321"},
			Timestamp:    time.Now().Format(time.RFC3339),
		},
	}

	for i, rec := range recommendations {
		err := domain.SendRecommendation(ctx, producer, "recommendations", rec)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to send recommendation %d: %v\n", i+1, err)
			continue
		}

		fmt.Printf("Recommendation %d sent successfully: %s (%.2f confidence)\n",
			i+1, rec.ProductName, rec.Confidence)
	}

	fmt.Println("\nAll recommendations sent successfully!")
}
