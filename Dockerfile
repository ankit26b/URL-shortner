# syntax=docker/dockerfile:1.7

FROM golang:1.23-alpine AS base
WORKDIR /src
RUN apk add --no-cache ca-certificates tzdata git
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

FROM base AS test
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go test ./...

FROM base AS build
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/url-shortener ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
WORKDIR /app
COPY --from=build /out/url-shortener /app/url-shortener
EXPOSE 8080
ENTRYPOINT ["/app/url-shortener"]

FROM base AS dev
COPY . .
EXPOSE 8080
CMD ["go", "run", "./cmd/api"]
