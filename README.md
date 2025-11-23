# Test Server

一个简单的 Go 服务器，用于暴露系统和网络指标给 Prometheus 监控。

## 功能特性

- **系统指标**：
  - CPU 使用率（每个核心）
  - 内存使用情况（已用/总计）
  - Go runtime 指标（goroutine 数量）

- **网络指标**：
  - 网络接口发送/接收字节数
  - 网络接口发送/接收包数
  - 网络错误统计（输入/输出）

- **HTTP 指标**：
  - HTTP 请求总数
  - HTTP 请求延迟

## 端点

- `GET /` - 根路径，显示服务器信息
- `GET /health` - 健康检查端点
- `GET /metrics` - Prometheus metrics 端点

## 构建和运行

### 本地运行

```bash
cd test-server
go mod download
go run main.go
```

服务器将在 `http://localhost:8080` 启动。

### 构建 Docker 镜像

```bash
docker build -t test-server:latest .
```

### 部署到 Kubernetes

### 方式 1: 使用 Cloud Build CI/CD (推荐)

1. 确保已创建 Artifact Registry repository：
```bash
gcloud artifacts repositories create caius-test-repo \
  --repository-format=docker \
  --location=asia-east1 \
  --description="Test repository for test-server"
```

2. 使用 Cloud Build 构建和部署：
```bash
cd test-project/test-server
gcloud builds submit --config=cicd/cloudbuild.yaml .
```

3. 查看构建状态：
```bash
gcloud builds list --limit=5
```

详细说明请参考 [cicd/README.md](cicd/README.md)

### 方式 2: 手动部署

1. 构建并推送镜像到你的容器仓库（或使用本地镜像）：
```bash
docker build -t your-registry/test-server:latest .
docker push your-registry/test-server:latest
```

2. 更新 `k8s/deployment.yaml` 中的镜像名称

3. 部署：
```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

## Prometheus 监控

服务器已经配置了 Prometheus annotations，Prometheus 会自动发现并抓取 metrics：

- `prometheus.io/scrape: "true"` - 启用抓取
- `prometheus.io/port: "8080"` - metrics 端口
- `prometheus.io/path: "/metrics"` - metrics 路径

## 可用的 Metrics

### 系统指标
- `test_server_cpu_usage_percent` - CPU 使用率百分比
- `test_server_memory_usage_bytes` - 内存使用量（字节）
- `test_server_memory_total_bytes` - 总内存（字节）
- `test_server_goroutines` - Goroutine 数量

### 网络指标
- `test_server_network_bytes_sent{interface="..."}` - 发送字节数
- `test_server_network_bytes_recv{interface="..."}` - 接收字节数
- `test_server_network_packets_sent{interface="..."}` - 发送包数
- `test_server_network_packets_recv{interface="..."}` - 接收包数
- `test_server_network_errors_in{interface="..."}` - 输入错误数
- `test_server_network_errors_out{interface="..."}` - 输出错误数

### HTTP 指标
- `test_server_http_requests_total{method, endpoint, status}` - HTTP 请求总数
- `test_server_http_request_duration_seconds{method, endpoint}` - HTTP 请求延迟

