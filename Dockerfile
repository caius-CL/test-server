# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制 go mod 文件和源代码
COPY go.mod main.go ./

# 生成 go.sum 并下载依赖
RUN go mod tidy && go mod download

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o test-server .

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/test-server .

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./test-server"]

