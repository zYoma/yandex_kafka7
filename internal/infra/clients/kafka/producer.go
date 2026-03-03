// Package kafka предоставляет реализацию Kafka продюсера.
package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avrov2"
	"github.com/zYoma/yandex_kafka7/internal/application"
	"github.com/zYoma/yandex_kafka7/internal/application/config"
	"github.com/zYoma/yandex_kafka7/internal/logger"
)

// KafkaProducer реализует интерфейс Producer для отправки сообщений в Kafka.
type KafkaProducer struct {
	Producer     *kafka.Producer
	Serializer   *avrov2.Serializer
	DeliveryChan chan kafka.Event
	Topic        string
}

// NewKafkaProducer создает новый экземпляр KafkaProducer
func NewKafkaProducer(serializer *avrov2.Serializer, cfg *config.Config) (*KafkaProducer, error) {
	// Конфигурация для Kafka Producer
	producer, err := kafka.NewProducer(cfg.GetProducerConfig())
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании продюсера: %w", err)
	}
	deliveryChan := make(chan kafka.Event)
	return &KafkaProducer{Serializer: serializer, Producer: producer, DeliveryChan: deliveryChan, Topic: cfg.Topic}, nil
}

// Stop останавливает продюсер.
func (p *KafkaProducer) Stop() {
	p.Producer.Close()
	close(p.DeliveryChan)
}

// SetTopic устанавливает топик для отправки сообщений.
func (p *KafkaProducer) SetTopic(topic string) {
	p.Topic = topic
}

// SendMessages отправляет пачку сообщений.
func (p *KafkaProducer) SendMessages(ctx context.Context, messages []interface{}) error {

	// Подготовка сообщений для отправки
	var kafkaMessages []*kafka.Message
	for _, msg := range messages {
		var payload []byte
		var err error

		// Попытка сериализовать через Avro Serializer
		if p.Serializer != nil {
			payload, err = p.Serializer.Serialize(p.Topic, msg)
			if err != nil {
				// Если Avro сериализация не удалась, пробуем JSON
				if jsonBytes, jsonErr := json.Marshal(msg); jsonErr == nil {
					payload = jsonBytes
				} else {
					return fmt.Errorf("ошибка при сериализации сообщения: %w", err)
				}
			}
		} else {
			// Если нет сериализатора, используем JSON
			payload, err = json.Marshal(msg)
			if err != nil {
				return fmt.Errorf("ошибка при JSON сериализации сообщения: %w", err)
			}
		}

		kafkaMessage := &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &p.Topic, Partition: kafka.PartitionAny},
			Value:          payload,
		}
		kafkaMessages = append(kafkaMessages, kafkaMessage)
	}

	// Отправка сообщений по одному с обработкой событий
	for _, kafkaMsg := range kafkaMessages {
		for {
			err := p.Producer.Produce(kafkaMsg, p.DeliveryChan)
			if err != nil {
				if kafkaErr, ok := err.(kafka.Error); ok {
					if kafkaErr.Code() == kafka.ErrQueueFull {
						// Очередь полна, ждем 1 секунду и пробуем снова
						time.Sleep(time.Second)
						continue
					}
				}
				return fmt.Errorf("ошибка при отправке сообщения: %w", err)
			}
			break
		}
	}

	// Ждем завершения отправки всех сообщений
	flushTimeout := time.After(30 * time.Second)

	// Создаем счетчик для отслеживания количества отправленных сообщений
	sentCount := len(kafkaMessages)
	deliveredCount := 0

	// Читаем события доставки из канала
	for deliveredCount < sentCount {
		select {
		case <-flushTimeout:
			return fmt.Errorf("таймаут ожидания отправки сообщений")
		case <-ctx.Done():
			return application.ErrAppStopped
		case ev := <-p.DeliveryChan:
			switch e := ev.(type) {
			case *kafka.Message:
				if e.TopicPartition.Error != nil {
					logger.Get().Sugar().Errorf("Delivery failed: %v", e.TopicPartition.Error)
				} else {
					// Обработка успешной доставки
					deliveredCount++
					logger.Get().Sugar().Infof("Сообщение доставлено в топик %s [%d] at offset %v",
						*e.TopicPartition.Topic, e.TopicPartition.Partition, e.TopicPartition.Offset)
				}
			case kafka.Error:
				if e.IsFatal() {
					// при фатальной ошибке, продолжение работы невозможно
					return fmt.Errorf("FATAL ERROR: %v", e)
				}
				// Обработка неудачной доставки
				logger.Get().Sugar().Errorf("Ошибка доставки: %v", e)
			default:
				// Обработка других типов событий
				logger.Get().Sugar().Warnf("Непредвиденное событие доставки: %v", e)
			}
		}
	}

	logger.Get().Info("Все сообщения отправлены!")
	return nil
}
