# Stage 1: Build
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o meshmeshgo .

# Stage 2: Runner
FROM alpine:3.20

WORKDIR /root/

COPY --from=builder /app/meshmeshgo .

COPY docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

EXPOSE 4040

# CMD ["./meshmeshgo"]
CMD ["./docker-entrypoint.sh"]