FROM golang:1.24.2-alpine AS builder
WORKDIR /greenlight
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api

FROM alpine:latest
WORKDIR /greenlight
COPY --from=builder /greenlight/bin/linux_amd64/ .
EXPOSE 4545
ENTRYPOINT ["./api"]
