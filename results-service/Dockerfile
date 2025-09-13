FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/results-service .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/results-service .
EXPOSE 8081
CMD ["./results-service"]
