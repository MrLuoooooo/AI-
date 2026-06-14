# ---- 构建阶段 ----
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build
ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/vision-assistant ./cmd/server

# ---- 运行阶段 ----
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata curl

WORKDIR /app
COPY --from=builder /build/vision-assistant ./vision-assistant
COPY --from=builder /build/web/index.html ./web/index.html
COPY --from=builder /build/configs/config.yaml /app/configs/config.yaml

RUN mkdir -p /app/logs

EXPOSE 8080

ENV CONFIG_PATH=/app/configs
ENV TZ=Asia/Shanghai

ENTRYPOINT ["./vision-assistant"]
