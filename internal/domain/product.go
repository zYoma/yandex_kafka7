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

func (e ErrAppStopped) Error() string {
	return "application stopped"
}

type ConsumerConfig interface {
	GetConsumerConfig() *kafkago.ConfigMap
	GetProducerConfig() *kafkago.ConfigMap
	GetBootstrapServers() string
	GetRequestsTopic() string
	GetRecommendationsTopic() string
}

type AnalyticsService struct {
	producer interfaces.Producer
	hdfs     interfaces.HDFSClient
	cfg      ConsumerConfig
}

func NewAnalyticsService(producer interfaces.Producer, hdfs interfaces.HDFSClient, cfg ConsumerConfig) *AnalyticsService {
	return &AnalyticsService{
		producer: producer,
		hdfs:     hdfs,
		cfg:      cfg,
	}
}

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
