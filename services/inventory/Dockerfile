FROM golang:1.23.7-alpine3.21 AS builder

RUN mkdir /app
ADD . /app
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/app/main.go

FROM alpine:3.21

RUN mkdir /app

WORKDIR /app

COPY --from=builder /app/main .

CMD ["./main"]
