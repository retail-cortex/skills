# Multi-Language Dockerfile Standards

## 1. Go Statically Linked Container (scratch / distroless)

```dockerfile
# Build Stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /service cmd/server/main.go

# Final Runtime
FROM gcr.io/distroless/static-debian12
COPY --from=builder /service /service
EXPOSE 8080 9090
USER nonroot:nonroot
ENTRYPOINT ["/service"]
```

## 2. Python FastAPI / ADK Container (uv slim)

```dockerfile
FROM python:3.13-slim AS builder
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/
WORKDIR /app
COPY pyproject.toml uv.lock ./
RUN uv sync --frozen --no-cache

FROM python:3.13-slim
WORKDIR /app
COPY --from=builder /app/.venv /app/.venv
COPY . .
ENV PATH="/app/.venv/bin:$PATH"
USER nobody
EXPOSE 8000
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```
