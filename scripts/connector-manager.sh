#!/bin/bash

set -e

CONNECTOR_FILE="$(dirname "$0")/../kafka-connect/sink-connector.json"

function show_help() {
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  deploy                        - Deploy the connector to Kafka Connect"
    echo "  delete                        - Delete the connector from Kafka Connect"
    echo "  status                        - Show connector status"
    echo "  help                          - Show this help message"
}

function deploy_connector() {
    echo "Deploying connector to Kafka Connect..."
    
    if [[ ! -f "$CONNECTOR_FILE" ]]; then
        echo "Error: Connector configuration file not found: $CONNECTOR_FILE"
        exit 1
    fi

    curl -X POST -H "Content-Type: application/json" \
         --data @"$CONNECTOR_FILE" \
         http://localhost:8083/connectors
    
    echo ""
    echo "Connector deployed successfully!"
    echo ""
    echo "Check status with: $0 status"
}

function delete_connector() {
    echo "Deleting connector from Kafka Connect..."
    
    curl -X DELETE http://localhost:8083/connectors/product-filter-sink
    
    echo ""
    echo "Connector deleted successfully!"
}

function show_status() {
    echo "Connector Status:"
    echo "-----------------"
    
    curl -s http://localhost:8083/connectors/product-filter-sink/status | jq '.'
    
    echo ""
    echo "Connector Config:"
    echo "-----------------"
    
    curl -s http://localhost:8083/connectors/product-filter-sink/config | jq '.'
}

command="$1"

case "$command" in
    deploy)
        deploy_connector
        ;;
    delete)
        delete_connector
        ;;
    status)
        show_status
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo "Error: Unknown command '$command'"
        echo ""
        show_help
        exit 1
        ;;
esac
