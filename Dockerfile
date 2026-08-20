FROM golang:1.26.3 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kota-siaga .

FROM alpine:3.22

WORKDIR /root/
RUN apk add --no-cache tzdata
COPY --from=builder /app/kota-siaga .
COPY entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

EXPOSE 8080
ENTRYPOINT ["./entrypoint.sh"]
