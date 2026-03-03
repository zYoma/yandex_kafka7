# Spark Analytics

## Описание

PySpark job для анализа поисковых запросов и генерации рекомендаций популярных продуктов.

## Что делает PySpark

1. Читает все JSON файлы запросов из HDFS
2. Использует Spark DataFrame API для анализа
3. Считает частоту поисковых запросов
4. Сохраняет топ-10 популярных продуктов в HDFS

## Запуск

```bash
docker-compose exec spark-master /opt/spark/bin/spark-submit \
  --master spark://spark-master:7077 \
  /opt/spark/jobs/popular_products.py hdfs://namenode:9000/kafka_data
```

## Мониторинг

- **Spark Master UI**: http://localhost:8080
- **Spark Worker UI**: http://localhost:18081
- **HDFS UI**: http://localhost:9870

## Результат

Файл рекомендаций (`part-00000`) содержит:
```json
[
  {"product_name": "товар1", "search_count": 15},
  {"product_name": "товар2", "search_count": 12}
]
```

## Использование в Go

Service `RecommendationService` в `analytic_client` загружает этот файл и использует популярные продукты для рекомендаций.