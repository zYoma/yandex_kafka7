package domain

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/zYoma/yandex_kafka7/internal/application/interfaces"
	"github.com/zYoma/yandex_kafka7/internal/logger"
)

type RecommendationService struct {
	producer      interfaces.Producer
	recTopic      string
	hdfs          interfaces.HDFSClient
	popularCache  []PopularProduct
	cacheUpdateMu sync.RWMutex
}

func NewRecommendationService(producer interfaces.Producer, recTopic string, hdfs interfaces.HDFSClient) *RecommendationService {
	rs := &RecommendationService{
		producer:     producer,
		recTopic:     recTopic,
		hdfs:         hdfs,
		popularCache: []PopularProduct{},
	}
	rs.loadPopularProducts()
	return rs
}

func (s *RecommendationService) loadPopularProducts() {
	ctx := context.Background()
	data, err := s.hdfs.ReadRecommendations(ctx)
	if err != nil {
		logger.Get().Sugar().Warnf("Failed to load recommendations: %v", err)
		return
	}

	var popular []PopularProduct
	json.Unmarshal(data, &popular)

	s.cacheUpdateMu.Lock()
	s.popularCache = popular
	s.cacheUpdateMu.Unlock()

	logger.Get().Sugar().Infof("Loaded %d popular products", len(popular))
}

func (s *RecommendationService) ProcessMessage(ctx context.Context, data []byte) error {
	var strData string
	decodedData := data

	if err := json.Unmarshal(data, &strData); err == nil {
		decodedBytes, err := base64.StdEncoding.DecodeString(strings.Trim(strData, "\""))
		if err == nil {
			decodedData = decodedBytes
		}
	}

	var request Request
	if err := json.Unmarshal(decodedData, &request); err != nil {
		logger.Get().Sugar().Errorf("Failed to unmarshal request: %v", err)
		return err
	}

	recProduct := s.getRecommendedProduct(request.ProductName)

	recommendation := Recommendation{
		ProductID:    getRecommendationID(recProduct),
		ProductName:  recProduct,
		Reason:       "Popular recommendation",
		Confidence:   0.9,
		Alternatives: s.getAlternatives(request.ProductName),
		Timestamp:    time.Now().Format(time.RFC3339),
	}

	return SendRecommendation(ctx, s.producer, s.recTopic, recommendation)
}

func (s *RecommendationService) getRecommendedProduct(searchTerm string) string {
	s.cacheUpdateMu.RLock()
	defer s.cacheUpdateMu.RUnlock()

	if len(s.popularCache) > 0 {
		for _, popular := range s.popularCache {
			if containsIgnoreCase(searchTerm, popular.ProductName) {
				return popular.ProductName + " - Popular"
			}
		}
		return s.popularCache[0].ProductName + " - Trending"
	}

	return searchTerm + " - Recommended"
}

func (s *RecommendationService) getAlternatives(searchTerm string) []string {
	s.cacheUpdateMu.RLock()
	defer s.cacheUpdateMu.RUnlock()

	alt := []string{}
	for i := 0; i < len(s.popularCache) && i < 3; i++ {
		if s.popularCache[i].ProductName != searchTerm {
			alt = append(alt, s.popularCache[i].ProductName)
		}
	}

	if len(alt) == 0 {
		return []string{"alt_1", "alt_2", "alt_3"}
	}

	return alt
}

func getRecommendationID(productName string) string {
	return "rec_" + productName
}

func containsIgnoreCase(str, substr string) bool {
	return strings.Contains(
		strings.ToLower(str),
		strings.ToLower(substr),
	)
}
