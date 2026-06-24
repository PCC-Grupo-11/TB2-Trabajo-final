FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH

# Build API binary
FROM builder AS build-api

RUN CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build \
    -trimpath \
    -ldflags="-s -w -buildid=" \
    -o /out/api \
    ./cmd/api

# Build inference binary
FROM builder AS build-inference

RUN CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build \
    -trimpath \
    -ldflags="-s -w -buildid=" \
    -o /out/inference \
    ./cmd/inference

# API image
FROM gcr.io/distroless/static-debian12 AS api

COPY --from=build-api /out/api /api
COPY --from=builder /src/data/artifacts/mappings /data/artifacts/mappings

ENV MAPPINGS_DIR=/data/artifacts/mappings
ENV API_PORT=8080

EXPOSE 8080

ENTRYPOINT ["/api"]

# Inference image
FROM gcr.io/distroless/static-debian12 AS inference

COPY --from=build-inference /out/inference /inference

ENV INFERENCE_PORT=9001

EXPOSE 9001

ENTRYPOINT ["/inference"]
