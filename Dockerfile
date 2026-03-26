FROM golang:1.26 AS builder

ARG VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-X main.Version=${VERSION}" -o /go-skylight .

FROM alpine:3

COPY --from=builder /go-skylight /usr/local/bin/go-skylight
ENTRYPOINT ["go-skylight"]
