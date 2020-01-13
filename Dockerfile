FROM golang:1.13.6-alpine AS builder

ENV PROJECT github.com/CloudplexPlatform/hostname-app
WORKDIR /go/src/$PROJECT

COPY . .
RUN go build -o /hostnameservice .

FROM alpine:3.11.2 AS release
RUN apk add --no-cache ca-certificates

WORKDIR /hostnameservice
COPY --from=builder /hostnameservice ./server
EXPOSE 3550
ENTRYPOINT ["/hostnameservice/server"]

