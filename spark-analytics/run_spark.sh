#!/bin/bash

HDFS_PATH=${1:-"hdfs://namenode:9000/kafka_data"}

echo "Starting Spark analytics job..."
echo "HDFS path: $HDFS_PATH"

# Путь к PySpark в официальном образе Apache Spark
docker-compose exec spark-master spark-submit \
  --master spark://spark-master:7077 \
  --name "Popular Products Analytics" \
  --executor-memory "512m" \
  /opt/spark/jobs/popular_products.py $HDFS_PATH

echo "Spark job completed!"