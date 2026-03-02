package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/zYoma/yandex_kafka7/internal/application/interfaces"
	"github.com/zYoma/yandex_kafka7/internal/domain"
)

type CLI struct {
	producer             interfaces.Producer
	consumer             interfaces.Consumer
	products             []domain.FilteredProduct
	requestsTopic        string
	recommendationsTopic string
}

func NewCLI(producer interfaces.Producer, consumer interfaces.Consumer, productsFile string, requestsTopic string, recommendationsTopic string) (*CLI, error) {
	products, err := domain.LoadFilteredProducts(productsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load products: %w", err)
	}

	return &CLI{
		producer:             producer,
		consumer:             consumer,
		products:             products,
		requestsTopic:        requestsTopic,
		recommendationsTopic: recommendationsTopic,
	}, nil
}

func (c *CLI) Search(ctx context.Context, productName string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	fmt.Printf("Searching for product: %s\n", productName)

	results := domain.SearchProductsByName(c.products, productName)

	if len(results) == 0 {
		fmt.Printf("No products found matching '%s'\n", productName)
	} else {
		fmt.Printf("Found %d product(s):\n\n", len(results))
		for _, p := range results {
			fmt.Printf("ID: %s\n", p.ProductID)
			fmt.Printf("Name: %s\n", p.Name)
			fmt.Printf("Category: %s\n", p.Category)
			fmt.Printf("Price: %.2f %s\n", p.Price.Amount, p.Price.Currency)
			fmt.Println("---")
		}
	}

	err := domain.SendSearchRequest(ctx, c.producer, c.requestsTopic, productName)
	if err != nil {
		return fmt.Errorf("failed to send search request: %w", err)
	}

	fmt.Println("Search request sent to analytics system")
	return nil
}

func (c *CLI) GetRecommendations(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	fmt.Println("Reading recommendation from Kafka...")

	msg, err := c.consumer.ReadOneMessage(ctx, c.recommendationsTopic)
	if err != nil {
		return fmt.Errorf("failed to read recommendation: %w", err)
	}

	fmt.Printf("\nRecommendation:\n%s\n", string(msg))

	return nil
}

func (c *CLI) Run(ctx context.Context, command string, args []string) error {
	switch command {
	case "search":
		if len(args) < 1 {
			return fmt.Errorf("search command requires a product name argument")
		}
		productName := args[0]
		return c.Search(ctx, productName)
	case "recommendations":
		return c.GetRecommendations(ctx)
	default:
		return fmt.Errorf("unknown command: %s\nAvailable commands: search, recommendations", command)
	}
}
