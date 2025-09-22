# Multi-stage Dockerfile for multiple services
# Use build-arg SERVICE_DIR to point at the service folder (e.g. polling-service or results-service)

ARG SERVICE_DIR=polling-service

FROM golang:1.24-alpine AS builder
ARG SERVICE_DIR
COPY common /app/common/
COPY ${SERVICE_DIR} /app/service/
WORKDIR /app/service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/service/service .

FROM alpine:latest AS production
WORKDIR /app
COPY --from=builder /app/service/service .
EXPOSE 8080
CMD ["./service"]

FROM golang:1.24-alpine AS development
ARG SERVICE_DIR
WORKDIR /app

# Install delve for live debugging in development images
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# In dev we expect the source to be mounted into /app/<SERVICE_DIR> by compose.override.yml.
# Use dlv in debug mode so it will build the binary inside the container from the mounted sources
# which lets breakpoints align with the local files.
WORKDIR /app/${SERVICE_DIR}
EXPOSE 8080
EXPOSE 40000

# Default command runs delve in headless mode, listening on 0.0.0.0:40000 and waiting for a debugger client.
CMD ["dlv", "debug", "--headless", "--listen=0.0.0.0:40000", "--api-version=2", "--accept-multiclient", "--continue"]
