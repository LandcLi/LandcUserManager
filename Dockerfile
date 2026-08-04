# ==================== Build Stage ====================
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /build

# 缓存依赖层（landc-go 依赖线上 v0.5.0，无需本地模块）
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/bin/server .

# ==================== Runtime Stage ====================
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/bin/server /app/server
COPY config.example.yml /app/config.yaml

EXPOSE 8080

ENTRYPOINT ["/app/server"]
