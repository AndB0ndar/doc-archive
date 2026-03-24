# ---- Build stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o doc-search cmd/server/main.go


# ---- Final stage ----
FROM alpine:latest

#We install the necessary packages:
#   ca-certificates – for HTTPS
#   curl requests – for healthcheck
#  poppler-utils – for pdftotext (extracting text from PDF
RUN apk --no-cache add ca-certificates curl poppler-utils

WORKDIR /app

COPY --from=builder /app/doc-search .

EXPOSE 8080

CMD ["./doc-search"]
