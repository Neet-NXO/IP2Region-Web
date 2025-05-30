# IP2Region Web

<div align="center">

![IP2Region Web](https://img.shields.io/badge/IP2Region-Web-blue)
![Go Version](https://img.shields.io/badge/Go-1.24+-green)
![Vue Version](https://img.shields.io/badge/Vue-3.0+-brightgreen)
![License](https://img.shields.io/badge/License-Apache%202.0-blue)
![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen)

一个现代化的 IP 地理位置查询 Web 应用，基于 ip2region 数据库构建

</div>

## 🙏 致谢

> **本项目基于 [ip2region](https://github.com/lionsoul2014/ip2region) 构建**
> 
> 感谢 [@lionsoul2014](https://github.com/lionsoul2014) 及其团队开发的优秀 ip2region 离线IP地址定位库。
> 
> ip2region 是一个离线IP地址定位库和IP定位数据管理框架，支持亿级别数据段，十微秒级查询性能。
> 
> - 原始项目: https://github.com/lionsoul2014/ip2region
> - 许可证: Apache License 2.0

## ✨ 特性

- 🚀 **快速查询**: 毫秒级 IP 地理位置查询
- 🌐 **现代界面**: 基于 Vue 3 + TypeScript 的响应式前端
- 📊 **数据可视化**: 查询统计图表和历史记录
- 🔍 **批量查询**: 支持批量 IP 地址查询
- 💾 **多数据库**: 支持 MySQL 和 SQLite
- 🔐 **用户认证**: JWT 令牌认证系统
- 📋 **API 文档**: 完整的 Swagger API 文档
- 🎯 **高性能**: Go 语言构建的高性能后端
- 📱 **移动友好**: 响应式设计，支持移动设备

## 🛠 技术栈

### 后端
- **语言**: Go 1.24+
- **框架**: Gin (HTTP 路由)
- **ORM**: GORM (数据库操作)
- **认证**: JWT (JSON Web Token)
- **文档**: Swagger/OpenAPI
- **数据库**: MySQL / SQLite

### 前端
- **框架**: Vue 3 + TypeScript
- **构建工具**: Vite
- **UI 库**: Element Plus
- **状态管理**: Pinia
- **路由**: Vue Router 4
- **HTTP 客户端**: Axios
- **图表**: ECharts

## 📦 快速开始

### 环境要求

- Go 1.24+
- Node.js 18+
- MySQL 8.0+ (可选，也可使用 SQLite)

### 安装与运行

#### 1. 克隆项目
```bash
git clone https://github.com/Neet-NXO/IP2Region-Web.git
cd IP2Region-Web
```

#### 2. 后端启动
```bash
# 下载依赖
go mod download

# 复制配置文件
cp config.example.yaml config.yaml

# 编辑配置文件
vim config.yaml

# 运行项目
go run main.go
```

#### 3. 前端启动
```bash
# 进入前端目录
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

#### 4. 使用构建脚本

**Linux/macOS:**
```bash
# 检查环境
make check

# 完整构建
make build

# 运行项目
make run
```

**Windows:**
```batch
# 检查环境
make.bat check

# 完整构建
make.bat build

# 运行项目
make.bat run
```

### 🐳 Docker 部署

```bash
# 构建镜像
docker build -t ip2region-web .

# 运行容器
docker run -d -p 8080:8080 ip2region-web
```

### Docker Compose
```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

## ⚙️ 配置说明

### 配置文件 (config.yaml)

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"  # debug, release, test

database:
  type: "sqlite"   # mysql, sqlite
  # MySQL 配置
  mysql:
    host: "localhost"
    port: 3306
    username: "root"
    password: "password"
    database: "ip2region"
  # SQLite 配置
  sqlite:
    path: "./data/ip2region.db"

cache:
  enabled: true
  ttl: 3600        # 缓存时间(秒)

security:
  jwt_secret: "your-secret-key"
  jwt_expire: 86400  # JWT 过期时间(秒)

ip2region:
  xdb_path: "./data/ip2region.xdb"
  update_interval: 86400  # 数据更新检查间隔(秒)
```

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `CONFIG_PATH` | 配置文件路径 | `./config.yaml` |
| `SERVER_PORT` | 服务端口 | `8080` |
| `DB_TYPE` | 数据库类型 | `sqlite` |
| `JWT_SECRET` | JWT 密钥 | `your-secret-key` |

## 📖 API 文档

启动服务后，访问 Swagger 文档：
- 开发环境: http://localhost:8080/swagger/index.html
- 生产环境: http://your-domain/swagger/index.html

### 主要 API 接口

#### IP 查询
```bash
# 单个 IP 查询
GET /api/v1/query?ip=8.8.8.8

# 批量 IP 查询
POST /api/v1/query/batch
{
  "ips": ["8.8.8.8", "114.114.114.114"]
}
```

#### 用户认证
```bash
# 用户登录
POST /api/v1/auth/login
{
  "username": "admin",
  "password": "password"
}

# 获取用户信息
GET /api/v1/auth/user
Authorization: Bearer <token>
```

#### 查询历史
```bash
# 获取查询历史
GET /api/v1/history?page=1&limit=20

# 查询统计
GET /api/v1/stats
```

## 🚀 部署指南

### 系统服务 (Linux)

1. 创建服务文件：
```bash
sudo vim /etc/systemd/system/ip2region-web.service
```

2. 添加服务配置：
```ini
[Unit]
Description=IP2Region Web Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/ip2region-web
ExecStart=/opt/ip2region-web/ip2region-web
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

3. 启动服务：
```bash
sudo systemctl daemon-reload
sudo systemctl enable ip2region-web
sudo systemctl start ip2region-web
```

### Nginx 反向代理

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 静态文件缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

## 🔧 开发指南

### 项目结构
```
IP2Region-Web/
├── api/                    # API 接口层
│   ├── controller/         # 控制器
│   ├── middleware/         # 中间件
│   └── router/            # 路由配置
├── internal/              # 内部包
│   ├── config/            # 配置管理
│   ├── database/          # 数据库
│   ├── model/             # 数据模型
│   ├── service/           # 业务逻辑
│   └── utils/             # 工具函数
├── frontend/              # 前端代码
│   ├── src/
│   │   ├── api/           # API 接口
│   │   ├── components/    # 组件
│   │   ├── views/         # 页面
│   │   ├── stores/        # 状态管理
│   │   └── utils/         # 工具函数
│   ├── public/            # 静态资源
│   └── package.json
├── data/                  # 数据文件
├── docs/                  # 文档
├── scripts/               # 脚本文件
├── go.mod
├── main.go
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

### 贡献指南

我们欢迎所有形式的贡献！请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详细信息。

#### 开发流程
1. Fork 项目
2. 创建特性分支
3. 提交代码
4. 创建 Pull Request

#### 代码规范
- 后端遵循 Go 标准代码风格
- 前端使用 ESLint + Prettier
- 提交信息遵循 Conventional Commits

## 📊 性能指标

- **查询响应时间**: < 1ms (本地查询)
- **并发处理能力**: 10,000+ QPS
- **内存占用**: < 100MB (基础运行)
- **数据库大小**: ~11MB (ip2region.xdb)

## 🤝 贡献者

感谢所有为项目做出贡献的开发者！

<a href="https://github.com/Neet-NXO/IP2Region-Web/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=Neet-NXO/IP2Region-Web" />
</a>

## 📄 许可证

本项目基于 [Apache License 2.0](LICENSE) 许可证开源。

## 🔗 相关链接

- **[ip2region 原始项目](https://github.com/lionsoul2014/ip2region)** - 离线IP地址定位库和数据管理框架
- [Go 官方文档](https://golang.org/doc/)
- [Vue 3 文档](https://v3.vuejs.org/)
- [Element Plus](https://element-plus.org/)

## 📮 联系我们

- 提交 Issue: [GitHub Issues](https://github.com/Neet-NXO/IP2Region-Web/issues)
- GitHub: [@Neet-NXO](https://github.com/Neet-NXO)

---

<div align="center">
  Made with ❤️ by IP2Region Web Contributors
</div>
