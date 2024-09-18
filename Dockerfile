FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o app ./cmd/server

FROM alpine:3.20

RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/app .
COPY env.yaml .

RUN addgroup -S appgroup && adduser -S appuser -G appgroup -h /home/appuser
RUN chown -R appuser:appgroup /app
USER appuser


CMD ["./app"]