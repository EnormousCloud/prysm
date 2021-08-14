FROM golang:1.16-alpine as builder
WORKDIR /app

ADD go.mod /app/go.mod
ADD go.sum /app/go.sum
RUN go mod download

ADD cmd /app/cmd
ADD rpc /app/rpc
ADD types /app/types
RUN go build -o ./bin/cacher ./cmd/cacher/

FROM alpine
WORKDIR /app
VOLUME /cache
COPY --from=0 /app/bin/cacher /app/cacher
ADD .env /app/.env

# default script will try only to watch the head
CMD ["sh", "-c", "while true; do /app/cacher -timeout 3 -offset 1 -limit 50; done"]
