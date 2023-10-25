# MQTT Exporter

# Presentation
An exporter for MQTT. It received metrics in various format (JSON, collectd, raw...) via MQTT, filters them using regular expression, transforms them and exposes them for consumption by [Prometheus](https://www.prometheus.io/).

# Purpose
[zigbee2mqtt](https://www.zigbee2mqtt.io/) or [collectd](https://collectd.org/) may expose metrics to MQTT.  However, the metric names and labels are fixed, cannot be filtered and if an entity is renamed by the source, the metric is renamed in prometheus and dashboard must be changed to follow thoses changes.

# Configuration
This exporter uses 2 configuration files:
- mqtt_exporter.json: To configure the listening port, the MQTT parameters (broken address, client Id) and the path to the configuration file (see below)
- configuration.json: To configure mapping entries to defines which entity is exported, the payload type, additional labels

## mqtt_exporter.json example
```
{
    "config": {
        "ListeningAddress": ":9393",
        "metricsPath": "/metrics",
        "configurationFile": "configuration.json"
    },
    "mqtt": {
        "broker": "tcp://<HOST>:1883"
        "clientId": "mqtt_exporter_client"
    }
}
```

## configuration.json example
```
{
    "prefix": "mqtt_exporter_",
    "purgeDelay": 3600,
    "topics": [
        "zigbee2mqtt/#"
    ],
    "sensors": {
        "sensors": {
            "payloadType": "json",
            "filter": "zigbee2mqtt/(?P<L1>prise_.+)",
            "labels": [],
            "values": {
                "linkquality": "$.linkquality",
                "voltage": "$.voltage",
                "state": "$.state",
                "battery": "$.battery"
            }
        }
    }
}
```

### Parameters:
- prefix: All prometheus are prefixed by this string
- purgeDelay: Metrics are deleted from the prometheus registry if no update occured after this delay
- topics: MQTT topics to listen
- sensors: Collection of sensor definitions with various parameters
    - payloadType: Payload type (json, collectd or raw)
    - filter: Filter the topic to keep and extract labels
    - labels: Prometheus labels to add
    - values (*json payloadType only*): json path of the value to extract

# Usage
* Build the container from the source:
```
docker build -t mqtt_exporter
```
* Start the container:
```
docker run -d -p 9103:9103 --name=mqtt_exporter --network bouchex --restart=always -v mqtt_exporter:/mqtt_exporter_data mqtt_exporter:latest /mqtt_exporter
```

# Dev
The source code are written in [Go](https://go.dev/) and uses various packages (to handle MQTT, prometheus, logging)