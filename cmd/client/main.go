package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/zYoma/yandex_kafka7/internal/application/config"
	"github.com/zYoma/yandex_kafka7/internal/cli"
	"github.com/zYoma/yandex_kafka7/internal/infra/clients/kafka"
)

func main() {
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

	deserializer, err := kafka.GetDeserializer(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating deserializer: %v\n", err)
		os.Exit(1)
	}

	consumer, err := kafka.NewKafkaConsumer(deserializer, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating consumer: %v\n", err)
		os.Exit(1)
	}

	cliClient, err := cli.NewCLI(producer, consumer, cfg.ProductsFile, cfg.RequestsTopic, cfg.RecommendationsTopic)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing CLI: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	err = cliClient.Run(ctx, command, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Kafka CLI Client")
	fmt.Println("\nUsage:")
	fmt.Println("  ./client search <product_name>    - Search for product by name")
	fmt.Println("  ./client recommendations          - Get personal recommendations")
	fmt.Println("\nConfiguration:")
	fmt.Println("  Create .env.client file with:")
	fmt.Println("    BOOTSTRAP_SERVER=localhost:9094,localhost:9095,localhost:9096")
	fmt.Println("    SASL_USERNAME=client")
	fmt.Println("    SASL_PASSWORD=client-secret")
	fmt.Println("    SSL_CA_LOCATION=./kafka-creds/ca.crt")
	fmt.Println("    PRODUCTS_FILE=kafka-connect/output/filtered-products.txt")
}
