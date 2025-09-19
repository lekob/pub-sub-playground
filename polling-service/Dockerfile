FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
WORKDIR /app/polling-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/polling-service .

FROM alpine:latest AS production
WORKDIR /app
COPY --from=builder /app/polling-service .
EXPOSE 8080
CMD ["./polling-service"]

FROM golang:1.24-alpine AS development
WORKDIR /app

# Install delve
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# In dev we expect the source to be mounted into /app/polling-service by compose.override.yml.
# Use dlv in debug mode so it will build the binary inside the container from the mounted sources
# which lets breakpoints align with the local files.
WORKDIR /app/polling-service
EXPOSE 8080
EXPOSE 40000

# Default command runs delve in headless mode, listening on 0.0.0.0:40000 and waiting for a debugger client.
CMD ["dlv", "debug", "--headless", "--listen=0.0.0.0:40000", "--api-version=2", "--accept-multiclient"]
