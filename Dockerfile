#docker build -t go-worker-credit .

FROM golang:1.22 As builder

WORKDIR /app
COPY . .

WORKDIR /app/cmd
RUN go build -o go-worker-credit -ldflags '-linkmode external -w -extldflags "-static"'

FROM alpine

WORKDIR /app
COPY --from=builder /app/cmd/go-worker-credit .

CMD ["/app/go-worker-credit"]