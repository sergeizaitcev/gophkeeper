# syntax=docker/dockerfile:1

FROM golang:1.21 AS build

ENV CGO_ENABLED 0
ENV GOARCH amd64
ENV GOOS linux

WORKDIR /go/src/gophkeeper

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download && go mod verify && \
    go build -a --ldflags "-s -w" \
    -o ./bin/gophkeeperd ./cmd/gophkeeper-server/ && \
    make cert

FROM alpine:3.19

WORKDIR /opt/gophkeeper

COPY --from=build /go/src/gophkeeper/bin/gophkeeperd ./gophkeeperd
COPY --from=build /go/src/gophkeeper/server.crt ./server.crt
COPY --from=build /go/src/gophkeeper/server.key ./server.key

RUN apk update && \
	apk add --no-cache ca-certificates && \
	chown nobody: /opt/gophkeeper -R

EXPOSE 433 8433

VOLUME /opt/gophkeeper/users

USER nobody

ENTRYPOINT [ "/opt/gophkeeper/gophkeeperd" ]
