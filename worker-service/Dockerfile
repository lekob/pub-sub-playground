FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/worker-service .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/worker-service .
CMD ["./worker-service"]
