# Cloud Build CI/CD 配置

这个目录包含用于自动构建和部署 test-server 的 Cloud Build 配置。

## 配置说明

### cloudbuild.yaml

Cloud Build 配置文件，包含以下步骤：

1. **构建 Docker 镜像** - 从 Dockerfile 构建镜像
2. **推送镜像到 Artifact Registry** - 推送镜像到 GCP Artifact Registry
3. **获取 GKE 集群凭证** - 连接到目标 Kubernetes 集群
4. **部署 Kubernetes 资源** - 应用 namespace、deployment 和 service
5. **等待部署完成** - 等待 rollout 完成
6. **验证部署** - 验证所有 pod 正常运行

## 使用方法

### 前置要求

1. 确保已创建 Artifact Registry repository：
```bash
gcloud artifacts repositories create caius-test-repo \
  --repository-format=docker \
  --location=asia-east1 \
  --description="Test repository for test-server"
```

2. 确保 Cloud Build Service Account 有必要的权限：
   - Artifact Registry Writer
   - Kubernetes Engine Developer
   - Service Account User

### 手动触发构建

```bash
cd test-project/test-server
gcloud builds submit --config=cicd/cloudbuild.yaml .
```

### 使用自定义参数

```bash
gcloud builds submit --config=cicd/cloudbuild.yaml . \
  --substitutions=_REPLICAS=3,_REGION=asia-east1
```

### 设置 Cloud Build 触发器

1. 在 GCP Console 中进入 Cloud Build > Triggers
2. 创建新的触发器
3. 连接到你的 Git 仓库（GitHub、GitLab、Cloud Source Repositories）
4. 配置触发器：
   - **名称**: test-server-deploy
   - **事件**: Push to a branch
   - **分支**: `^main$` 或 `^master$`
   - **配置文件路径**: `test-project/test-server/cicd/cloudbuild.yaml`
   - **替换变量**: 可以覆盖默认的 substitutions

## 配置变量

可以在 `cloudbuild.yaml` 的 `substitutions` 部分修改以下变量：

- `_REGION`: Artifact Registry 区域
- `_REPO_NAME`: Artifact Registry 仓库名称
- `_SERVICE_NAME`: 服务名称（用于镜像命名）
- `_CLUSTER_NAME`: GKE 集群名称
- `_ZONE_NAME`: GKE 集群区域
- `_NAMESPACE_NAME`: Kubernetes namespace
- `_DEPLOYMENT_NAME`: Deployment 名称
- `_APP_LABEL`: 应用标签
- `_CONTAINER_PORT`: 容器端口
- `_SERVICE_PORT`: Service 端口
- `_REPLICAS`: Pod 副本数
- `_MAX_SURGE`: 滚动更新最大 surge
- `_MAX_UNAVAILABLE`: 滚动更新最大不可用数
- `_INITIAL_DELAY_SECONDS`: 健康检查初始延迟
- `_PERIOD_SECONDS`: 健康检查周期

## 镜像标签

- `${SHORT_SHA}`: Git commit SHA 的前 7 位字符（自动生成）
- `latest`: 最新构建的镜像

## 故障排查

### 构建失败

1. 检查 Cloud Build 日志：
```bash
gcloud builds list --limit=5
gcloud builds log <BUILD_ID>
```

2. 检查权限：
```bash
gcloud projects get-iam-policy <PROJECT_ID> \
  --flatten="bindings[].members" \
  --filter="bindings.members:serviceAccount:PROJECT_NUMBER@cloudbuild.gserviceaccount.com"
```

### 部署失败

1. 检查 Kubernetes 资源：
```bash
kubectl get pods -n test-server
kubectl describe pod <POD_NAME> -n test-server
kubectl logs <POD_NAME> -n test-server
```

2. 检查镜像是否存在：
```bash
gcloud artifacts docker images list \
  asia-east1-docker.pkg.dev/<PROJECT_ID>/caius-test-repo/test-server
```

## 监控

部署完成后，Prometheus 会自动发现并开始抓取 metrics。可以在 Prometheus UI 中查询：

- `test_server_cpu_usage_percent`
- `test_server_network_bytes_sent`
- `test_server_http_requests_total`

