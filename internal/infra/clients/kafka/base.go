// Package kafka предоставляет утилиты для работы с Kafka и Schema Registry.
package kafka

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avrov2"
	"github.com/zYoma/yandex_kafka7/internal/application/config"
)

// GetDeserializer создает десериализатор Avro для Schema Registry.
func GetDeserializer(cfg *config.Config) (*avrov2.Deserializer, error) {
	// Конфигурация для Schema Registry
	srClient, err := schemaregistry.NewClient(schemaregistry.NewConfig(cfg.SchemaRegistryServiceURL))
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании клиента Schema Registry: %w", err)
	}

	// Avro deserializer
	deser, err := avrov2.NewDeserializer(srClient, serde.ValueSerde, avrov2.NewDeserializerConfig())
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании десериализатора: %w", err)
	}

	return deser, nil
}

// GetSerializer создает сериализатор Avro для Schema Registry.
func GetSerializer(cfg *config.Config) (*avrov2.Serializer, error) {
	// Конфигурация для Schema Registry
	srClient, err := schemaregistry.NewClient(schemaregistry.NewConfig(cfg.SchemaRegistryServiceURL))
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании клиента Schema Registry: %w", err)
	}

	// Avro serializer
	ser, err := avrov2.NewSerializer(srClient, serde.ValueSerde, avrov2.NewSerializerConfig())
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании сериализатора: %w", err)
	}

	return ser, nil
}
