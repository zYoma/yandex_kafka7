package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/zYoma/yandex_kafka7/internal/application/interfaces"
)

type Price struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type Stock struct {
	Available *int `json:"available"`
	Reserved  *int `json:"reserved"`
}

type Image struct {
	Url *string `json:"url"`
	Alt *string `json:"alt"`
}

type Product struct {
	ProductID      string             `json:"product_id"`
	Name           string             `json:"name"`
	Description    *string            `json:"description"`
	Price          *Price             `json:"price"`
	Category       *string            `json:"category"`
	Brand          *string            `json:"brand"`
	Stock          *Stock             `json:"stock"`
	SKU            *string            `json:"sku"`
	Tags           *[]string          `json:"tags"`
	Images         *[]Image           `json:"images"`
	Specifications *map[string]string `json:"specifications"`
	CreatedAt      *string            `json:"created_at"`
	UpdatedAt      *string            `json:"updated_at"`
	Index          *string            `json:"index"`
	StoreID        *string            `json:"store_id"`
}

type Request struct {
	ProductName string `json:"product_name"`
	Timestamp   string `json:"timestamp"`
}

type Recommendation struct {
	ProductID    string   `json:"product_id"`
	ProductName  string   `json:"product_name"`
	Reason       string   `json:"reason"`
	Confidence   float64  `json:"confidence"`
	Alternatives []string `json:"alternatives"`
	Timestamp    string   `json:"timestamp"`
}

func SendProductsFromFile(ctx context.Context, producer interfaces.Producer, topic string, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var products []Product
	if err := json.Unmarshal(data, &products); err != nil {
		return err
	}

	messages := make([]interface{}, len(products))
	for i := range products {
		messages[i] = &products[i]
	}

	return producer.SendMessages(ctx, messages)
}

type FilteredProduct struct {
	Price     Price
	ProductID string
	Name      string
	Category  string
}

func parsePrice(priceStr string) (Price, error) {
	re := regexp.MustCompile(`amount=([0-9.]+),\s*currency=(\w+)`)
	matches := re.FindStringSubmatch(priceStr)
	if len(matches) != 3 {
		return Price{}, fmt.Errorf("invalid price format")
	}

	var amount float64
	_, err := fmt.Sscanf(matches[1], "%f", &amount)
	if err != nil {
		return Price{}, fmt.Errorf("invalid amount: %w", err)
	}

	return Price{
		Amount:   amount,
		Currency: matches[2],
	}, nil
}

func parseFilteredProductLine(line string) (FilteredProduct, error) {
	product := FilteredProduct{}

	re := regexp.MustCompile(`price=\{([^}]+)\}`)
	priceMatches := re.FindStringSubmatch(line)
	if len(priceMatches) == 2 {
		price, err := parsePrice(priceMatches[1])
		if err == nil {
			product.Price = price
		}
	}

	re = regexp.MustCompile(`product_id=([^,}]+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) == 2 {
		product.ProductID = strings.TrimSpace(matches[1])
	}

	re = regexp.MustCompile(`name=([^,}]+)`)
	matches = re.FindStringSubmatch(line)
	if len(matches) == 2 {
		product.Name = strings.TrimSpace(matches[1])
	}

	re = regexp.MustCompile(`category=([^,}]+)`)
	matches = re.FindStringSubmatch(line)
	if len(matches) == 2 {
		product.Category = strings.TrimSpace(matches[1])
	}

	return product, nil
}

func LoadFilteredProducts(filePath string) ([]FilteredProduct, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var products []FilteredProduct

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		product, err := parseFilteredProductLine(line)
		if err != nil {
			continue
		}

		products = append(products, product)
	}

	return products, nil
}

func SearchProductsByName(products []FilteredProduct, name string) []FilteredProduct {
	var results []FilteredProduct
	nameLower := strings.ToLower(name)

	for _, p := range products {
		if strings.Contains(strings.ToLower(p.Name), nameLower) {
			results = append(results, p)
		}
	}

	return results
}

func SendSearchRequest(ctx context.Context, producer interfaces.Producer, topic string, productName string) error {
	request := Request{
		ProductName: productName,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	value, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	producer.SetTopic(topic)
	return producer.SendMessages(ctx, []interface{}{value})
}

func SendRecommendation(ctx context.Context, producer interfaces.Producer, topic string, recommendation Recommendation) error {
	value, err := json.Marshal(recommendation)
	if err != nil {
		return fmt.Errorf("failed to marshal recommendation: %w", err)
	}

	producer.SetTopic(topic)
	return producer.SendMessages(ctx, []interface{}{value})
}
