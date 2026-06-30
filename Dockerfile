FROM golang:1.22-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o wireguard-manager .

FROM alpine:3.19
RUN apk add --no-cache wireguard-tools ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/wireguard-manager .
ENV CONFIG_FILE=/app/config.yml
EXPOSE 8080
ENTRYPOINT ["./wireguard-manager"]
