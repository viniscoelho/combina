FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY src/ src/
RUN go build -mod=vendor -o combina ./src/main.go

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/combina .
COPY web/ web/
EXPOSE 3000
CMD ["sh", "-c", "./combina -init-db; ./combina"]
