FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY src/*.go ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-w -s" -tags lambda.norpc -o bootstrap .

FROM scratch

COPY --from=builder /build/bootstrap /bootstrap

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/bootstrap"]
