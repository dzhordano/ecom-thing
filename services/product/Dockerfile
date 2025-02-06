FROM golang:1.23.5-alpine3.20 AS builder

RUN mkdir /app
ADD . /app
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/app/main.go

FROM alpine:3.20

RUN mkdir /app

WORKDIR /app

COPY --from=builder /app/main .

CMD ["./main"]