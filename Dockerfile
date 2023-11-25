FROM golang:latest AS builder

WORKDIR /go/src/discord-stock-ticker

COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o discord-stock-ticker

# Stage 2
FROM alpine:latest  
LABEL org.opencontainers.image.source https://github.com/bocchi-the-crypto/discord-stock-ticker

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /go/src/discord-stock-ticker/discord-stock-ticker .

# expose API server port
EXPOSE 7439

CMD ["./discord-stock-ticker", "-db=/app/discordbottickers.db", "-address=0.0.0.0:7439"]
