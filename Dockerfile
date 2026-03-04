FROM golang:1.26-alpine AS builder

WORKDIR /build-dir

COPY go.mod go.sum ./

RUN go mod download

COPY .. /build-dir

RUN CGO_ENABLE=0 go build -ldflags="-w -s" -o ./peer ./main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /build-dir/peer ./start

# --env-file .env

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

CMD [ "/app/start" ]
