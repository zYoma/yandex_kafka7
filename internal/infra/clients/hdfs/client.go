package hdfs

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/colinmarc/hdfs/v2"
	"github.com/zYoma/yandex_kafka7/internal/logger"
)

type HDFSWriter struct {
	client    *hdfs.Client
	basePath  string
	batchSize int
}

type HDFSWriterConfig struct {
	Addresses string
	Port      string
	BasePath  string
	BatchSize int
}

func NewHDFSWriter(cfg *HDFSWriterConfig) (*HDFSWriter, error) {
	client, err := hdfs.NewClient(hdfs.ClientOptions{
		Addresses: strings.Split(cfg.Addresses, ","),
		User:      "hadoop",
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании клиента HDFS: %w", err)
	}

	err = client.MkdirAll(cfg.BasePath, 0777)
	if err != nil && !strings.Contains(err.Error(), "exists") {
		client.Close()
		return nil, fmt.Errorf("ошибка при создании директории %s: %w", cfg.BasePath, err)
	}

	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 1000
	}

	return &HDFSWriter{
		client:    client,
		basePath:  cfg.BasePath,
		batchSize: batchSize,
	}, nil
}

func (w *HDFSWriter) WriteRequestData(ctx context.Context, data []byte, filename string) error {
	logger.Get().Sugar().Infof("Запись данных запроса в HDFS: %s", filename)

	timestamp := time.Now().Format("20060102_150405_000")
	fullFilename := path.Join(w.basePath, "requests", fmt.Sprintf("%s_%s.json", filename, timestamp))

	if err := w.client.MkdirAll(path.Join(w.basePath, "requests"), 0777); err != nil {
		return fmt.Errorf("не удалось создать директорию requests в HDFS: %w", err)
	}

	file, err := w.client.Create(fullFilename)
	if err != nil {
		return fmt.Errorf("ошибка при создании файла %s в HDFS: %w", fullFilename, err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("ошибка при записи данных в файл %s: %w", fullFilename, err)
	}

	logger.Get().Sugar().Infof("Успешно записано %d байт в HDFS: %s", len(data), fullFilename)
	return nil
}

func (w *HDFSWriter) Close() error {
	if w.client != nil {
		return w.client.Close()
	}
	return nil
}
