# MQTT Exporter

# Presentation
An exporter for MQTT. It received metrics in JSON format via MQTT, transforms them and exposes them for consumption by [Prometheus](https://www.prometheus.io/).

# Purpose
Collectd expose metrics to MQTT.  However, the metric names and labels are fixed, cannot be filtered and if an entity is renamed, the metric is renamed in prometheus and data are lost without changing the dashboard.

# Configuration
This exporter uses a configuration file containing mapping entries to defines which entity is exported, if it's a gauge or a counter.
The metric prefix can be configured as well.

## Example
```
{
    "config": {
        "ListeningAddress": ":9393",
        "metricsPath": "/metrics",
        "configurationFile": "configuration.json"
    },
    "mqtt": {
        "broker": "tcp://<HOST>:1883"
        "topic: "/collectd/#"
    }
}
```

# Usage
* Build the container from the source:
```
docker build -t mqtt_exporter
```
* Start the container:
```
docker run -d -p 9103:9103 --name=mqtt_exporter --network bouchex --restart=always -v mqtt_exporter:/mqtt_exporter_data mqtt_exporter:latest /mqtt_exporter
```
