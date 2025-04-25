# Builder image
FROM library/golang@sha256:d9db32125db0c3a680cfb7a1afcaefb89c898a075ec148fdc2f0f646cc2ed509 AS builder
WORKDIR /build
ADD go.mod .
COPY . .
RUN go build -o mqtt_exporter main.go

# Runner image
FROM library/alpine@sha256:a8560b36e8b8210634f77d9f7f9efd7ffa463e380b75e2e74aff4511df3ef88c
LABEL org.opencontainers.image.description="MQTT Exporter"
LABEL org.opencontainers.image.source=https://github.com/sbouchex/mqtt_exporter
LABEL org.opencontainers.image.licenses=Apache-2.0
WORKDIR /mqtt_exporter_data
COPY --from=builder /build/mqtt_exporter /mqtt_exporter
CMD ["/mqtt_exporter"]
EXPOSE 9393
