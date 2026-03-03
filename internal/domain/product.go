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

	kafkago "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/zYoma/yandex_kafka7/internal/application/interfaces"
	"github.com/zYoma/yandex_kafka7/internal/logger"
)

type ErrAppStopped struct{}

// Error возвращает строковое представление ошибки остановки приложения.
func (e ErrAppStopped) Error() string {
	return "application stopped"
}

// ConsumerConfig определяет интерфейс для получения конфигурации консьюмера.
type ConsumerConfig interface {
	GetConsumerConfig() *kafkago.ConfigMap
	GetProducerConfig() *kafkago.ConfigMap
	GetBootstrapServers() string
	GetRequestsTopic() string
	GetRecommendationsTopic() string
}

// AnalyticsService обрабатывает поисковые запросы, сохраняет их в HDFS и отправляет рекомендации.
type AnalyticsService struct {
	producer interfaces.Producer
	hdfs     interfaces.HDFSClient
	cfg      ConsumerConfig
}

// NewAnalyticsService создает новый сервис аналитики с указанными зависимостями.
func NewAnalyticsService(producer interfaces.Producer, hdfs interfaces.HDFSClient, cfg ConsumerConfig) *AnalyticsService {
	return &AnalyticsService{
		producer: producer,
		hdfs:     hdfs,
		cfg:      cfg,
	}
}

// Run запускает сервис аналитики для обработки поисковых запросов из Kafka.
func (s *AnalyticsService) Run(ctx context.Context) error {
	cfgMap := s.cfg.GetConsumerConfig()
	kafkaConsumer, err := kafkago.NewConsumer(cfgMap)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}
	defer kafkaConsumer.Close()

	err = kafkaConsumer.SubscribeTopics([]string{s.cfg.GetRequestsTopic()}, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	logger.Get().Sugar().Infof("Analytics consumer started, listening to requests topic...")

	for {
		select {
		case <-ctx.Done():
			return ErrAppStopped{}
		default:
			msg, err := kafkaConsumer.ReadMessage(100)
			if err != nil {
				if kafkaErr, ok := err.(kafkago.Error); ok && kafkaErr.Code() == kafkago.ErrTimedOut {
					continue
				}
				logger.Get().Sugar().Errorf("Consumer error: %v", err)
				continue
			}

			value := msg.Value
			timestamp := time.Now().Format("20060102_150405")
			msgID := fmt.Sprintf("%s_%d", timestamp, msg.TopicPartition.Offset)

			logger.Get().Sugar().Infof("Received message from requests topic: %s", msgID)

			if err := s.hdfs.WriteRequestData(ctx, value, msgID); err != nil {
				logger.Get().Sugar().Errorf("Failed to write to HDFS: %v", err)
			} else {
				logger.Get().Sugar().Infof("Successfully saved to HDFS: %s", msgID)
			}

			_, err = kafkaConsumer.CommitMessage(msg)
			if err != nil {
				logger.Get().Sugar().Warnf("Failed to commit message: %v", err)
			}

			go s.sendRecommendation(ctx, value)
		}
	}
}

// sendRecommendation генерирует и отправляет персонализированную рекомендацию в Kafka.
func (s *AnalyticsService) sendRecommendation(ctx context.Context, requestData []byte) {
	data := requestData

	var strData string
	decodedData := data

	if err := json.Unmarshal(data, &strData); err == nil {
		logger.Get().Sugar().Infof("Request is string-encoded: %s", strData)
		decodedBytes, err := base64.StdEncoding.DecodeString(strings.Trim(strData, "\""))
		if err == nil {
			decodedData = decodedBytes
			logger.Get().Sugar().Infof("Decoded base64 data: %s", string(decodedBytes))
		}
	}

	var request Request
	if err := json.Unmarshal(decodedData, &request); err != nil {
		logger.Get().Sugar().Errorf("Failed to unmarshal request: %v, data: %s", err, string(decodedData))
		return
	}

	recommendation := Recommendation{
		ProductID:    "rec_" + fmt.Sprintf("%d", time.Now().UnixNano()),
		ProductName:  request.ProductName + " - Recommended",
		Reason:       "Based on your search history",
		Confidence:   0.85,
		Alternatives: []string{"alt_1", "alt_2", "alt_3"},
		Timestamp:    time.Now().Format(time.RFC3339),
	}

	if err := SendRecommendation(ctx, s.producer, s.cfg.GetRecommendationsTopic(), recommendation); err != nil {
		logger.Get().Sugar().Errorf("Failed to send recommendation: %v", err)
	} else {
		logger.Get().Sugar().Infof("Successfully sent recommendation for product: %s", request.ProductName)
	}
}

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
