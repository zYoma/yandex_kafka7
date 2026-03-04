// Package hdfs предоставляет клиент для работы с HDFS через библиотеку colinmarc/hdfs/v2.
package hdfs

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/colinmarc/hdfs/v2"
	"github.com/zYoma/yandex_kafka7/internal/logger"
)

// HDFSWriter реализует запись данных в HDFS через нативный клиент.
type HDFSWriter struct {
	client    *hdfs.Client
	basePath  string
	batchSize int
}

// HDFSWriterConfig содержит конфигурацию для подключения к HDFS.
type HDFSWriterConfig struct {
	Addresses string
	Port      string
	BasePath  string
	BatchSize int
}

// NewHDFSWriter создает новый клиент HDFS с указанной конфигурацией.
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

// WriteRequestData записывает данные запроса в файл в HDFS с уникальным именем на основе timestamp.
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

// Close закрывает соединение с HDFS.
func (w *HDFSWriter) Close() error {
	if w.client != nil {
		return w.client.Close()
	}
	return nil
}

func (w *HDFSWriter) ReadRecommendations(ctx context.Context) ([]byte, error) {
	recPath := path.Join(w.basePath, "recommendations/part-00000")
	file, err := w.client.Open(recPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open recommendations file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read recommendations: %w", err)
	}

	return data, nil
}

func (w *HDFSWriter) ListFiles(ctx context.Context, dirPath string) ([]string, error) {
	entries, err := w.client.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, path.Join(dirPath, entry.Name()))
		}
	}

	return files, nil
}

func (w *HDFSWriter) ReadFile(ctx context.Context, filePath string) ([]byte, error) {
	file, err := w.client.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}
