FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /vatsapartment .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata curl
RUN adduser -D -g '' appuser
WORKDIR /app
COPY --from=builder /vatsapartment .
COPY templates/ ./templates/
COPY static/ ./static/
RUN chown -R appuser:appuser /app
USER appuser
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -sf http://localhost:8080/health || exit 1
CMD ["./vatsapartment"]
