# build stage
FROM golang:1.16-alpine AS build-env
RUN apk --no-cache add build-base git gcc
WORKDIR /src

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o game

# final stage
FROM alpine:3
RUN apk --no-cache add tcpdump iputils iperf iperf3 ethtool
WORKDIR /app
COPY --from=build-env /src/game /app/
CMD ["ash"]