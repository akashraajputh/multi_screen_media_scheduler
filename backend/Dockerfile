FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN apk add --no-cache gcc musl-dev
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags='-s -w' -o /app/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/server /app/server
ENV PORT=8080
ENV DB_PATH=/tmp/media_scheduler.db
EXPOSE 8080
CMD ["/app/server"]
