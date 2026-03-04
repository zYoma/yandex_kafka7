#!/bin/bash

CLIENT_CONTAINER="yandex_kafka7-client-1"

case "$1" in
  search)
    if [ -z "$2" ]; then
      echo "Usage: $0 search <product_name>"
      exit 1
    fi
    docker compose run --rm client /app/yandex-kafka-client search "$2"
    ;;

  recommendations)
    docker compose run --rm client /app/yandex-kafka-client recommendations
    ;;

  send-recommendations)
    docker compose run --rm analytic_client /app/yandex-kafka-analytic
    ;;

  bash)
    docker compose run --rm client /bin/bash
    ;;

  *)
    echo "Kafka CLI Helper (Docker)"
    echo ""
    echo "Usage:"
    echo "  $0 search <product_name>        - Search for product by name"
    echo "  $0 recommendations              - Get personal recommendations"
    echo "  $0 send-recommendations        - Send sample recommendations to Kafka"
    echo "  $0 bash                        - Start bash shell in client container"
    echo ""
    echo "Direct docker compose commands:"
    echo "  docker compose run client /app/yandex-kafka-client search 'Product'"
    echo "  docker compose run client /app/yandex-kafka-client recommendations"
    echo "  docker compose run analytic_client /app/yandex-kafka-analytic"
    exit 1
    ;;
esac
