# ---- Build stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o doc-search cmd/server/main.go


# ---- Final stage ----
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /app

COPY --from=builder /app/doc-search .
COPY --from=builder /app/migrations ./migrations

# Create empty folder uploads (will be overwritten by volume)
RUN mkdir -p uploads

EXPOSE 8080

CMD ["./doc-search"]
