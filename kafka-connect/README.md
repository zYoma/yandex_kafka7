# Kafka Connect Plugin - Product Filter Transformer

## Структура

```
kafka-connect/
├── allowed-products.json         # Список разрешенных product_id
├── sink-connector.json           # Конфигурация коннектора
├── build-plugin.sh               # Скрипт сборки плагина
├── output/                       # Директория для вывода отфильтрованных данных
│   └── filtered-products.txt
├── plugins/                      # Плагины для Kafka Connect
│   ├── lib/                      # Скомпилированные JAR (может быть в репозитории)
│   │   └── product-filter-transformer-1.0.0.jar
│   ├── pom.xml                   # Maven конфигурация
│   └── src/main/java/com/example/kafka/connect/transform/
│       └── ProductFilter.java    # Кастомный SMT трансформер
└── README.md
```

## Компиляция плагина

### Проверка готового JAR

Сначала проверьте, существует ли уже скомпилированный JAR в репозитории:

```bash
ls kafka-connect/plugins/lib/product-filter-transformer-1.0.0.jar
```

Если файл существует, плагин уже готов к использованию без компиляции.

### Сборка плагина (при необходимости)

```bash
cd kafka-connect
./build-plugin.sh
```

Скрипт поддерживает:
- Использование локального Maven (если установлен)
- Использование Docker с Maven (если локальный Maven недоступен)
- Пропуск сборки, если JAR уже существует
- Принудительную пересборку с флагом `--force`

После сборки JAR файл будет находиться в `plugins/lib/product-filter-transformer-1.0.0.jar`

### Хранение JAR в репозитории

Для удобства использования скомпилированный JAR может храниться в репозитории. Это рекомендуется для учебных проектов:

**Преимущества:**
- Не требует установки Maven для обычных пользователей
- Быстрый старт проекта
- Гарантированная совместимость версий

**Недостатки:**
- Увеличение размера репозитория
- Необходимо пересобирать при изменении исходного кода

## Конфигурация трансформера

Трансформер `ProductFilter` фильтрует записи по полю `product_id` и пропускает только те идентификаторы, которые указаны в файле `allowed-products.json`.

### Параметры

- `allowed.products.file` - путь к JSON файлу с разрешенными product_id

### Формат JSON

```json
{
  "allowed_product_ids": ["12345", "54321"]
}
```

## Коннектор

Коннектор использует `FileStreamSinkConnector` для записи отфильтрованных данных в файл.

Конфигурация:
- Топик: `shops_data`
- Выходной файл: `/data/output/filtered-products.txt`
- Трансформер: `ProductFilter`
