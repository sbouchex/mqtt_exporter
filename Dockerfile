FROM golang:alpine AS builder
WORKDIR /build
ADD go.mod .
COPY . .
RUN go build -o mqtt_exporter main.go
FROM alpine
WORKDIR /mqtt_exporter_data
COPY --from=builder /build/mqtt_exporter /mqtt_exporter
CMD ["/mqtt_exporter"]
EXPOSE 9393
