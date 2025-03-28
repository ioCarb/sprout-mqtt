# ghcr.io/machinefi/coordinator:latest
FROM golang:1.22-alpine AS builder

ENV GO111MODULE=on

WORKDIR /go/src
COPY ./ ./

RUN cd ./cmd/coordinator && CGO_ENABLED=0 go build -ldflags "-s -w -extldflags '-static'" -o coordinator

FROM --platform=linux/amd64 alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/src/cmd/coordinator/coordinator /go/bin/coordinator
EXPOSE 9001

WORKDIR /go/bin
ENTRYPOINT ["/go/bin/coordinator"]
