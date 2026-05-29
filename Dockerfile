FROM golang:1.26 AS builder

WORKDIR /build

COPY go.mod go.sum ./

COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/url-shortener ./cmd/app/main.go

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/url-shortener ./url-shortener

CMD ["./url-shortener"]
