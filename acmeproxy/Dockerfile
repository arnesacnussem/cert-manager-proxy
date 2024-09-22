ARG go=1.23.1
ARG alpine=3.20
FROM golang:${go}-alpine${alpine} AS builder
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags "-s -w" -o acmeproxy .

FROM alpine:${alpine}
COPY --from=builder /build/acmeproxy .

ENV CONFIG_PATH="/config/config.yaml"
CMD ["/acmeproxy"]