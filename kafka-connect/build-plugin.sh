#!/bin/bash

set -e

SCRIPT_DIR="$(dirname "$0")"
PLUGIN_DIR="$(dirname "$0")/../plugins"

cd "$PLUGIN_DIR"

if [ -f "lib/product-filter-transformer-1.0.0.jar" ]; then
    echo "JAR file already exists at: lib/product-filter-transformer-1.0.0.jar"
    echo "Skipping build. Use --force to rebuild."
    echo ""
    echo "To rebuild, run:"
    echo "  $0 --force"
    exit 0
fi

if [ "$1" = "--force" ]; then
    echo "Building ProductFilterTransformer..."
    echo ""
    
    if command -v mvn &> /dev/null; then
        echo "Using local Maven..."
        mvn clean package -q
    else
        echo "Maven not found locally, using Docker..."
        docker run --rm \
          -v "$(pwd):/workspace" \
          -w /workspace \
          maven:3.9-eclipse-temurin-11 \
          mvn clean package -q
    fi
    
    echo "Build complete!"
else
    echo "JAR file not found. Building plugin..."
    echo ""
    
    if command -v mvn &> /dev/null; then
        echo "Using local Maven..."
        mvn clean package -q
    else
        echo "Maven not found locally, using Docker..."
        docker run --rm \
          -v "$(pwd):/workspace" \
          -w /workspace \
          maven:3.9-eclipse-temurin-11 \
          mvn clean package -q
    fi
    
    echo "Build complete!"
fi

echo ""

JAR_FILE="$PLUGIN_DIR/target/product-filter-transformer-1.0.0.jar"
OUTPUT_DIR="$PLUGIN_DIR/lib"

mkdir -p "$OUTPUT_DIR"
cp "$JAR_FILE" "$OUTPUT_DIR/"

echo "JAR copied to: $OUTPUT_DIR/"
echo ""
echo "Restart Kafka Connect to apply the plugin:"
echo "  docker compose restart kafka-connect"
