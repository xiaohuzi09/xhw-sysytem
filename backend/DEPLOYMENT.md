# xhw-service 部署指南

## 项目概述

xhw-service 是一个基于 Go + Gin 框架开发的 RESTful API 服务，使用 MySQL 作为数据库，集成了对象存储（RustFS/S3）和火山引擎 ARK API（视觉识别）。

## 技术栈

- **语言**: Go 1.26.1
- **Web 框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL
- **对象存储**: AWS S3 SDK (兼容 RustFS)
- **API 文档**: Swagger
- **部署方式**: Docker

## 环境要求

### 开发环境

- Go 1.26.1+
- MySQL 5.7+
- Git

### 生产环境

- Docker 20.10+
- Docker Compose (可选)
- MySQL 数据库服务

## 项目结构

```
xhw-service/
├── config/           # 配置文件
│   ├── config.go     # 配置加载逻辑
│   ├── config.yaml   # 开发环境配置
│   └── config.prod.yaml  # 生产环境配置
├── controllers/      # 控制器层
├── docs/             # Swagger 文档
├── middleware/       # 中间件
├── models/           # 数据模型
├── routes/           # 路由配置
├── services/         # 业务逻辑层
├── utils/            # 工具函数
├── Dockerfile        # Docker 构建文件
├── go.mod            # Go 依赖管理
└── main.go           # 应用入口
```

## 本地开发部署

### 1. 克隆代码

```bash
git clone <repository-url>
cd xhw-service
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置环境

编辑 `config/config.yaml` 文件，根据本地环境修改以下配置：

```yaml
server:
  port: 8080
  mode: debug

database:
  host: localhost
  port: 3306
  user: root
  password: "your_password"
  dbname: xhw_service

rustfs:
  endpoint: "https://your-rustfs-endpoint.com"
  access_key_id: "your_access_key"
  secret_access_key: "your_secret_key"
  region: "us-east-1"
  use_path_style: true

ark:
  api_key: "your_ark_api_key"
  model: "doubao-seed-2-0-lite-260215"

jwt:
  secret: "your-jwt-secret-key"
  expire_time: 86400

auth:
  admin_usernames:
    - admin
```

### 4. 创建数据库

```bash
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS xhw_service CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

### 5. 运行应用

```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动。

### 6. 查看 API 文档

访问 Swagger UI: `http://localhost:8080/swagger/index.html`

## Docker 部署

### 构建镜像

```bash
docker build -t xhw-service:latest .
```

### 运行容器

```bash
docker run -d \
  --name xhw-service \
  -p 8080:8080 \
  -e ENV=prod \
  -v $(pwd)/config:/app/config \
  xhw-service:latest
```

### 使用 Docker Compose (推荐)

创建 `docker-compose.yml`:

```yaml
version: '3.8'

services:
  app:
    build: .
    container_name: xhw-service
    ports:
      - "8080:8080"
    environment:
      - ENV=prod
    volumes:
      - ./config:/app/config
    depends_on:
      - mysql
    restart: unless-stopped

  mysql:
    image: mysql:8.0
    container_name: xhw-mysql
    environment:
      MYSQL_ROOT_PASSWORD: your_password
      MYSQL_DATABASE: xhw_service
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    restart: unless-stopped

volumes:
  mysql_data:
```

启动服务：

```bash
docker-compose up -d
```

## 生产环境部署

### 1. 准备生产配置

确保 `config/config.prod.yaml` 已正确配置：

- 使用生产环境数据库
- 修改 JWT Secret 为强密码
- 配置正确的对象存储凭证
- 设置管理员白名单

### 2. 设置环境变量

```bash
export ENV=prod
```

### 3. 构建并部署

```bash
# 构建镜像
docker build -t xhw-service:prod .

# 运行容器
docker run -d \
  --name xhw-service-prod \
  -p 8080:8080 \
  -e ENV=prod \
  --restart unless-stopped \
  xhw-service:prod
```

## 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `ENV` | 运行环境 (`dev`/`prod`) | dev |
| `GIN_MODE` | Gin 框架模式 | 根据配置 |

### 配置文件优先级

1. 根据 `ENV` 环境变量选择配置文件
2. `ENV=prod` 或 `ENV=production` → `config.prod.yaml`
3. 其他情况 → `config.yaml`

## 健康检查

服务启动后，可以通过以下方式检查健康状态：

```bash
curl http://localhost:8080/health
```

## 日志查看

### Docker 部署

```bash
docker logs -f xhw-service
```

### 本地运行

日志直接输出到控制台。

## 常见问题

### 1. 数据库连接失败

- 检查 MySQL 服务是否运行
- 验证数据库配置（host、port、user、password）
- 确认数据库已创建

### 2. 端口被占用

修改 `config.yaml` 中的 `server.port` 为其他端口。

### 3. 配置文件加载失败

- 确认配置文件路径正确
- 检查 YAML 格式是否合法
- 查看应用启动日志获取详细信息

### 4. Docker 构建失败

- 检查 Dockerfile 语法
- 确认 Go 代理可访问（已配置 GOPROXY=https://goproxy.cn）
- 查看构建日志排查问题

## 更新部署

### 代码更新

```bash
# 拉取最新代码
git pull origin main

# 重新构建并部署
docker-compose down
docker-compose up -d --build
```

### 数据库迁移

应用启动时会自动执行数据库迁移，无需手动操作。

## 安全建议

1. **生产环境务必修改 JWT Secret**
2. **使用强密码保护数据库**
3. **定期更换对象存储凭证**
4. **限制管理员账号数量**
5. **使用 HTTPS 部署**
6. **配置防火墙规则，仅开放必要端口**

## 联系与支持

如有部署问题，请联系开发团队或提交 Issue。
