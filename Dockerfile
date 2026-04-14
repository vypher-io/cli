# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.version=${VERSION}" -o vypher .

# Final stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/vypher /usr/local/bin/vypher

ENTRYPOINT ["vypher"]
CMD ["--help"]
