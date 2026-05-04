FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-s -w' -o /bin/url-shortener ./cmd/api

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /bin/url-shortener /app/url-shortener

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/url-shortener"]
