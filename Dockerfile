# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS build
WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/barghman .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
	&& adduser -D -H -u 1000 barghman

WORKDIR /home/barghman

# Default paths: mount your config and optionally a cache volume.
ENV HOME=/home/barghman \
	XDG_CACHE_HOME=/var/cache/barghman

RUN mkdir -p /etc/barghman /var/cache/barghman \
	&& chown -R barghman:barghman /home/barghman /var/cache/barghman

COPY --from=build /out/barghman /usr/local/bin/barghman
COPY example.toml /etc/barghman/config.toml.example

USER barghman

VOLUME ["/etc/barghman", "/var/cache/barghman"]

ENTRYPOINT ["barghman", "-file", "/etc/barghman/config.toml"]
