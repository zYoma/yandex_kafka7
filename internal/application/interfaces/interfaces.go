// Package interfaces определяет абстракции для работы с Kafka и HDFS.
package interfaces

import (
	"context"
)

type Producer interface {
	SendMessages(ctx context.Context, messages []interface{}) error
	SetTopic(topic string)
}

type Consumer interface {
	StartBatchMessage(ctx context.Context) error
	ReadOneMessage(ctx context.Context, topic string) ([]byte, error)
}

type HDFSClient interface {
	WriteRequestData(ctx context.Context, data []byte, filename string) error
	ReadRecommendations(ctx context.Context) ([]byte, error)
	ListFiles(ctx context.Context, path string) ([]string, error)
	ReadFile(ctx context.Context, filePath string) ([]byte, error)
	Close() error
}

type MessageProcessor interface {
	ProcessMessage(ctx context.Context, data []byte) error
}
