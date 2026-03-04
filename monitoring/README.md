# Kafka Monitoring Stack

Этот каталог содержит конфигурацию для мониторинга Kafka кластеров с использованием:
- Prometheus - сбор метрик
- Grafana - визуализация
- Alertmanager - оповещения
- JMX Exporter - экспорт JMX метрик от Kafka

## Структура

```
monitoring/
├── prometheus.yml                 - конфигурация Prometheus
├── alerts.yml                     - правила алертов
├── alertmanager.yml               - конфигурация Alertmanager
├── grafana-dashboard.json         - дашборд Grafana
├── grafana-datasources.yml        - datasource для Grafana
├── jmx_exporter_config.yml        - конфигурация JMX Exporter
├── README.md                      - эта документация
├── start-monitoring.sh            - запуск всего стэка
└── stop-monitoring.sh             - остановка всего стэка
```

## Запуск

Все сервисы (включая мониторинг) запускаются через корневой docker-compose.yml:

```bash
docker-compose up -d
```

Или используя скрипт:

```bash
./monitoring/start-monitoring.sh
```

## Доступ к сервисам

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Alertmanager**: http://localhost:9093

## JMX Exporter

JMX Exporter подключен к каждому брокеру как sidecar контейнер и доступен на портах:
- kafka-0-jmx: 7071
- kafka-1-jmx: 7072
- kafka-2-jmx: 7073
- kafka-3-jmx: 7074
- kafka-4-jmx: 7075
- kafka-5-jmx: 7076

Проверка работы JMX Exporter:
```bash
curl http://localhost:7071/metrics
```

## Мониторинг

### Prometheus

Адрес: http://localhost:9090

Примеры запросов:
```promql
# Статус брокеров
up{job=~"kafka-cluster-.*"}

# Подключенные брокеры
kafka_controller_KafkaController_ActiveControllerCount

# Недореплицированные партиции
kafka_server_ReplicaManager_UnderReplicatedPartitions

# Офлайн партиции
kafka_server_ReplicaManager_OfflinePartitionsCount

# Загрузка сети
rate(kafka_server_BrokerTopicMetrics_BytesInPerSec_total[5m])
rate(kafka_server_BrokerTopicMetrics_BytesOutPerSec_total[5m])

# Использование памяти JVM
(jvm_memory_heap_bytes{area="heap"} / jvm_memory_heap_bytes_max{area="heap"}) * 100
```

### Grafana

Адрес: http://localhost:3000

Дашборд "Kafka Monitoring Dashboard" автоматически provisioning при запуске.

### Настройка Email оповещений

Для получения email оповещений нужно изменить `monitoring/alertmanager.yml`:

```yaml
global:
  smtp_smarthost: "your-smtp-server:587"
  smtp_from: "alertmanager@yourdomain.com"
  smtp_auth_username: "your-smtp-username"
  smtp_auth_password: "your-smtp-password"
  smtp_require_tls: true
```

И заменить email адреса в секции `receivers`.

## Алерты

Созданы следующие алерты:

| Название | Уровень | Описание |
|----------|---------|----------|
| KafkaBrokerDown | critical | Брокер недоступен более 1 мин |
| KafkaBrokerUnderReplicatedPartitions | warning | Недореплицированные партиции |
| KafkaBrokerOfflinePartitionsCount | critical | Партиции в офлайн состоянии |
| KafkaBrokerActiveControllerCount | critical | Нет активного контроллера |
| KafkaBrokerRequestHandlerIdle | warning | Нагрузка на обработчики запросов |
| KafkajvmMemoryHigh | warning | Высокое использование памяти JVM |
| KafkaBrokerProducerRequestRateLow | info | Низкая скорость запросов |

## Проверка

Проверить что Prometheus собирает метрики:
```
http://localhost:9090/targets
```

## Остановка

```bash
docker-compose down
```

Или используя скрипт:
```bash
./monitoring/stop-monitoring.sh
```
