FROM golang:1.21 AS build

RUN mkdir -p /src

WORKDIR /src

COPY go.mod go.sum Makefile /src/
RUN make mod

COPY . /src
RUN make build-linux

FROM debian:11.4-slim

RUN apt update && apt install -y ca-certificates tzdata

COPY --from=build /src/matrix-on-call-bot /usr/local/bin/

CMD ["/usr/local/bin/matrix-on-call-bot"]
