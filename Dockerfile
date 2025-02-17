FROM golang:1.23-alpine AS builder
WORKDIR /build
ADD go.mod .
COPY . .
RUN go build -o mqtt_exporter main.go
FROM alpine
LABEL org.opencontainers.image.description="MQTT Exporter"
LABEL org.opencontainers.image.source=https://github.com/sbouchex/mqtt_exporter
LABEL org.opencontainers.image.licenses=Apache-2.0
WORKDIR /mqtt_exporter_data
COPY --from=builder /build/mqtt_exporter /mqtt_exporter
CMD ["/mqtt_exporter"]
EXPOSE 9393
