# syntax=docker/dockerfile:1

# Which example under examples/ to build into the image.
# Override with: docker build --build-arg EXAMPLE=01-simple-upload .
ARG EXAMPLE=11-production-ready-service
ARG GO_VERSION=1.27

# ---------- build ----------
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS build

WORKDIR /src

# Cached dependency layer: only re-runs when go.mod/go.sum change.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

ARG EXAMPLE
ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/app ./examples/${EXAMPLE}

# ---------- runtime ----------
FROM alpine:3.22

# S3 is reached over HTTPS, so the root certificates are required.
RUN apk add --no-cache ca-certificates \
    && adduser -D -u 10001 app

COPY --from=build /out/app /usr/local/bin/app

USER app
WORKDIR /home/app

ENTRYPOINT ["/usr/local/bin/app"]
