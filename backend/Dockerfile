ARG GO_IMAGE=golang:1.25.13-alpine@sha256:1e0126852075c9c60731c8ba49088448b91f63e2aed97ca9d1a9791622a05946

FROM ${GO_IMAGE} AS build-base

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

FROM build-base AS development

RUN apk add --no-cache ca-certificates wget \
    && go install github.com/swaggo/swag/cmd/swag@v1.16.6 \
    && go install github.com/air-verse/air@v1.64.5

RUN addgroup -S -g 10001 biblios \
    && adduser -S -D -H -u 10001 -G biblios biblios \
    && mkdir -p /app/data/covers /app/tmp \
    && chown -R biblios:biblios /app

WORKDIR /app
COPY --chown=biblios:biblios . .

ENV PATH="/go/bin:${PATH}"
USER biblios

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=15s --retries=5 \
    CMD wget -qO- http://127.0.0.1:8080/health >/dev/null || exit 1

CMD ["air"]

FROM build-base AS production-build

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/biblios-api ./cmd/api

FROM alpine:3.22.1@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1 AS production

RUN apk add --no-cache ca-certificates wget \
    && addgroup -S -g 10001 biblios \
    && adduser -S -D -H -u 10001 -G biblios biblios \
    && mkdir -p /app/data/covers \
    && chown -R biblios:biblios /app

COPY --from=production-build /out/biblios-api /usr/local/bin/biblios-api

WORKDIR /app
USER biblios

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=15s --retries=5 \
    CMD wget -qO- http://127.0.0.1:8080/health >/dev/null || exit 1

ENTRYPOINT ["/usr/local/bin/biblios-api"]
