# syntax=docker/dockerfile:1.7

FROM node:22-alpine AS web-deps
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm,sharing=locked \
    npm ci

FROM web-deps AS web-build
COPY web/ ./
RUN npm run build

FROM golang:1.25-alpine AS go-deps
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    go mod download

FROM go-deps AS go-build
ARG TARGETOS
ARG TARGETARCH
COPY api ./api
COPY internal ./internal
COPY manifest/migration ./manifest/migration
COPY utility ./utility
COPY main.go ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-$(go env GOARCH)} \
    go build -trimpath -ldflags="-s -w" -o /out/repomind .

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app app
WORKDIR /app
COPY --from=go-build --chown=app:app /out/repomind ./repomind
COPY --from=web-build --chown=app:app /src/web/dist ./resource/public
COPY --chown=app:app manifest/docker/config.yaml ./config/config.yaml
ENV GF_GCFG_PATH=/app/config \
    GF_GCFG_FILE=config.yaml
EXPOSE 8000
USER app
HEALTHCHECK --interval=10s --timeout=5s --retries=5 \
    CMD wget -qO- http://127.0.0.1:8000/readyz >/dev/null || exit 1
CMD ["./repomind"]
