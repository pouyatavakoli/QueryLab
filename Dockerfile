# ---------- Build stage ----------
FROM golang:1.22-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o querylab ./cmd/server

# ---------- Runtime stage ----------
FROM alpine:3.19

WORKDIR /app
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/querylab /app/querylab
COPY frontend /app/frontend
COPY init.sql /app/init.sql

EXPOSE 8080

ENTRYPOINT ["/app/querylab"]
