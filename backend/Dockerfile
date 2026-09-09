FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/server /app/server
ENV PORT=8080
ENV DB_PATH=/data/media_scheduler.db
RUN mkdir -p /data
EXPOSE 8080
CMD ["/app/server"]
