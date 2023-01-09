FROM golang:1.19-alpine as builder

WORKDIR /root
RUN apk update --no-cache && \
    apk upgrade --no-cache && \
    apk add --no-cache \
    make \
    build-base

COPY go.mod .
RUN go mod download

COPY . .
RUN go build -o zru main.go

FROM alpine:3.17.0

WORKDIR /app

COPY --from=builder /root/zru .

ENTRYPOINT [ "/app/zru" ]
