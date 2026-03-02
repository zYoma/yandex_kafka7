#!/bin/bash

ALLOWED_PRODUCTS_FILE="$(dirname "$0")/../kafka-connect/allowed-products.json"

function show_help() {
    echo "Usage: $0 [command] [args]"
    echo ""
    echo "Commands:"
    echo "  list                          - Show all allowed product IDs"
    echo "  add <product_id>              - Add a product ID to allowed list"
    echo "  remove <product_id>           - Remove a product ID from allowed list"
    echo "  clear                         - Clear all allowed product IDs"
    echo "  help                          - Show this help message"
}

function ensure_file_exists() {
    if [[ ! -f "$ALLOWED_PRODUCTS_FILE" ]]; then
        echo '{"allowed_product_ids": []}' > "$ALLOWED_PRODUCTS_FILE"
    fi
}

function list_products() {
    ensure_file_exists
    echo "Allowed Product IDs:"
    echo "-------------------"
    cat "$ALLOWED_PRODUCTS_FILE" | jq -r '.allowed_product_ids[]' | nl
}

function add_product() {
    local product_id="$1"
    
    if [[ -z "$product_id" ]]; then
        echo "Error: Product ID is required"
        echo "Usage: $0 add <product_id>"
        exit 1
    fi

    ensure_file_exists
    
    if cat "$ALLOWED_PRODUCTS_FILE" | grep -q "\"$product_id\""; then
        echo "Product ID '$product_id' already exists in allowed list"
        exit 0
    fi

    local temp_file=$(mktemp)
    cat "$ALLOWED_PRODUCTS_FILE" | jq ".allowed_product_ids += [\"$product_id\"]" > "$temp_file"
    mv "$temp_file" "$ALLOWED_PRODUCTS_FILE"
    
    echo "Added product ID: $product_id"
    echo ""
    echo "Tip: Restart Kafka Connect to apply changes:"
    echo "  docker compose restart kafka-connect"
}

function remove_product() {
    local product_id="$1"
    
    if [[ -z "$product_id" ]]; then
        echo "Error: Product ID is required"
        echo "Usage: $0 remove <product_id>"
        exit 1
    fi

    ensure_file_exists
    
    if ! cat "$ALLOWED_PRODUCTS_FILE" | grep -q "\"$product_id\""; then
        echo "Product ID '$product_id' not found in allowed list"
        exit 1
    fi

    local temp_file=$(mktemp)
    cat "$ALLOWED_PRODUCTS_FILE" | jq ".allowed_product_ids -= [\"$product_id\"]" > "$temp_file"
    mv "$temp_file" "$ALLOWED_PRODUCTS_FILE"
    
    echo "Removed product ID: $product_id"
    echo ""
    echo "Tip: Restart Kafka Connect to apply changes:"
    echo "  docker compose restart kafka-connect"
}

function clear_products() {
    ensure_file_exists
    
    cat "$ALLOWED_PRODUCTS_FILE" | jq ".allowed_product_ids = []" > "$ALLOWED_PRODUCTS_FILE"
    
    echo "Cleared all allowed product IDs"
    echo ""
    echo "Tip: Restart Kafka Connect to apply changes:"
    echo "  docker compose restart kafka-connect"
}

command="$1"
shift || true

case "$command" in
    list)
        list_products
        ;;
    add)
        add_product "$@"
        ;;
    remove)
        remove_product "$@"
        ;;
    clear)
        clear_products
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
