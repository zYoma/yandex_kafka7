# Скрипты настройки Kafka

Этот каталог содержит скрипты для настройки конфигурации Kafka и ACL.

## Скрипты

### setup-all.sh (Рекомендуется)
Запускает все шаги настройки последовательно:
1. Создаёт конфигурацию клиента
2. Настраивает внутренние ACL Kafka для admin
3. Настраивает ACL для Schema Registry
4. Настраивает ACL для интер-брокер пользователя
5. Создаёт топик `shops_data`
6. Настраивает ACL для клиентов топика (shop_client, analytic_client, filter_client)
7. Показывает все настроенные ACL

## Использование

### После docker compose up
```bash
./scripts/setup-all.sh
```

## Предварительные требования
- Docker контейнеры должны быть запущены
- Kafka брокеры должны быть готовы (подождать ~20 секунд после `docker compose up -d`)
- Скрипты используют контейнер `yandex_kafka7-kafka-0-1` для операций управления

## Текущая конфигурация
- **Пользователи:**
  - `admin` - Суперпользователь для управления и Schema Registry
  - `user` - Интер-брокер коммуникация
  - `shop_client` - Может писать в `shops_data`
  - `analytic_client` - Может читать из `shops_data`
  - `filter_client` - Может читать из `shops_data`

- **Топики:**
  - `shops_data` - 3 партиции, 3 реплики
  - `_schemas` - Внутренний топик Schema Registry

- **Сервисы:**
  - Кластер 1: kafka-0 (9094), kafka-1 (9095), kafka-2 (9096)
  - Кластер 2: kafka-3 (9097), kafka-4 (9098), kafka-5 (9099)
  - Schema Registry: http://localhost:8081 (подключен к Кластеру 1)

## Безопасность
- Протокол: SASL_SSL
- Механизм: PLAIN
- ACL: Включены в строгом режиме (`ALLOW_EVERYONE_IF_NO_ACL_FOUND: false`)