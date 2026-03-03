// Package domain содержит бизнес-логику и сущности: Product, Recommendation, Request.
package domain

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/zYoma/yandex_kafka7/internal/application/interfaces"
	"github.com/zYoma/yandex_kafka7/internal/logger"
)

type Price struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// Stock представляет информацию о наличии товара на складе.
type Stock struct {
	Available *int `json:"available"`
	Reserved  *int `json:"reserved"`
}

// Image представляет информацию об изображении товара.
type Image struct {
	Url *string `json:"url"`
	Alt *string `json:"alt"`
}

// RecommendationProcessor обрабатывает запросы и генерирует рекомендации.
type RecommendationProcessor struct {
	producer interfaces.Producer
	recTopic string
}

func NewRecommendationProcessor(producer interfaces.Producer, recommendationsTopic string) *RecommendationProcessor {
	return &RecommendationProcessor{
		producer: producer,
		recTopic: recommendationsTopic,
	}
}

func (p *RecommendationProcessor) ProcessMessage(ctx context.Context, data []byte) error {
	var strData string
	decodedData := data

	if err := json.Unmarshal(data, &strData); err == nil {
		logger.Get().Sugar().Infof("Request is string-encoded: %s", strData)
		decodedBytes, err := base64.StdEncoding.DecodeString(strings.Trim(strData, "\""))
		if err == nil {
			decodedData = decodedBytes
			logger.Get().Sugar().Infof("Decoded base64 data: %s", string(decodedData))
		}
	}

	var request Request
	if err := json.Unmarshal(decodedData, &request); err != nil {
		logger.Get().Sugar().Errorf("Failed to unmarshal request: %v, data: %s", err, string(decodedData))
		return err
	}

	recommendation := Recommendation{
		ProductID:    "rec_" + fmt.Sprintf("%d", time.Now().UnixNano()),
		ProductName:  request.ProductName + " - Recommended",
		Reason:       "Based on your search history",
		Confidence:   0.85,
		Alternatives: []string{"alt_1", "alt_2", "alt_3"},
		Timestamp:    time.Now().Format(time.RFC3339),
	}

	if err := SendRecommendation(ctx, p.producer, p.recTopic, recommendation); err != nil {
		logger.Get().Sugar().Errorf("Failed to send recommendation: %v", err)
		return err
	} else {
		logger.Get().Sugar().Infof("Successfully sent recommendation for product: %s", request.ProductName)
	}

	return nil
}

// Product представляет товар в системе электронной коммерции.
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

// Request представляет поисковый запрос от пользователя.
type Request struct {
	ProductName string `json:"product_name"`
	Timestamp   string `json:"timestamp"`
}

// Recommendation представляет рекомендованную продукцию с указанием причины.
type Recommendation struct {
	ProductID    string   `json:"product_id"`
	ProductName  string   `json:"product_name"`
	Reason       string   `json:"reason"`
	Confidence   float64  `json:"confidence"`
	Alternatives []string `json:"alternatives"`
	Timestamp    string   `json:"timestamp"`
}

// ProcessProducts обрабатывает список продуктов и сохраняет их в HDFS.
func ProcessProducts(ctx context.Context, products []Product, hdfsClient interfaces.HDFSClient) bool {
	for _, product := range products {
		data, err := json.Marshal(product)
		if err != nil {
			logger.Get().Sugar().Errorf("Failed to marshal product: %v", err)
			return false
		}

		msgID := product.ProductID
		if err := hdfsClient.WriteRequestData(ctx, data, msgID); err != nil {
			logger.Get().Sugar().Errorf("Failed to write to HDFS: %v", err)
			return false
		}
	}

	return true
}

// SendProductsFromFile читает товары из JSON файла и отправляет их в Kafka.
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

// FilteredProduct представляет отфильтрованный продукт из файла вывода Kafka Connect.
type FilteredProduct struct {
	Price     Price
	ProductID string
	Name      string
	Category  string
}

// parsePrice парсит строку цены из формата "amount=X, currency=Y".
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

// parseFilteredProductLine парсит строку из файла filtered-products.txt в структуру FilteredProduct.
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

// LoadFilteredProducts загружает отфильтрованные продукты из файла.
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

// SearchProductsByName выполняет поиск продуктов по имени без учета регистра.
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

// SendSearchRequest отправляет поисковый запрос в Kafka топик для аналитики.
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

// SendRecommendation отправляет рекомендацию в Kafka топик.
func SendRecommendation(ctx context.Context, producer interfaces.Producer, topic string, recommendation Recommendation) error {
	value, err := json.Marshal(recommendation)
	if err != nil {
		return fmt.Errorf("failed to marshal recommendation: %w", err)
	}

	producer.SetTopic(topic)
	return producer.SendMessages(ctx, []interface{}{value})
}
