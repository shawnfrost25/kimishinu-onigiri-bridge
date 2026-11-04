FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/kimishinu ./cmd/sheena

FROM scratch

# Think of this as moving the binary we made ("kimishinu") from /app (the stage before) to this one
COPY --from=builder /app/bin/kimishinu ./
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE 8080

CMD [ "./kimishinu" ]
