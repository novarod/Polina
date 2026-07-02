FROM golang:1.26-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20 AS runner

RUN apk add --no-cache ca-certificates wget \
  && addgroup -S app && adduser -S -G app app

WORKDIR /app
COPY --from=builder /out/server /app/server

USER app

ENV PORT=8080
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- "http://localhost:${PORT}/health" >/dev/null 2>&1 || exit 1

ENTRYPOINT ["/app/server"]
