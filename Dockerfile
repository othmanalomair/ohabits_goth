# ---- Build Stage ----
FROM docker.io/library/golang:1.24.1 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

# ---- Runtime Stage ----
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/app .
# Copy the .env file and templates folder so they’re available at runtime
COPY .env .env
COPY templates/ templates/

ENV DATABASE_URL="postgres://most3mr:50998577@most3mr.com:5432/ohabits?sslmode=disable"
ENV JWT_SECRET="most3mr123"

EXPOSE 8080
CMD ["./app"]
