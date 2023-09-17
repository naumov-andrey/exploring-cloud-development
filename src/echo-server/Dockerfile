FROM golang:1.19-alpine as builder

WORKDIR /build
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build  -o ./echo-server ./cmd/main.go

FROM alpine:latest

WORKDIR /service
COPY --from=builder /build/echo-server ./echo-server

ENTRYPOINT ["./echo-server"]
