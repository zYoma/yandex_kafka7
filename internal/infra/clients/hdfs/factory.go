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
