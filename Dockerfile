FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /vatsapartment .

# Build Tailwind CSS
FROM alpine:3.21 AS tailwind
RUN apk add --no-cache curl
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v4.1.17/tailwindcss-linux-x64 \
 && chmod +x tailwindcss-linux-x64
COPY templates/ ./templates/
# Create input CSS with Tailwind directives and full theme matching site design
RUN echo '@import "tailwindcss";' > input.css \
 && echo '@theme { --color-primary: #1a1a2e; --color-background: #f8fafc; --color-on-surface: #0f172a; --color-on-sf-variant: #64748b; --color-outline: #94a3b8; --color-secondary: #f59e0b; --color-on-secondary: #1e293b; --color-sec-container: #fef3c7; --color-on-sec-container: #92400e; --color-sf-lowest: #ffffff; --color-sf-low: #f1f5f9; --color-sf-high: #e2e8f0; --color-sf-highest: #f8fafc; --color-sf-variant: #cbd5e1; --color-ol-variant: #e2e8f0; }' >> input.css
# Scan all template files and generate single minified CSS
RUN ./tailwindcss-linux-x64 -i input.css -o /app-tailwind.css --minify --content "templates/**/*.html"

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata curl
RUN adduser -D -g '' appuser
WORKDIR /app
COPY --from=builder /vatsapartment .
COPY templates/ ./templates/
COPY static/ ./static/
COPY --from=tailwind /app-tailwind.css static/app-tailwind.css
RUN chown -R appuser:appuser /app
USER appuser
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -sf http://localhost:8080/health || exit 1
CMD ["./vatsapartment"]
