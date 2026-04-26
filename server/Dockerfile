# ================================
# Build
# ================================
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o indexer ./cmd/indexer

# ================================
# Final image
# ================================
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata curl poppler-utils

WORKDIR /app

COPY --from=builder /build/api .
COPY --from=builder /build/indexer .

RUN adduser -D -u 1000 appuser && chown -R appuser /app
USER appuser

CMD ["./api"]

