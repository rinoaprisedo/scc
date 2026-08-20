FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
RUN go install github.com/swaggo/swag/cmd/swag@latest
COPY . .
RUN swag init -g main.go -o docs
RUN CGO_ENABLED=0 GOOS=linux go build -o /baseadmin main.go

FROM alpine:3.24
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /baseadmin ./baseadmin
COPY --from=builder /app/storage ./storage
EXPOSE 8880
CMD ["./baseadmin"]
