#!/bin/bash

set -e

SCHEMA_REGISTRY_URL="${SCHEMA_REGISTRY_URL:-http://localhost:8081}"
TOPIC_NAME="${TOPIC_NAME:-shops_data_json}"
SCHEMA_FILE="$(dirname "$0")/../kafka-connect/product-schema-avro.json"

function show_help() {
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  register                      - Register schema in Schema Registry"
    echo "  delete                        - Delete schema from Schema Registry"
    echo "  list                          - List all schemas in Schema Registry"
    echo "  get                           - Get current schema version for topic"
    echo "  help                          - Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  SCHEMA_REGISTRY_URL           - Schema Registry URL (default: http://localhost:8081)"
    echo "  TOPIC_NAME                    - Topic name (default: shops_data_json)"
}

function check_schema_registry() {
    if ! curl -s --connect-timeout 5 "${SCHEMA_REGISTRY_URL}" > /dev/null; then
        echo "Error: Cannot connect to Schema Registry at ${SCHEMA_REGISTRY_URL}"
        echo "Please ensure Schema Registry is running:"
        echo "  docker compose ps schema-registry"
        exit 1
    fi
}

function register_schema() {
    echo "Registering Avro schema for topic: ${TOPIC_NAME}"
    echo ""
    
    if [[ ! -f "$SCHEMA_FILE" ]]; then
        echo "Error: Schema file not found: $SCHEMA_FILE"
        exit 1
    fi
    
    # Read schema and escape for JSON
    SCHEMA_CONTENT=$(cat "$SCHEMA_FILE" | jq -c . | sed 's/"/\\"/g')
    
    # Register Avro schema
    RESPONSE=$(curl -s -X POST \
        -H "Content-Type: application/vnd.schemaregistry.v1+json" \
        --data "{\"schema\": \"${SCHEMA_CONTENT}\"}" \
        "${SCHEMA_REGISTRY_URL}/subjects/${TOPIC_NAME}-value/versions")
    
    SCHEMA_ID=$(echo "$RESPONSE" | jq -r '.id')
    
    if [[ "$SCHEMA_ID" != "null" && -n "$SCHEMA_ID" ]]; then
        echo "Schema registered successfully with ID: ${SCHEMA_ID}"
        echo ""
        echo "Schema URL: ${SCHEMA_REGISTRY_URL}/subjects/${TOPIC_NAME}-value/versions/latest"
    else
        echo "Error registering schema:"
        echo "$RESPONSE"
        exit 1
    fi
}

function delete_schema() {
    echo "Deleting schema for topic: ${TOPIC_NAME}"
    echo ""
    
    curl -s -X DELETE "${SCHEMA_REGISTRY_URL}/subjects/${TOPIC_NAME}-value"
    
    echo ""
    echo "Schema deleted successfully"
}

function list_schemas() {
    echo "All schemas in Schema Registry:"
    echo "-------------------------------"
    curl -s "${SCHEMA_REGISTRY_URL}/subjects" | jq -r '.[]'
}

function get_schema() {
    echo "Current schema for topic: ${TOPIC_NAME}"
    echo "--------------------------------------"
    
    curl -s "${SCHEMA_REGISTRY_URL}/subjects/${TOPIC_NAME}-value/versions/latest" | jq '.'
}

command="$1"

case "$command" in
    register)
        check_schema_registry
        register_schema
        ;;
    delete)
        check_schema_registry
        delete_schema
        ;;
    list)
        check_schema_registry
        list_schemas
        ;;
    get)
        check_schema_registry
        get_schema
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
