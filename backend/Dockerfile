# 构建阶段
FROM golang:1.26-alpine AS builder

# 设置工作目录
WORKDIR /app

# 设置 Go 代理和工具链
ENV GOPROXY=https://goproxy.cn,direct
ENV GOTOOLCHAIN=auto

# 复制 go.mod 和 go.sum 先下载依赖
COPY go.mod go.sum ./
RUN go mod tidy


# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o xhw-service .

# 运行阶段
FROM alpine:3.19

# 安装时区数据
RUN apk --no-cache add tzdata

# 设置时区为上海
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /app/xhw-service .
# 复制配置文件
COPY --from=builder /app/config ./config

# 暴露端口（根据实际情况调整）
EXPOSE 8080

# 运行应用
CMD ["./xhw-service"]