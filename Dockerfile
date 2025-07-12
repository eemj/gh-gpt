# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /gh-gpt ./cmd/gh-gpt

# Final stage
FROM alpine:latest

COPY --from=builder /gh-gpt /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/gh-gpt"]
