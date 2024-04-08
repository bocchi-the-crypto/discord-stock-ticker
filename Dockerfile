FROM golang:latest AS builder

WORKDIR /go/src/discord-stock-ticker

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o discord-stock-ticker

FROM alpine:latest

WORKDIR /app/

COPY --from=builder /go/src/discord-stock-ticker/discord-stock-ticker .

ENTRYPOINT ["./discord-stock-ticker"]