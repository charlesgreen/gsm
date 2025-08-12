FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gsm-server ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1001 -S gsm && \
    adduser -u 1001 -S gsm -G gsm

WORKDIR /app

COPY --from=builder /app/gsm-server .

RUN chown -R gsm:gsm /app

USER gsm

EXPOSE 8085

VOLUME ["/app/data"]

ENV GSM_PORT=8085
ENV GSM_HOST=0.0.0.0
ENV GSM_LOG_LEVEL=info
ENV GSM_ENABLE_CORS=true

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8085/health || exit 1

CMD ["./gsm-server"]