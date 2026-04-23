# Resep yang sama persis kita gunakan untuk aplikasi SSO
FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o sso-api main.go

# Stage 2: Image minimalis
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/sso-api .

EXPOSE 8081

CMD ["./sso-api"]
