# Yandex Kafka Проект

Учебный проект с двумя кластерами Kafka и Schema Registry.

## Быстрый старт

```bash
# Запуск всех сервисов
docker compose up -d

# Подождать инициализации сервисов (~25 секунд)

# Настройка топиков, пользователей и ACL
./scripts/setup-all.sh
```

## Сервисы

### Кластеры Kafka
- **Кластер 1**: kafka-0 (9094), kafka-1 (9095), kafka-2 (9096)
- **Кластер 2**: kafka-3 (9097), kafka-4 (9098), kafka-5 (9099)

Оба кластера настроены с:
- SASL_SSL аутентификацией
- PLAIN механизмом
- ACL авторизацией включена

### Schema Registry
- URL: http://localhost:8081
- Подключен к: Кластер 1
- Хранит схемы в топике `_schemas`

## Пользователи и права доступа

| Пользователь    | Пароль        | Роль                  | Доступ                                   |
|-----------------|---------------|-----------------------|------------------------------------------|
| admin           | admin-secret  | Суперпользователь     | Все операции, управление                 |
| user            | user-secret   | Интер-брокер          | Необходим для внутренней репликации Kafka |
| shop_client     | shop-secret   | Продюсер              | WRITE в `shops_data`                      |
| analytic_client | analytic-secret| Консьюмер           | READ из `shops_data`                      |
| filter_client   | filter-secret  | Консьюмер           | READ из `shops_data`                      |

## Топики

| Топик      | Партиции | Репликация | Продюсер      | Консьюмеры                         |
|------------|-----------|------------|---------------|-------------------------------------|
| shops_data | 3         | 3          | shop_client   | analytic_client, filter_client      |

## Скрипты

### scripts/setup-all.sh
Единый скрипт настройки, который:
1. Создаёт конфигурацию клиента
2. Настраивает внутренние ACL Kafka
3. Настраивает ACL для Schema Registry
4. Настраивает интер-брокер коммуникацию
5. Создаёт топики с правильными ACL
6. Показывает все настройки

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

## Остановка сервисов

```bash
docker compose down
```
