// Package hdfs предоставляет реализацию клиентов для работы с HDFS.
package hdfs

import (
	"context"

	"github.com/zYoma/yandex_kafka7/internal/application/config"
	"github.com/zYoma/yandex_kafka7/internal/application/interfaces"
)

// NewHDFSClient создает клиент HDFS или MockHDFSWriter в зависимости от конфигурации.
func NewHDFSClient(cfg *config.Config) (interfaces.HDFSClient, error) {
	hdfsCfg := &HDFSWriterConfig{
		Addresses: cfg.HDFSHDFSAddresses,
		Port:      cfg.HDFSWebHDFSPort,
		BasePath:  cfg.HDFSKafkaDataPath,
		BatchSize: 1000,
	}

	if hdfsCfg.Addresses == "" {
		return &MockHDFSWriter{}, nil
	}

	return NewWebHDFSWriter(hdfsCfg)
}

// MockHDFSWriter является mock-реализацией клиента HDFS для тестирования.
type MockHDFSWriter struct{}

// WriteRequestData mock-метод для записи данных запроса в HDFS.
func (m *MockHDFSWriter) WriteRequestData(_ context.Context, _ []byte, _ string) error {
	return nil
}

// Close mock-метод для закрытия соединения с HDFS.
func (m *MockHDFSWriter) Close() error {
	return nil
}

func (m *MockHDFSWriter) ReadRecommendations(_ context.Context) ([]byte, error) {
	mockData := `[
		{"product_name": "智能手机", "search_count": 15},
		{"product_name": "笔记本电脑", "search_count": 12},
		{"product_name": "耳机", "search_count": 10}
	]`
	return []byte(mockData), nil
}

func (m *MockHDFSWriter) ListFiles(_ context.Context, path string) ([]string, error) {
	return []string{
		path + "/request1.json",
		path + "/request2.json",
	}, nil
}

func (m *MockHDFSWriter) ReadFile(_ context.Context, filePath string) ([]byte, error) {
	mockRequest := `{"product_name": "test_product", "timestamp": "2024-01-01T00:00:00Z"}`
	return []byte(mockRequest), nil
}
