ARG GO_VERSION=1.26.3

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/onepractice-api ./cmd/server

FROM alpine:3.22 AS runner

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && addgroup -S app && adduser -S app -G app

COPY --from=builder /out/onepractice-api /app/onepractice-api
COPY openapi /app/openapi

USER app

EXPOSE 8080

CMD ["/app/onepractice-api"]
