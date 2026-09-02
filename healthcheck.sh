#!/bin/sh
# Script to perform health check by reading the configured port from mqtt_exporter.json

CONFIG_FILE="/mqtt_exporter_data/mqtt_exporter.json"
DEFAULT_PORT="9393"

# Try to extract the port from the configuration file
if [ -f "$CONFIG_FILE" ]; then
    PORT=$(grep -o '"listeningAddress"[^,}]*' "$CONFIG_FILE" | grep -o '[0-9]\+' | tail -1)
    if [ -z "$PORT" ]; then
        PORT=$DEFAULT_PORT
    fi
else
    PORT=$DEFAULT_PORT
fi

# Perform the health check
wget --no-verbose --tries=1 --spider "http://localhost:${PORT}/healthz" || exit 1
