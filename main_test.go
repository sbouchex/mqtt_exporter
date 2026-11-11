package main

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestPushMetricsCountByTopicAndSensor(t *testing.T) {
	metrics := newPushMetrics("mqtt_exporter_")
	registry := prometheus.NewRegistry()
	if err := registry.Register(metrics.topic); err != nil {
		t.Fatal(err)
	}
	if err := registry.Register(metrics.sensor); err != nil {
		t.Fatal(err)
	}

	metrics.topic.WithLabelValues("zigbee2mqtt/living_room").Inc()
	metrics.topic.WithLabelValues("zigbee2mqtt/living_room").Inc()
	metrics.sensor.WithLabelValues("sensors_prises").Inc()

	families, err := registry.Gather()
	if err != nil {
		t.Fatal(err)
	}

	checkCounter := func(name, labelName, labelValue string, want float64) {
		t.Helper()
		for _, family := range families {
			if family.GetName() != name {
				continue
			}
			for _, metric := range family.GetMetric() {
				for _, label := range metric.GetLabel() {
					if label.GetName() == labelName && label.GetValue() == labelValue {
						if got := metric.GetCounter().GetValue(); got != want {
							t.Errorf("%s{%s=%q} = %v, want %v", name, labelName, labelValue, got, want)
						}
						return
					}
				}
			}
		}
		t.Errorf("metric %s{%s=%q} not found", name, labelName, labelValue)
	}

	checkCounter("mqtt_exporter_pushes_per_topic_total", "topic", "zigbee2mqtt/living_room", 2)
	checkCounter("mqtt_exporter_pushes_per_sensor_total", "sensor", "sensors_prises", 1)
}
