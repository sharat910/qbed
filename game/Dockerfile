# build stage
FROM golang:1.14-alpine AS build-env
RUN apk --no-cache add build-base git gcc
ADD . /src
RUN cd /src && go build -o game

# final stage
FROM alpine:3
RUN apk --no-cache add tcpdump iputils iperf iperf3
WORKDIR /app
COPY --from=build-env /src/game /app/
CMD ["ash"]