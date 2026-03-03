package application

import (
	"context"

	"github.com/zYoma/yandex_kafka7/internal/application/interfaces"
)

type AnalyticsApp struct {
	producer     interfaces.Producer
	hdfs         interfaces.HDFSClient
	msgProcessor interfaces.MessageProcessor
	recTopic     string
}

func NewAnalyticsApp(producer interfaces.Producer, hdfs interfaces.HDFSClient, msgProcessor interfaces.MessageProcessor, recommendationsTopic string) *AnalyticsApp {
	return &AnalyticsApp{
		producer:     producer,
		hdfs:         hdfs,
		msgProcessor: msgProcessor,
		recTopic:     recommendationsTopic,
	}
}

func (a *AnalyticsApp) ProcessRequest(ctx context.Context, requestID string, data []byte) error {
	if err := a.hdfs.WriteRequestData(ctx, data, requestID); err != nil {
		return err
	}

	a.msgProcessor.ProcessMessage(ctx, data)

	return nil
}
