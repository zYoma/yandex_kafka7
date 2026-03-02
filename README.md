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

Каталог `scripts/` содержит скрипты для настройки Kafka, ACL и управления Kafka Connect.

Подробнее о скриптах см. [scripts/README.md](scripts/README.md)

**Основные скрипты:**
- `setup-all.sh` - настройка первого кластера
- `setup-target-cluster-acls.sh` - настройка второго кластера
- `setup-kafka-connect.sh` - настройка Kafka Connect
- `connector-manager.sh` - управление коннекторами
- `manage-products.sh` - управление списком разрешённых product_id
- `test-replication.sh` - тестирование репликации

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

## Kafka Connect

Kafka Connect сервис настроен для фильтрации данных из топика shops_data_json и записи в файл с использованием Avro формата и Schema Registry.

⚡ **JAR плагин уже скомпилирован и готов к использованию!**

### Настройка Schema Registry

Схемы для товаров хранятся в Schema Registry.

```bash
# 1. Зарегистрировать схему для товаров
./scripts/schema-manager.sh register

# 2. Проверить зарегистрированные схемы
./scripts/schema-manager.sh list

# 3. Получить текущую схему
./scripts/schema-manager.sh get
```

### Настройка Kafka Connect

```bash
# 1. Настроить ACL и топики для Connect
./scripts/setup-kafka-connect.sh

# 2. Перезапустить Connect для загрузки плагина
docker compose restart kafka-connect

# 3. Развернуть коннектор
./scripts/connector-manager.sh deploy
```

### Сборка плагина (при необходимости)

Если вы измените исходный код трансформера или хотите перекомпилировать JAR:

```bash
cd kafka-connect
./build-plugin.sh --force
```

Скрипт `build-plugin.sh` автоматически использует Maven, если он установлен, или Docker с Maven в противном случае.

### Хранение JAR в репозитории

Для учебных projects скомпилированный JAR хранится в репозитории (`kafka-connect/plugins/lib/`), что упрощает использование без установки Maven.

**Преимущества:**
- Не требует установки Maven для обычных пользователей
- Быстрый старт проекта
- Гарантированная совместимость версий

### Конфигурация с Schema Registry

Kafka Connect настроен для использования Avro формата с Schema Registry:

- **Value Converter**: `io.confluent.connect.avro.AvroConverter`
- **Schema Registry URL**: `http://schema-registry:8081`
- **Топик со схемой**: `shops_data_json-value`

Для изменения схемы необходимо:
1. Обновить файл `kafka-connect/product-schema-avro.json`
2. Зарегистрировать новую версию: `./scripts/schema-manager.sh register`
3. Отправить сообщения, соответствующие схеме

### Управление списком разрешенных товаров

CLI скрипт для управления списком product_id, которые пропускаются фильтром:

```bash
# Показать все разрешенные product_id
./scripts/manage-products.sh list

# Добавить product_id
./scripts/manage-products.sh add "99999"

# Удалить product_id
./scripts/manage-products.sh remove "99999"

# Очистить весь список
./scripts/manage-products.sh clear
```

После изменения списка нужно перезапустить Kafka Connect:
```bash
docker compose restart kafka-connect
```

### Управление коннектором

```bash
# Развернуть коннектор
./scripts/connector-manager.sh deploy

# Удалить коннектор
./scripts/connector-manager.sh delete

# Проверить статус коннектора
./scripts/connector-manager.sh status

# Просмотреть логи
docker compose logs kafka-connect

# Просмотреть отфильтрованные данные
cat kafka-connect/output/filtered-products.txt
```

### Топики Kafka Connect

| Топик            | Назначение                           |
|------------------|--------------------------------------|
| connect-configs  | Хранение конфигураций коннекторов    |
| connect-offsets  | Хранение offset'ов коннекторов       |
| connect-status   | Хранение статусов коннекторов        |

## Мониторинг Kafka

Проект включает полнофункциональную систему мониторинга с использованием Prometheus, Grafana и Alertmanager.

### Структура файлов мониторинга

В каталоге `monitoring/` расположены следующие файлы:

| Файл | Назначение |
|------|------------|
| **prometheus.yml** | Конфигурация Prometheus - определяет откуда собирать метрики и интервалы сбора |
| **alerts.yml** | Правила алертов - определяет условия для алертов (упавший брокер, офлайн партиции и т.д.) |
| **alertmanager.yml** | Конфигурация Alertmanager - настройка маршрутизации и отправки уведомлений |
| **grafana-dashboard.json** | Готовый дашборд Grafana с визуализацией метрик Kafka |
| **grafana-datasources.yml** | Конфигурация datasource для Grafana (подключение к Prometheus) |
| **jmx_exporter_config.yml** | Конфигурация JMX Exporter - какие метрики экспортировать из Kafka |
| **Dockerfile.jmx-exporter** | Dockerfile для построения образа JMX Exporter |
| **entrypoint.sh** | Entrypoint скрипт JMX Exporter - запускает экспорт метрик |
| **start-monitoring.sh** | Скрипт для запуска всех сервисов мониторинга |
| **stop-monitoring.sh** | Скрипт для остановки всех сервисов мониторинга |
| **README.md** | Подробная документация по мониторингу |

### Запуск мониторинга

```bash
# Запуск всех сервисов включая мониторинг
docker compose up -d

# Или через специальный скрипт
./monitoring/start-monitoring.sh
```

Мониторинг автоматически запускается вместе с Kafka:
- **JMX Exporters** - подключены к каждому брокеру как sidecar контейнеры
- **Prometheus** - собирает метрики с всех брокеров обоих кластеров
- **Alertmanager** - обрабатывает алерты и отправляет уведомления
- **Grafana** - визуализирует метрики через дашборд

### Доступ к сервисам мониторинга

| Сервис | URL | Логин/Пароль |
|--------|-----|--------------|
| **Prometheus** | http://localhost:9090 | - |
| **Grafana** | http://localhost:3000 | admin/admin |
| **Alertmanager** | http://localhost:9093 | - |

### Основные метрики

**Prometheus запросы:**

```promql
# Статус брокеров
up{job=~"kafka-cluster-.*"}

# Активные брокеры в кластере
kafka_controller_kafkacontroller_activebrokercount

# Недореплицированные партиции
kafka_server_replicamanager_underreplicatedpartitions

# Офлайн партиции
kafka_server_replicamanager_offlinepartitionscount

# Загрузка сети (байты/сек)
rate(kafka_server_brokertopicmetrics_bytesinpersec_total[5m])
rate(kafka_server_brokertopicmetrics_bytesoutpersec_total[5m])

# Использование JVM памяти
(jvm_memory_heap_bytes{area="heap"} / jvm_memory_heap_bytes_max{area="heap"}) * 100

# Speed of requests
rate(kafka_server_brokertopicmetrics_totalproducerequestspersec_total[5m])
```

### Настроенные алерты

Созданы следующие алерты:

| Название | Уровень | Условие | Описание |
|----------|---------|---------|----------|
| **KafkaBrokerDown** | critical | broker недоступен > 1 мин | Брокер упал и не отвечает |
| **KafkaBrokerOfflinePartitionsCount** | critical | офлайн партиции > 0 | Партиции в офлайн состоянии |
| **KafkaBrokerActiveControllerCount** | critical | нет активного контроллера > 2 мин | Потеря контроллера кластера |
| **KafkaBrokerUnderReplicatedPartitions** | warning | недореплицированные партиции > 0 | Проблемы с репликацией |
| **KafkaBrokerRequestHandlerIdle** | warning | idle < 10% > 10 мин | Нагрузка на обработчики запросов |
| **KafkajvmMemoryHigh** | warning | память > 90% > 5 мин | Высокое использование памяти JVM |
| **KafkaBrokerProducerRequestRateLow** | info | низкая скорость запросов | Низкая активность продюсеров |

### Настройка email-оповещений

Для получения email оповещений редактируйте `monitoring/alertmanager.yml`:

```yaml
global:
  smtp_smarthost: "your-smtp-server:587"
  smtp_from: "alertmanager@yourdomain.com"
  smtp_auth_username: "your-smtp-username"
  smtp_auth_password: "your-smtp-password"
  smtp_require_tls: true
```

Замените email адреса в секции `receivers` на свои.

### Проверка мониторинга

```bash
# Проверить статус JMX Exporter
curl http://localhost:7071/metrics

# Проверить статус Prometheus targets
curl 'http://localhost:9090/api/v1/targets' | grep -A 2 "health"

# Проверить метрики в Prometheus
curl 'http://localhost:9090/api/v1/query?query=kafka_controller_kafkacontroller_activebrokercount'

# Проверить алерты в Prometheus
http://localhost:9090/alerts

# Проверить логи сервисов
docker compose logs prometheus
docker compose logs alertmanager
docker compose logs grafana
docker logs kafka-0-jmx
```

### Дашборды Grafana

1. Откройте Grafana: http://localhost:3000
2. Логин: admin, пароль: admin
3. Дашборд "Kafka Monitoring Dashboard" автоматически provisioning при запуске

**Панели дашборда:**
- Статус брокеров (онлайн/офлайн)
- Скорость передачи данных (Bytes In/Out)
- Запросы продюсеров/консьюмеров
- Недореплицированные партиции
- Офлайн партиции
- Использование JVM памяти
- Загрузка request handler'ов
- Статус активного контроллера
- Латентность запросов

### JMX и оба кластера

**Кластер 1:**
- kafka-0-jmx: localhost:7071
- kafka-1-jmx: localhost:7072
- kafka-2-jmx: localhost:7073

**Кластер 2:**
- kafka-3-jmx: localhost:7074
- kafka-4-jmx: localhost:7075
- kafka-5-jmx: localhost:7076

Прямой доступ к метрикам:
```bash
curl http://localhost:7071/metrics  # kafka-0
curl http://localhost:7072/metrics  # kafka-1
# и т.д.
```

### Полная документация

Подробная документация по мониторингу: [monitoring/README.md](monitoring/README.md)

## Остановка сервисов

```bash
docker compose down
```
