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

## Остановка сервисов

```bash
docker compose down
```
