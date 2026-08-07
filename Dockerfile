FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN apk add --no-cache make

RUN make build


FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/passwordmanager-server ./server
COPY --from=builder /app/bin/passwordmanager-migrator ./migrator

COPY migrations ./migrations

COPY server.crt .
COPY server.key .

CMD ["sh", "-c", "./migrator && ./server"]