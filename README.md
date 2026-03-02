# Yandex Kafka Проект

Учебный проект с двумя кластерами Kafka и Schema Registry с автоматической репликацией данных.

## Быстрый старт

```bash
# Запуск всех сервисов
docker compose up -d

# Подождать инициализации сервисов (~25 секунд)

# Настройка топиков, пользователей и ACL для обоих кластеров
./scripts/setup-all.sh
./scripts/setup-target-cluster-acls.sh

# (Опционально) Тестирование репликации
./scripts/test-replication.sh
```

## Сервисы

### Кластеры Kafka
- **Кластер 1 (Source)**: kafka-0 (9094), kafka-1 (9095), kafka-2 (9096)
- **Кластер 2 (Target)**: kafka-3 (9097), kafka-4 (9098), kafka-5 (9099)

Оба кластера настроены с:
- SASL_SSL аутентификацией
- PLAIN механизмом
- ACL авторизацией включена

### Schema Registry
- URL: http://localhost:8081
- Подключен к: Кластер 1 (Source)
- Хранит схемы в топике `_schemas`

### MirrorMaker 2
- Реплицирует данные из Кластер 1 в Кластер 2
- Реплицирует топики автоматически с настройкой `.*` Regex
- Создаёт топики в целевом кластере с такими же параметрами
- Реплицирует метаданные о топиках и конфигурации

## Пользователи и права доступа

| Пользователь    | Пароль        | Роль                  | Доступ                                   |
|-----------------|---------------|-----------------------|------------------------------------------|
| admin           | admin-secret  | Суперпользователь     | Все операции, управление                 |
| user            | user-secret   | Интер-брокер          | Необходим для внутренней репликации Kafka |
| shop_client     | shop-secret   | Продюсер              | WRITE в `shops_data`                      |
| analytic_client | analytic-secret| Консьюмер           | READ из `shops_data`                      |
| filter_client   | filter-secret  | Консьюмер           | READ из `shops_data`                      |

## Топики

| Топик      | Кластер   | Партиции | Репликация | Продюсер      | Консьюмеры                         |
|------------|-----------|-----------|------------|---------------|-------------------------------------|
| shops_data | Both      | 3         | 3          | shop_client   | analytic_client, filter_client      |

Все топики из Кластер 1 автоматически реплицируются в Кластер 2 с помощью MirrorMaker 2.

## Репликация данных

### Как работает репликация

1. **MirrorMaker 2** реплицирует все топики из Кластер 1 (Source) в Кластер 2 (Target)
2. Репликация происходит в реальном времени (с небольшими задержками)
3. Конфигурация: `source->target.topics = .*` (все топики)
4. Конфигурация: `source->target.replication.policy.class = org.apache.kafka.connect.mirror.IdentityReplicationPolicy` (сохраняет имена топиков)

### Топики репликации

MirrorMaker 2 автоматически реплицирует:
- Пользовательские топики (например, `shops_data`)
- Системные топики (например, `_schemas`)
- Создаёт служебные топики для отслеживания репликации

### Проверка репликации

```bash
# Создать тестовое сообщение в кластере 1
./scripts/test-replication.sh

# Проверить сообщения в кластере 2 (целевой)
docker exec yandex_kafka7-kafka-3-1 kafka-console-consumer.sh \
  --bootstrap-server kafka-3:9094 \
  --topic shops_data \
  --from-beginning \
  --max-messages 10 \
  --consumer.config /opt/bitnami/kafka/config/client.properties

# Список всех топиков в целевом кластере
docker exec yandex_kafka7-kafka-3-1 kafka-topics.sh \
  --bootstrap-server kafka-3:9091 \
  --list \
  --command-config /opt/bitnami/kafka/config/client.properties

# Статус репликации
docker compose logs mirror-maker | tail -20
```

## Скрипты

### scripts/setup-all.sh
Единый скрипт настройки для первого кластера, который:
1. Создаёт конфигурацию клиента
2. Настраивает внутренние ACL Kafka
3. Настраивает ACL для Schema Registry
4. Настраивает интер-брокер коммуникацию
5. Создаёт топики с правильными ACL
6. Показывает все настройки

### scripts/setup-target-cluster-acls.sh
Скрипт настройки для второго (целевого) кластера, который:
1. Настраивает ACL для admin на целевом кластере
2. Настраивает интер-брокер коммуникацию для целевого кластера
3. Настраивает права для MirrorMaker 2

### scripts/test-replication.sh
Скрипт тестирования репликации данных между кластерами, который:
1. Создаёт тестовые сообщения в исходном кластере
2. Читает сообщения из исходного кластера

## Конфигурация безопасности

- **Протокол**: SASL_SSL
- **Механизм**: PLAIN
- **SSL**: Вся внутренняя и внешняя коммуникация зашифрована
- **ACL**: Строгий режим (deny by default)

## Разработка

Кластер настроен для учебных целей с:
- Чётким разделением ролей продюсер/консьюмер
- Правильным enforcement'ом ACL
- Интеграцией Schema Registry
- Мульти-кластерной настройкой для тестирования репликации

## Примеры подключения

### Продюсер (shop_client)
```bash
security.protocol=SASL_SSL
sasl.mechanism=PLAIN
sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username="shop_client" password="shop-secret";
ssl.truststore.location=/path/to/kafka.truststore.jks
ssl.truststore.password=password
ssl.endpoint.identification.algorithm=
bootstrap.servers=localhost:9094,localhost:9095,localhost:9096
```

### Консьюмер (analytic_client/filter_client)
```bash
security.protocol=SASL_SSL
sasl.mechanism=PLAIN
sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username="analytic_client" password="analytic-secret";
ssl.truststore.location=/path/to/kafka.truststore.jks
ssl.truststore.password=password
ssl.endpoint.identification.algorithm=
bootstrap.servers=localhost:9094,localhost:9095,localhost:9096
```

## Устранение проблем

Проверить логи:
```bash
docker compose logs kafka-0
docker compose logs schema-registry
```

Проверить ACL:
```bash
docker exec yandex_kafka7-kafka-0-1 kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --list
```

Проверить репликацию:
```bash
# Логи MirrorMaker
docker compose logs mirror-maker

# Ловг целевого кластера
docker compose logs kafka-3

# Список топиков в обоих кластерах
docker exec yandex_kafka7-kafka-0-1 kafka-topics.sh --bootstrap-server kafka-0:9091 --list --command-config /opt/bitnami/kafka/config/client.properties
docker exec yandex_kafka7-kafka-3-1 kafka-topics.sh --bootstrap-server kafka-3:9091 --list --command-config /opt/bitnami/kafka/config/client.properties
```

## Остановка сервисов

```bash
docker compose down
```
