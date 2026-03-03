package hdfs

import (
	"context"

	"github.com/zYoma/yandex_kafka7/internal/application/config"
	"github.com/zYoma/yandex_kafka7/internal/application/interfaces"
)

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

type MockHDFSWriter struct{}

func (m *MockHDFSWriter) WriteRequestData(_ context.Context, _ []byte, _ string) error {
	return nil
}

func (m *MockHDFSWriter) Close() error {
	return nil
}
