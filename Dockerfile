FROM golang:1.15.0-alpine3.12 as builder

RUN apk add --no-cache \
  build-base

WORKDIR /build
COPY go.* ./
RUN go mod download

ENV GOOS=linux
ENV CGO_ENABLED=0
ENV GOARCH=amd64

COPY . .
RUN go build -ldflags "-linkmode external -extldflags -static" -a cmd/cmd2server/main.go

FROM scratch
COPY --from=builder /build/main /cmd2server
