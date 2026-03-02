# Скрипты настройки Kafka

Этот каталог содержит скрипты для настройки конфигурации Kafka, ACL и Kafka Connect.

## Скрипты

### setup-all.sh (Рекомендуется)
Запускает все шаги настройки для первого кластера последовательно:
1. Создаёт конфигурацию клиента
2. Настраивает внутренние ACL Kafka для admin
3. Настраивает ACL для Schema Registry
4. Настраивает ACL для интер-брокер пользователя
5. Создаёт топик `shops_data`
6. Настраивает ACL для клиентов топика (shop_client, analytic_client, filter_client)
7. Показывает все настроенные ACL

### setup-target-cluster-acls.sh
Настраивает ACL для второго (целевого) кластера:
1. Настраивает ACL для admin на целевом кластере
2. Настраивает интер-брокер коммуникацию для целевого кластера
3. Настраивает права для MirrorMaker 2

### setup-kafka-connect.sh
Настраивает Kafka Connect:
1. Создаёт топики для Connect (connect-configs, connect-offsets, connect-status)
2. Создаёт топик shops_data_json для данных в формате Avro
3. Настраивает ACL для filter_client (READ и DESCRIBE на shops_data_json)
4. Настраивает ACL для consumer groups Connect (connect-*)

### connector-manager.sh
Управляет коннекторами Kafka Connect:
- `deploy` - развёртывает коннектор на Connect
- `delete` - удаляет коннектор
- `status` - показывает статус и конфигурацию коннектора

```bash
./scripts/connector-manager.sh deploy
./scripts/connector-manager.sh status
./scripts/connector-manager.sh delete
```

### schema-manager.sh
Управляет схемами в Schema Registry:
- `register` - регистрирует схему для топика
- `delete` - удаляет схему
- `list` - показывает все зарегистрированные схемы
- `get` - получает текущую схему для топика

```bash
./scripts/schema-manager.sh register
./scripts/schema-manager.sh list
./scripts/schema-manager.sh get
```

**Переменные окружения:**
- `SCHEMA_REGISTRY_URL` - URL Schema Registry (по умолчанию: http://localhost:8081)
- `TOPIC_NAME` - имя топика (по умолчанию: shops_data_json)

### manage-products.sh
Управляет списком разрешённых product_id для фильтрации:
- `list` - показывает все разрешённые product_id
- `add <product_id>` - добавляет product_id в разрешённый список
- `remove <product_id>` - удаляет product_id из списка
- `clear` - очищает весь список

```bash
./scripts/manage-products.sh list
./scripts/manage-products.sh add "99999"
./scripts/manage-products.sh remove "99999"
./scripts/manage-products.sh clear
```

> После изменения списка нужно перезапустить Kafka Connect: `docker compose restart kafka-connect`


## Использование

### Первичная настройка всех сервисов

```bash
# 1. Запуск контейнеров
docker compose up -d

# 2. Настройка первого кластера
./scripts/setup-all.sh

# 3. Настройка второго кластера
./scripts/setup-target-cluster-acls.sh

# 4. Настройка Kafka Connect
./scripts/setup-kafka-connect.sh

# 5. Регистрация схемы в Schema Registry
./scripts/schema-manager.sh register

# 6. Перезапуск Kafka Connect для загрузки конфигурации
docker compose restart kafka-connect

# 7. Развертывание коннектора
./scripts/connector-manager.sh deploy
```

### Управление списком разрешённых товаров
```bash
# Показать список
./scripts/manage-products.sh list

# Добавить product_id
./scripts/manage-products.sh add "99999"

# Удалить product_id
./scripts/manage-products.sh remove "99999"
```

### Управление коннектором
```bash
# Развернуть
./scripts/connector-manager.sh deploy

# Проверить статус
./scripts/connector-manager.sh status

# Удалить
./scripts/connector-manager.sh delete
```

### Управление схемами Schema Registry
```bash
# Зарегистрировать схему
./scripts/schema-manager.sh register

# Список всех схем
./scripts/schema-manager.sh list

# Получить схему для топика
./scripts/schema-manager.sh get

# Удалить схему
./scripts/schema-manager.sh delete
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