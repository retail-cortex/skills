FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY examples/service.go .
RUN go build -o /app/service service.go

FROM scratch
COPY --from=builder /app/service /service
ENTRYPOINT ["/service"]
