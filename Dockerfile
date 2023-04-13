FROM golang:1.20.3-alpine3.17 AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -ldflags "-s -w" -o ./goapi-template ./main.go

FROM alpine:3.17 AS runner
WORKDIR /app
COPY --from=builder /app/goapi-template .
COPY --from=builder /app/auth/authz.rego ./auth/authz.rego
EXPOSE 8000
ENTRYPOINT ["./goapi-template"]