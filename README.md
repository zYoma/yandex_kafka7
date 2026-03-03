# Аналитическая платформа для маркетплейса

Проект представляет собой полнофункциональную аналитическую платформу для маркетплейса на базе Apache Kafka. Платформа обрабатывает данные от магазинов, формирует персонализированные рекомендации и предоставляет клиентам интерфейс для поиска товаров.

## Назначение проекта

Аналитическая платформа решает следующие задачи:
- Сбор информации о товарах от магазинов через SHOP API
- Обработка поисковых запросов клиентов через CLIENT API
- Фильтрация товаров на основе списка запрещенных продуктов
- Хранение и поиск данных через Kafka Connect
- Аналитическая обработка данных и формирование рекомендаций
- Мониторинг всех компонентов инфраструктуры

## Используемые технологии

### Основные компоненты
- **Apache Kafka 3.x** — распределённая платформа потоковой обработки данных
- **Schema Registry** — управление схемами данных в формате Avro
- **MirrorMaker 2** — репликация данных между кластерами Kafka
- **Kafka Connect** — интеграция с внешними системами

### Безопасность
- **SASL_SSL** — безопасная передача данных с SSL-шифрованием
- **PLAIN механизм** — аутентификация пользователей
- **ACL** — управление доступом к топикам и операциям

### Аналитика и хранение данных
- **HDFS** — хранение больших объемов данных для аналитики
- **Kafka Connect** — потоковая обработка и фильтрация данных

### Мониторинг
- **Prometheus** — сбор метрик с брокеров Kafka
- **Grafana** — визуализация метрик на дашбордах
- **Alertmanager** — обработка алертов и отправка уведомлений
- **JMX Exporter** — экспорт метрик JVM Kafka

### Языки и библиотеки
- **Go** — реализация SHOP API, CLIENT API и аналитического сервиса
- **confluent-kafka-go** — официальный клиент Kafka для Go
- **zap** — структурированное логирование

## Архитектура проекта

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           Магазины (Клиенты)                             │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                              SHOP API                                    │
│                     (cmd/shop/main.go)                                  │
│  Чтение товаров из JSON файла → Отправка в Kafka                        │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    Кластер Kafka 1 (Source)                             │
│  ┌─────────────┬─────────────┬─────────────┐                           │
│  │  kafka-0    │  kafka-1    │  kafka-2    │                           │
│  │  (9094)     │  (9095)     │  (9096)     │                           │
│  └─────────────┴─────────────┴─────────────┘                           │
│                                                                         │
│  Топики:                                                                │
│  • shops_data — товары от магазинов                                     │
│  • requests — поисковые запросы клиентов                                │
│  • recommendations — персональные рекомендации                          │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                    ┌───────────────┴───────────────┐
                    │                               │
                    ▼                               ▼
┌───────────────────────────┐      ┌──────────────────────────────────────┐
│    Kafka Connect          │      │     MirrorMaker 2                   │
│  (Потоковая обработка)    │      │  (Репликация данных)                │
│  • Фильтрация товаров     │      │                                      │
│  • Запись в файл          │      │                                      │
└───────────────────────────┘      └──────────────────────────────────────┘
                    │                               │
                    ▼                               ▼
┌───────────────────────────┐    ┌──────────────────────────────────────────┐
│   Файл данных             │    │  Кластер Kafka 2 (Target)              │
│  filtered-products.txt    │    │  ┌────────┬────────┬────────┐           │
└───────────────────────────┘    │  │kafka-3 │kafka-4 │kafka-5 │           │
                                 │  └────────┴────────┴────────┘           │
                                 └──────────────────────────────────────────┘
                                                   │
                                                   ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    Аналитическая система                                │
│                   (cmd/analytic_client/main.go)                         │
│  Чтение запросов → Сохранение в HDFS → Генерация рекомендаций          │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    Топик recommendations                                 │
│                      (в обоих кластерах)                                 │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                              CLIENT API                                  │
│                      (cmd/client/main.go)                               │
│  • Поиск товаров по имени                                               │
│  • Получение персональных рекомендаций                                 │
└─────────────────────────────────────────────────────────────────────────┘
```

## Инструкция по запуску

### Предварительные требования

- Docker (версия 20.x и выше)
- Docker Compose (версия 2.x и выше)
- Go (версия 1.21 и выше) — для сборки приложений
- Не менее 8 GB свободной оперативной памяти

### Шаг 1. Клонирование репозитория и сборка приложений

```bash
# Клонирование репозитория
git clone <repository_url>
cd yandex_kafka7

# Сборка Go приложений
go build -o bin/shop ./cmd/shop
go build -o bin/analytic_client ./cmd/analytic_client
go build -o bin/client ./cmd/client
```

### Шаг 2. Запуск инфраструктуры Kafka

```bash
# Запуск всех сервисов (Kafka, Schema Registry, MirrorMaker)
docker compose up -d

# Ожидание инициализации сервисов (~25-30 секунд)
sleep 30

# Проверка статуса сервисов
docker compose ps
```

### Шаг 3. Настройка первого кластера Kafka (Source)

```bash
# Настройка топиков, пользователей и ACL для первого кластера
./scripts/setup-all.sh
```

**Что делает скрипт:**
- Создает пользователей Kafka (admin, user, shop_client, analytic_client, filter_client, client)
- Создает топики: `shops_data`, `requests`, `recommendations`
- Настраивает ACL права для каждого пользователя
- Настраивает минимальное количество синхронных реплик (min.insync.replicas = 2)

### Шаг 4. Настройка второго кластера Kafka (Target)

```bash
# Настройка ACL для второго кластера
./scripts/setup-target-cluster-acls.sh
```

**Что делает скрипт:**
- Создает пользователей во втором кластере
- Настраивает ACL права для аналитического клиента
- Проверяет репликацию топиков из первого кластера

### Шаг 5. Настройка Schema Registry

```bash
# Регистрация Avro схемы для товаров
./scripts/schema-manager.sh register

# Проверка зарегистрированных схем
./scripts/schema-manager.sh list
```

**Что делает скрипт:**
- Регистрирует схему товара в Schema Registry
- Позволяет просматривать версии схемы

### Шаг 6. Настройка Kafka Connect

```bash
# Настройка ACL и топиков для Kafka Connect
./scripts/setup-kafka-connect.sh

# Перезапуск Connect для загрузки плагинов
docker compose restart kafka-connect

# Ожидание запуска (~10 секунд)
sleep 10

# Развертывание коннектора для фильтрации товаров
./scripts/connector-manager.sh deploy
```

**Что делают скрипты:**
- `setup-kafka-connect.sh`: создает пользователей и ACL для Connect
- `connector-manager.sh deploy`: настраивает коннектор для чтения из `shops_data`, фильтрации по списку разрешенных товаров и записи в файл

### Шаг 7. Управление списком разрешенных товаров

```bash
# Просмотр списка разрешенных product_id
./scripts/manage-products.sh list

# Добавление товара в разрешенный список
./scripts/manage-products.sh add "12345"

# Удаление товара из списка
./scripts/manage-products.sh remove "12345"
```

После изменения списка нужно перезапустить Kafka Connect:
```bash
docker compose restart kafka-connect
```

### Шаг 8. Запуск SHOP API (Отправка товаров)

```bash
# Отправка товаров из файла в Kafka
docker compose run --rm shop /bin/shop
```

**Что делает SHOP API:**
- Читает товары из файла `products_source.json`
- Отправляет данные в топик `shops_data` в формате Avro
- Данные проходят через Schema Registry

### Шаг 9. Запуск аналитического сервиса

```bash
# Запуск аналитического консьюмера
docker compose run --rm analytic_client /bin/analytic_client
```

**Что делает аналитический сервис:**
- Читает поисковые запросы из топика `requests`
- Сохраняет запросы в HDFS для аналитики
- Генерирует персонализированные рекомендации
- Отправляет рекомендации в топик `recommendations`

### Шаг 10. Использование CLIENT API

```bash
# Поиск товара по имени
docker compose run --rm client /bin/client search "часы"

# Получение персонализированных рекомендаций
docker compose run --rm client /bin/client recommendations
```

**Что делает CLIENT API:**
- `search`: ищет товары в файле отфильтрованных данных, отправляет запрос в Kafka для аналитики
- `recommendations`: читает рекомендацию из топика `recommendations`

### Шаг 11. Проверка репликации данных

```bash
# Тестирование репликации между кластерами
./scripts/test-replication.sh

# Проверка сообщений во втором кластере
docker exec yandex_kafka7-kafka-3-1 kafka-console-consumer.sh \
  --bootstrap-server kafka-3:9094 \
  --topic shops_data \
  --from-beginning \
  --max-messages 5 \
  --consumer.config /opt/bitnami/kafka/config/client.properties
```

### Шаг 12. Проверка мониторинга

```bash
# Проверка метрик JMX
curl http://localhost:7071/metrics

# Проверка статуса Prometheus
curl http://localhost:9090/api/v1/targets

# Проверка алертов в Prometheus
curl http://localhost:9090/api/v1/alerts
```

**Доступ к дашбордам:**
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (логин: `admin`, пароль: `admin`)
- **Alertmanager**: http://localhost:9093

## Проверка выполнения критериев

### ✓ Kafka успешно передаёт данные между сервисами

**Проверка:**
```bash
# Проверить сообщения в топике shops_data
docker exec yandex_kafka7-kafka-0-1 kafka-console-consumer.sh \
  --bootstrap-server kafka-0:9094 \
  --topic shops_data \
  --from-beginning \
  --max-messages 5 \
  --consumer.config /opt/bitnami/kafka/config/client.properties
```

### ✓ Включена защита TLS и работают ACL

**Проверка ACL:**
```bash
# Проверить ACL для топика shops_data
docker exec yandex_kafka7-kafka-0-1 kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --list \
  --command-config /opt/bitnami/kafka/config/client.properties
```

**Пользователи и права:**
| Пользователь    | Пароль        | Роль            | Доступ                                      |
|-----------------|---------------|-----------------|---------------------------------------------|
| admin           | admin-secret  | Суперпользователь | Все операции, управление                     |
| shop_client     | shop-secret   | Продюсер         | WRITE в `shops_data`                         |
| analytic_client | analytic-secret| Консьюмер/Продюсер | READ из `requests`, WRITE в `recommendations`|
| filter_client   | filter-secret  | Консьюмер        | READ из `shops_data`                         |
| client          | client-secret  | CLI Пользователь  | READ из `recommendations`, WRITE в `requests`|

### ✓ Реализована репликация топиков и задано минимальное число реплик

**Проверка конфигурации топика:**
```bash
docker exec yandex_kafka7-kafka-0-1 kafka-topics.sh \
  --bootstrap-server kafka-0:9091 \
  --topic shops_data \
  --describe \
  --command-config /opt/bitnami/kafka/config/client.properties
```

Ожидаемый вывод:
```
Topic: shops_data	PartitionCount: 3	ReplicationFactor: 3	min.insync.replicas: 2
```

### ✓ Настроено дублирование данных во второй Kafka-кластер

**Проверка топиков во втором кластере:**
```bash
docker exec yandex_kafka7-kafka-3-1 kafka-topics.sh \
  --bootstrap-server kafka-3:9091 \
  --list \
  --command-config /opt/bitnami/kafka/config/client.properties
```

**Проверка логов MirrorMaker:**
```bash
docker compose logs mirror-maker | grep -i "replication"
```

### ✓ Выполнена фильтрация запрещенных товаров

**Проверка отфильтрованных данных:**
```bash
# Просмотр файла с отфильтрованными товарами
cat kafka-connect/output/filtered-products.txt

# Проверка статуса коннектора
./scripts/connector-manager.sh status
```

Фильтрация выполняется через Kafka Connect SMT (Single Message Transformer) на основе списка разрешенных product_id.

### ✓ Данные после фильтрации записываются в систему хранения

**Реализация:** Базовый вариант — запись в файл через Kafka Connect

**Проверка:**
```bash
# Проверка файла с данными
ls -lh kafka-connect/output/filtered-products.txt

# Просмотр содержимого
cat kafka-connect/output/filtered-products.txt
```

### ✓ Реализована аналитическая обработка данных

**Реализация:** Базовый вариант — перенос данных в HDFS, аналитика рекомендаций

**Проверка:**
```bash
# Проверка логов аналитического сервиса
docker compose logs analytic_client

# Поиск запросов сохраненных в HDFS (если настроен)
docker exec yandex_kafka7-hdfs-1 hdfs dfs -ls /kafka_data/requests
```

### ✓ Рекомендации записываются в отдельный топик Kafka

**Проверка:**
```bash
# Чтение рекомендаций из топика
docker exec yandex_kafka7-kafka-0-1 kafka-console-consumer.sh \
  --bootstrap-server kafka-0:9094 \
  --topic recommendations \
  --from-beginning \
  --max-messages 3 \
  --consumer.config /opt/bitnami/kafka/config/client.properties
```

### ✓ Настроен мониторинг

**Дашборд Grafana:**
1. Откройте http://localhost:3000
2. Логин: `admin`, пароль: `admin`
3. Перейдите в дашборд "Kafka Monitoring Dashboard"

**Панели дашборда:**
- Статус брокеров (онлайн/офлайн)
- Скорость передачи данных (Bytes In/Out)
- Запросы продюсеров/консьюмеров
- Недореплицированные партиции
- Офлайн партиции
- Использование JVM памяти
- Загрузка request handler'ов
- Статус активного контроллера

**Алерты:**
- `KafkaBrokerDown` (critical) — брокер недоступен > 1 мин
- `KafkaBrokerOfflinePartitionsCount` (critical) — офлайн партиции > 0
- `KafkaBrokerActiveControllerCount` (critical) — нет активного контроллера > 2 мин
- `KafkaBrokerUnderReplicatedPartitions` (warning) — недореплицированные партиции > 0
- `KafkaBrokerRequestHandlerIdle` (warning) — idle < 10% > 10 мин
- `KafkajvmMemoryHigh` (warning) — память > 90% > 5 мин

**Метрики собираются через:**
- Prometheus (сбор метрик)
- JMX Exporter (экспорт метрик JVM Kafka с портов 7071-7076)

## Структура проекта

```
yandex_kafka7/
├── cmd/                          # Точки входа приложений
│   ├── shop/                     # SHOP API — отправка товаров
│   ├── analytic_client/          # Аналитический сервис
│   └── client/                   # CLIENT API — поиск и рекомендации
├── internal/
│   ├── application/              # Логика приложений
│   │   ├── app.go               # ProducerApp и ConsumerApp
│   │   ├── config/              # Конфигурация из переменных окружения
│   │   └── interfaces.go        # Интерфейсы для зависимостей
│   ├── domain/                   # Бизнес-логика и сущности
│   │   └── product.go           # Product, Recommendation, Request
│   ├── infra/                    # Инфраструктурный слой
│   │   └── clients/
│   │       ├── kafka/           # Kafka продюсер и консьюмер
│   │       └── hdfs/            # HDFS клиенты
│   ├── cli/                      # CLI клиент
│   └── logger/                   # Логгер на базе zap
├── kafka-connect/                # Kafka Connect плагин фильтрации
│   ├── ProductFilterTransformer.java
│   └── plugins/lib/             # Скомпилированный JAR
├── scripts/                     # Скрипты настройки
│   ├── setup-all.sh
│   ├── setup-target-cluster-acls.sh
│   ├── setup-kafka-connect.sh
│   ├── connector-manager.sh
│   ├── manage-products.sh
│   ├── schema-manager.sh
│   └── test-replication.sh
├── monitoring/                  # Конфигурации мониторинга
│   ├── prometheus.yml
│   ├── alerts.yml
│   ├── alertmanager.yml
│   ├── grafana-dashboard.json
│   └── jmx_exporter_config.yml
├── products_source.json        # Источник данных товаров
├── kafka-connect/output/       # Файлы с отфильтрованными данными
├── docker-compose.yml          # Композиция сервисов
└── README.md                   # Этот файл
```

## Остановка сервисов

```bash
# Остановка всех сервисов
docker compose down

# Остановка с удалением томов (перезапуск с чистого листа)
docker compose down -v
```

## Устранение проблем

### Брокеры не запускаются

```bash
# Проверить логи брокеров
docker compose logs kafka-0 kafka-1 kafka-2

# Проверить инициализацию кластера
docker compose logs zookeeper-server
```

### MirrorMaker не реплицирует данные

```bash
# Проверить логи MirrorMaker
docker compose logs mirror-maker

# Проверить ACL во втором кластере
docker exec yandex_kafka7-kafka-3-1 kafka-acls.sh \
  --bootstrap-server kafka-3:9091 \
  --list \
  --command-config /opt/bitnami/kafka/config/client.properties
```

### Kafka Connect не записывает данные

```bash
# Проверить логи Connect
docker compose logs kafka-connect

# Проверить статус коннектора
./scripts/connector-manager.sh status

# Проверить список разрешенных товаров
./scripts/manage-products.sh list
```

### Нет метрик в Prometheus

```bash
# Проверить работоспособность JMX Exporter
curl http://localhost:7071/metrics

# Проверить логи Prometheus
docker compose logs prometheus

# Проверить статус targets в Prometheus
curl http://localhost:9090/api/v1/targets
```

## Полезные команды

### Работа с Kafka

```bash
# Просмотр топиков
docker exec yandex_kafka7-kafka-0-1 kafka-topics.sh \
  --bootstrap-server kafka-0:9091 \
  --list \
  --command-config /opt/bitnami/kafka/config/client.properties

# Просмотр информации о топике
docker exec yandex_kafka7-kafka-0-1 kafka-topics.sh \
  --bootstrap-server kafka-0:9091 \
  --topic shops_data \
  --describe \
  --command-config /opt/bitnami/kafka/config/client.properties

# Чтение сообщений из топика
docker exec yandex_kafka7-kafka-0-1 kafka-console-consumer.sh \
  --bootstrap-server kafka-0:9094 \
  --topic shops_data \
  --from-beginning \
  --max-messages 10 \
  --consumer.config /opt/bitnami/kafka/config/client.properties
```

### Проверка ACL

```bash
# Просмотр всех ACL
docker exec yandex_kafka7-kafka-0-1 kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --list \
  --command-config /opt/bitnami/kafka/config/client.properties

# Просмотр ACL для конкретного пользователя
docker exec yandex_kafka7-kafka-0-1 kafka-acls.sh \
  --bootstrap-server kafka-0:9091 \
  --list --principal User:shop_client \
  --command-config /opt/bitnami/kafka/config/client.properties
```

### Мониторинг

```bash
# Запрос метрик через Prometheus API
curl 'http://localhost:9090/api/v1/query?query=kafka_server_replicamanager_underreplicatedpartitions'

curl 'http://localhost:9090/api/v1/query?query=kafka_controller_kafkacontroller_activebrokercount'

# Просмотр активных алертов
curl http://localhost:9090/api/v1/alerts
```

## Дополнительная информация

- Подробная документация по мониторингу: [monitoring/README.md](monitoring/README.md)
- Документация по скриптам: [scripts/README.md](scripts/README.md)
- Конфигурация Kafka Connect: [kafka-connect/README.md](kafka-connect/README.md)